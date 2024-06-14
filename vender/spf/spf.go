// Package spf implements SPF (Sender Policy Framework) lookup and validation.
//
// Sender Policy Framework (SPF) is a simple email-validation system designed
// to detect email spoofing by providing a mechanism to allow receiving mail
// exchangers to check that incoming mail from a domain comes from a host
// authorized by that domain's administrators [Wikipedia].
//
// This package is intended to be used by SMTP servers to implement SPF
// validation.
//
// All mechanisms and modifiers are supported:
//
//	all
//	include
//	a
//	mx
//	ptr
//	ip4
//	ip6
//	exists
//	redirect
//	exp (ignored)
//	Macros
//
// References:
//
//	https://tools.ietf.org/html/rfc7208
//	https://en.wikipedia.org/wiki/Sender_Policy_Framework
package spf // import "blitiri.com.ar/go/spf"

import (
	"easymail/internal/easydns"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// The Result of an SPF check. Note the values have meaning, we use them in
// headers.  https://tools.ietf.org/html/rfc7208#section-8
type Result string

// Valid results.
var (
	// https://tools.ietf.org/html/rfc7208#section-8.1
	// Not able to reach any conclusion.
	None = Result("none")

	// https://tools.ietf.org/html/rfc7208#section-8.2
	// No definite assertion (positive or negative).
	Neutral = Result("neutral")

	// https://tools.ietf.org/html/rfc7208#section-8.3
	// Client is authorized to inject mail.
	Pass = Result("pass")

	// https://tools.ietf.org/html/rfc7208#section-8.4
	// Client is *not* authorized to use the domain.
	Fail = Result("fail")

	// https://tools.ietf.org/html/rfc7208#section-8.5
	// Not authorized, but unwilling to make a strong policy statement.
	SoftFail = Result("softfail")

	// https://tools.ietf.org/html/rfc7208#section-8.6
	// Transient error while performing the check.
	TempError = Result("temperror")

	// https://tools.ietf.org/html/rfc7208#section-8.7
	// Records could not be correctly interpreted.
	PermError = Result("permerror")
)

var qualToResult = map[byte]Result{
	'+': Pass,
	'-': Fail,
	'~': SoftFail,
	'?': Neutral,
}

// Errors returned by the library. Note that the errors returned in different
// situations may change over time, and new ones may be added. Be careful
// about over-relying on these.
var (
	// Errors related to an invalid SPF record.
	ErrUnknownField  = errors.New("unknown field")
	ErrInvalidIP     = errors.New("invalid ipX value")
	ErrInvalidMask   = errors.New("invalid mask")
	ErrInvalidMacro  = errors.New("invalid macro")
	ErrInvalidDomain = errors.New("invalid domain")

	// Errors related to DNS lookups.
	// Note that the library functions may also return net.DNSError.
	ErrNoResult               = errors.New("no DNS record found")
	ErrLookupLimitReached     = errors.New("lookup limit reached")
	ErrVoidLookupLimitReached = errors.New("void lookup limit reached")
	ErrTooManyMXRecords       = errors.New("too many MX records")
	ErrMultipleRecords        = errors.New("multiple matching DNS records")

	// Errors returned on a successful match.
	ErrMatchedAll    = errors.New("matched all")
	ErrMatchedA      = errors.New("matched a")
	ErrMatchedIP     = errors.New("matched ip")
	ErrMatchedMX     = errors.New("matched mx")
	ErrMatchedPTR    = errors.New("matched ptr")
	ErrMatchedExists = errors.New("matched exists")
)

const (
	// Default value for the maximum number of DNS lookups while resolving SPF.
	// RFC is quite clear 10 must be the maximum allowed.
	// https://tools.ietf.org/html/rfc7208#section-4.6.4
	defaultMaxLookups = 10

	// Default value for the maximum number of DNS void lookups while
	// resolving SPF.  RFC suggests that implementations SHOULD limit these
	// with a configurable default of 2.
	// https://tools.ietf.org/html/rfc7208#section-4.6.4
	defaultMaxVoidLookups = 2
)

// CheckHostWithSender fetches SPF records for `sender`'s domain, parses them,
// and evaluates them to determine if `ip` is permitted to send mail for it.
// The `helo` domain is used if the sender has no domain part.
//
// The `opts` optional parameter can be used to adjust some specific
// behaviours, such as the maximum number of DNS lookups allowed.
//
// The function returns a Result, which corresponds with the SPF result for
// the check as per RFC, as well as an error for debugging purposes. Note that
// the error may be non-nil even on successful checks.
//
// Reference: https://tools.ietf.org/html/rfc7208#section-4
func CheckHostWithSender(resolver *easydns.Resolver, ip net.IP, helo, sender string) (Result, error) {
	_, domain := split(sender)
	if domain == "" {
		domain = helo
	}

	r := &resolution{
		ip:           ip,
		maxCount:     defaultMaxLookups,
		maxVoidCount: defaultMaxVoidLookups,
		helo:         helo,
		sender:       sender,
		resolver:     resolver,
	}
	defer func() {
		r.resolver = nil
	}()

	return r.Check(domain)
}

// split an user@domain address into user and domain.
func split(addr string) (string, string) {
	ps := strings.SplitN(addr, "@", 2)
	if len(ps) != 2 {
		return addr, ""
	}

	return ps[0], ps[1]
}

type resolution struct {
	ip           net.IP
	count        uint
	maxCount     uint
	voidCount    uint
	maxVoidCount uint

	helo   string
	sender string

	// Result of doing a reverse lookup for ip (so we only do it once).
	ipNames []string

	// DNS resolver to use.
	resolver *easydns.Resolver
}

var aField = regexp.MustCompile(`^(a$|a:|a/)`)
var mxField = regexp.MustCompile(`^(mx$|mx:|mx/)`)
var ptrField = regexp.MustCompile(`^(ptr$|ptr:)`)

func (r *resolution) Check(domain string) (Result, error) {
	txt, err := r.getDNSRecord(domain)
	if err != nil {
		if isNotFound(err) {
			// NXDOMAIN -> None.
			// https://datatracker.ietf.org/doc/html/rfc7208#section-4.3
			return None, ErrNoResult
		}
		if isTemporary(err) {
			return TempError, err
		}
		if err == ErrMultipleRecords {
			return PermError, err
		}
		// Got another, permanent error.
		// https://datatracker.ietf.org/doc/html/rfc7208#section-2.6.7
		return PermError, err
	}

	if txt == "" {
		// No record => None.
		// https://tools.ietf.org/html/rfc7208#section-4.5
		return None, ErrNoResult
	}

	fields := strings.Split(txt, " ")

	// Redirects must be handled after the rest; instead of having two loops,
	// we just move them to the end.
	var newFields, redirects []string
	for _, field := range fields {
		if strings.HasPrefix(field, "redirect=") {
			redirects = append(redirects, field)
		} else {
			newFields = append(newFields, field)
		}
	}
	if len(redirects) > 1 {
		// At most a single redirect is allowed.
		// https://tools.ietf.org/html/rfc7208#section-6
		return PermError, ErrInvalidDomain
	}
	fields = append(newFields, redirects...)

	for _, field := range fields {
		if field == "" {
			continue
		}

		// The version check should be case-insensitive (it's a
		// case-insensitive constant in the standard).
		// https://tools.ietf.org/html/rfc7208#section-12
		if strings.HasPrefix(field, "v=") || strings.HasPrefix(field, "V=") {
			continue
		}

		// Limit the number of resolutions.
		// https://tools.ietf.org/html/rfc7208#section-4.6.4
		if r.count > r.maxCount {
			return PermError, ErrLookupLimitReached
		}

		if r.voidCount > r.maxVoidCount {
			return PermError, ErrVoidLookupLimitReached
		}

		// See if we have a qualifier, defaulting to + (pass).
		// https://tools.ietf.org/html/rfc7208#section-4.6.2
		result, ok := qualToResult[field[0]]
		if ok {
			field = field[1:]
		} else {
			result = Pass
		}

		// Mechanism and modifier names are case-insensitive.
		// https://tools.ietf.org/html/rfc7208#section-4.6.1
		lField := strings.ToLower(field)

		if lField == "all" {
			// https://tools.ietf.org/html/rfc7208#section-5.1
			return result, ErrMatchedAll
		} else if strings.HasPrefix(lField, "include:") {
			if ok, res, err := r.includeField(result, field, domain); ok {
				return res, err
			}
		} else if aField.MatchString(lField) {
			if ok, res, err := r.aField(result, field, domain); ok {
				return res, err
			}
		} else if mxField.MatchString(lField) {
			if ok, res, err := r.mxField(result, field, domain); ok {
				return res, err
			}
		} else if strings.HasPrefix(lField, "ip4:") || strings.HasPrefix(lField, "ip6:") {
			if ok, res, err := r.ipField(result, field); ok {
				return res, err
			}
		} else if ptrField.MatchString(lField) {
			if ok, res, err := r.ptrField(result, field, domain); ok {
				return res, err
			}
		} else if strings.HasPrefix(lField, "exists:") {
			if ok, res, err := r.existsField(result, field, domain); ok {
				return res, err
			}
		} else if strings.HasPrefix(lField, "exp=") {
			continue
		} else if strings.HasPrefix(lField, "redirect=") {
			res, err := r.redirectField(field, domain)
			return res, err
		} else {
			return PermError, ErrUnknownField
		}
	}

	// Got to the end of the evaluation without a result => Neutral.
	// https://tools.ietf.org/html/rfc7208#section-4.7
	return Neutral, nil
}

// getDNSRecord gets TXT records from the given domain, and returns the SPF
// (if any).  Note that at most one SPF is allowed per a given domain:
// https://tools.ietf.org/html/rfc7208#section-3
// https://tools.ietf.org/html/rfc7208#section-3.2
// https://tools.ietf.org/html/rfc7208#section-4.5
func (r *resolution) getDNSRecord(domain string) (string, error) {
	txtList, err := r.resolver.LookupTXT(domain)
	if err != nil {
		return "", err
	}

	records := []string{}
	for _, txt := range txtList {
		// The version check should be case-insensitive (it's a
		// case-insensitive constant in the standard).
		// https://tools.ietf.org/html/rfc7208#section-12
		if strings.HasPrefix(strings.ToLower(txt), "v=spf1 ") {
			records = append(records, txt)
		}

		// An empty record is explicitly allowed:
		// https://tools.ietf.org/html/rfc7208#section-4.5
		if strings.ToLower(txt) == "v=spf1" {
			records = append(records, txt)
		}
	}

	// 0 records is ok, handled by the parent.
	// 1 record is what we expect, return the record.
	// More than that, it's a permanent error:
	// https://tools.ietf.org/html/rfc7208#section-4.5
	l := len(records)
	if l == 0 {
		return "", nil
	} else if l == 1 {
		return records[0], nil
	}
	return "", ErrMultipleRecords
}

func isTemporary(err error) bool {
	dnsErr, ok := err.(*net.DNSError)
	return ok && dnsErr.Temporary()
}

func isNotFound(err error) bool {
	dnsErr, ok := err.(*net.DNSError)
	return ok && dnsErr.IsNotFound
}

// Check if the given DNS error is a "void lookup" (0 answers, or nxdomain),
// and if so increment the void lookup counter.
func (r *resolution) checkVoidLookup(answerCount int, err error) {
	if err == nil && answerCount == 0 {
		r.voidCount++
		return
	}

	dnsErr, ok := err.(*net.DNSError)
	if !ok {
		return
	}

	if dnsErr.IsNotFound {
		r.voidCount++
	}
}

// ipField processes an "ip" field.
func (r *resolution) ipField(res Result, field string) (bool, Result, error) {
	fip := field[4:]
	if strings.Contains(fip, "/") {
		_, ipNet, err := net.ParseCIDR(fip)
		if err != nil {
			return true, PermError, ErrInvalidMask
		}
		if ipNet.Contains(r.ip) {
			return true, res, ErrMatchedIP
		}
	} else {
		ip := net.ParseIP(fip)
		if ip == nil {
			return true, PermError, ErrInvalidIP
		}
		if ip.Equal(r.ip) {
			return true, res, ErrMatchedIP
		}
	}

	return false, "", nil
}

// ptrField processes a "ptr" field.
func (r *resolution) ptrField(res Result, field, domain string) (bool, Result, error) {
	// Extract the domain if the field is in the form "ptr:domain".
	ptrDomain := domain
	if len(field) >= 4 {
		ptrDomain = field[4:]

	}
	ptrDomain, err := r.expandMacros(ptrDomain, domain)
	if err != nil {
		return true, PermError, ErrInvalidMacro
	}

	if ptrDomain == "" {
		return true, PermError, ErrInvalidDomain
	}

	if r.ipNames == nil {
		r.ipNames = []string{}
		r.count++
		ns, err := r.resolver.LookupPtr(r.ip.String())
		r.checkVoidLookup(len(ns), err)
		if err != nil {
			// https://tools.ietf.org/html/rfc7208#section-5
			if isNotFound(err) {
				return false, "", err
			}
			return true, TempError, err
		}

		// Only take the first 10 names, ignore the rest.
		// Each A/AAAA lookup in this context is NOT included in the overall
		// count. The RFC defines this separate logic and limits.
		// https://datatracker.ietf.org/doc/html/rfc7208#section-4.6.4
		if len(ns) > 10 {
			ns = ns[:10]
		}

		for _, n := range ns {
			// Validate the record by doing a forward resolution: it has to
			// have some A/AAAA.
			ips, err := r.resolver.LookupIPAddr(n)
			if err != nil {
				// RFC explicitly says to skip domains which error here.
				continue
			}
			if len(ips) > 0 {
				// Append the lower-case variants so we do a case-insensitive
				// lookup below.
				r.ipNames = append(r.ipNames, strings.ToLower(n))
			}
		}
	}

	ptrDomain = strings.ToLower(ptrDomain)
	for _, n := range r.ipNames {
		if strings.HasSuffix(n, ptrDomain+".") {
			return true, res, ErrMatchedPTR
		}
	}

	return false, "", nil
}

// existsField processes a "exists" field.
// https://tools.ietf.org/html/rfc7208#section-5.7
func (r *resolution) existsField(res Result, field, domain string) (bool, Result, error) {
	// The field is in the form "exists:<domain>".
	eDomain := field[7:]
	eDomain, err := r.expandMacros(eDomain, domain)
	if err != nil {
		return true, PermError, ErrInvalidMacro
	}

	if eDomain == "" {
		return true, PermError, ErrInvalidDomain
	}

	r.count++
	ips, err := r.resolver.LookupIPAddr(eDomain)
	r.checkVoidLookup(len(ips), err)
	if err != nil {
		// https://tools.ietf.org/html/rfc7208#section-5
		if isNotFound(err) {
			return false, "", err
		}
		return true, TempError, err
	}

	// Exists only counts if there are IPv4 matches.
	for _, ip := range ips {
		// convert ip string to net.IP
		ipObj := net.ParseIP(ip)
		if ipObj.To4() != nil {
			return true, res, ErrMatchedExists
		}
	}
	return false, "", nil
}

// includeField processes an "include" field.
func (r *resolution) includeField(res Result, field, domain string) (bool, Result, error) {
	// https://tools.ietf.org/html/rfc7208#section-5.2
	incDomain := field[len("include:"):]
	incDomain, err := r.expandMacros(incDomain, domain)
	if err != nil {
		return true, PermError, ErrInvalidMacro
	}
	r.count++
	ir, err := r.Check(incDomain)
	switch ir {
	case Pass:
		return true, res, err
	case Fail, SoftFail, Neutral:
		return false, ir, err
	case TempError:
		return true, TempError, err
	case PermError:
		return true, PermError, err
	case None:
		return true, PermError, err
	}

	return false, "", fmt.Errorf("this should never be reached")
}

type dualMasks struct {
	v4 net.IPMask
	v6 net.IPMask
}

func maskToStr(m net.IPMask) string {
	ones, bits := m.Size()
	if ones == 0 && bits == 0 {
		return m.String()
	}
	return fmt.Sprintf("/%d", ones)
}

func (m dualMasks) String() string {
	return fmt.Sprintf("[%v, %v]", maskToStr(m.v4), maskToStr(m.v6))
}

func ipMatch(ip, distIp net.IP, masks dualMasks) bool {
	mask := net.IPMask(nil)
	if distIp.To4() != nil && masks.v4 != nil {
		mask = masks.v4
	} else if distIp.To4() == nil && masks.v6 != nil {
		mask = masks.v6
	}

	if mask != nil {
		ipNet := net.IPNet{IP: distIp, Mask: mask}
		return ipNet.Contains(ip)
	}

	return ip.Equal(distIp)
}

var aRegexp = regexp.MustCompile(`^[aA](:([^/]+))?(/(\w+))?(//(\w+))?$`)
var mxRegexp = regexp.MustCompile(`^[mM][xX](:([^/]+))?(/(\w+))?(//(\w+))?$`)

func domainAndMask(re *regexp.Regexp, field, domain string) (string, dualMasks, error) {
	masks := dualMasks{}
	groups := re.FindStringSubmatch(field)
	if groups != nil {
		if groups[2] != "" {
			domain = groups[2]
		}
		if groups[4] != "" {
			i, err := strconv.Atoi(groups[4])
			mask4 := net.CIDRMask(i, 32)
			if err != nil || mask4 == nil {
				return "", masks, ErrInvalidMask
			}
			masks.v4 = mask4
		}
		if groups[6] != "" {
			i, err := strconv.Atoi(groups[6])
			mask6 := net.CIDRMask(i, 128)
			if err != nil || mask6 == nil {
				return "", masks, ErrInvalidMask
			}
			masks.v6 = mask6
		}
	}

	// Test to catch malformed entries: if there's a /, there must be at least
	// one mask.
	if strings.Contains(field, "/") && masks.v4 == nil && masks.v6 == nil {
		return "", masks, ErrInvalidMask
	}

	return domain, masks, nil
}

// aField processes an "a" field.
func (r *resolution) aField(res Result, field, domain string) (bool, Result, error) {
	// https://tools.ietf.org/html/rfc7208#section-5.3
	aDomain, masks, err := domainAndMask(aRegexp, field, domain)
	if err != nil {
		return true, PermError, err
	}
	aDomain, err = r.expandMacros(aDomain, domain)
	if err != nil {
		return true, PermError, ErrInvalidMacro
	}

	r.count++
	ips, err := r.resolver.LookupIPAddr(aDomain)
	r.checkVoidLookup(len(ips), err)
	if err != nil {
		// https://tools.ietf.org/html/rfc7208#section-5
		if isNotFound(err) {
			return false, "", err
		}
		return true, TempError, err
	}
	for _, ip := range ips {
		if ipMatch(r.ip, net.ParseIP(ip), masks) {
			return true, res, ErrMatchedA
		}
	}

	return false, "", nil
}

// mxField processes "mx" field.
func (r *resolution) mxField(res Result, field, domain string) (bool, Result, error) {
	// https://tools.ietf.org/html/rfc7208#section-5.4
	mxDomain, masks, err := domainAndMask(mxRegexp, field, domain)
	if err != nil {
		return true, PermError, err
	}
	mxDomain, err = r.expandMacros(mxDomain, domain)
	if err != nil {
		return true, PermError, ErrInvalidMacro
	}

	r.count++
	mxs, err := r.resolver.LookupMX(mxDomain)
	r.checkVoidLookup(len(mxs), err)

	// If we get some results, use them even if we get an error alongisde.
	// This happens when one of the records is invalid, because Go library can
	// be quite strict about it. The RFC is not clear about this specific
	// situation, and other SPF libraries and implementations just skip the
	// invalid value, so we match common practice.
	if err != nil && len(mxs) == 0 {
		// https://tools.ietf.org/html/rfc7208#section-5
		if isNotFound(err) {
			return false, "", err
		}
		return true, TempError, err
	}

	// There's an explicit maximum of 10 MX records per match.
	// https://tools.ietf.org/html/rfc7208#section-4.6.4
	if len(mxs) > 10 {
		return true, PermError, ErrTooManyMXRecords
	}

	mxIpList := []net.IP{}
	for _, mx := range mxs {
		ips, err := r.resolver.LookupIPAddr(mx)
		if err != nil {
			// If the address of the MX record was not found, we just skip it.
			// https://tools.ietf.org/html/rfc7208#section-5
			if isNotFound(err) {
				continue
			}
			return true, TempError, err
		}
		for _, ip := range ips {
			mxIpList = append(mxIpList, net.ParseIP(ip))
		}
	}

	for _, ip := range mxIpList {
		if ipMatch(r.ip, ip, masks) {
			return true, res, ErrMatchedMX
		}
	}

	return false, "", nil
}

// redirectField processes a "redirect=" field.
func (r *resolution) redirectField(field, domain string) (Result, error) {
	rDomain := field[len("redirect="):]
	rDomain, err := r.expandMacros(rDomain, domain)
	if err != nil {
		return PermError, ErrInvalidMacro
	}

	if rDomain == "" {
		return PermError, ErrInvalidDomain
	}

	// https://tools.ietf.org/html/rfc7208#section-6.1
	r.count++
	result, err := r.Check(rDomain)
	if result == None {
		result = PermError
	}
	return result, err
}

// Group extraction of macro-string from the formal specification.
// https://tools.ietf.org/html/rfc7208#section-7.1
var macroRegexp = regexp.MustCompile(
	`([slodiphcrtvSLODIPHCRTV])([0-9]+)?([rR])?([-.+,/_=]+)?`)

// Expand macros, return the expanded string.
// This expects to be passed the domain-spec within a field, not an entire
// field or larger (that has problematic security implications).
// https://tools.ietf.org/html/rfc7208#section-7
func (r *resolution) expandMacros(s, domain string) (string, error) {
	// Macros/domains shouldn't contain CIDR. Our parsing should prevent it
	// from happening in case where it matters (a, mx), but for the ones which
	// doesn't, prevent them from sneaking through.
	if strings.Contains(s, "/") {
		return "", ErrInvalidDomain
	}

	// Bypass the complex logic if there are no macros present.
	if !strings.Contains(s, "%") {
		return s, nil
	}

	// Are we processing the character right after "%"?
	afterPercent := false

	// Are we inside a macro definition (%{...}) ?
	inMacroDefinition := false

	// Macro string, where we accumulate the values inside the definition.
	macroS := ""

	var err error
	n := ""
	for _, c := range s {
		if afterPercent {
			afterPercent = false
			switch c {
			case '%':
				n += "%"
				continue
			case '_':
				n += " "
				continue
			case '-':
				n += "%20"
				continue
			case '{':
				inMacroDefinition = true
				continue
			}
			return "", ErrInvalidMacro
		}
		if inMacroDefinition {
			if c != '}' {
				macroS += string(c)
				continue
			}
			inMacroDefinition = false

			// Extract letter, digit transformer, reverse transformer, and
			// delimiters.
			groups := macroRegexp.FindStringSubmatch(macroS)
			macroS = ""
			if groups == nil {
				return "", ErrInvalidMacro
			}
			letter := groups[1]

			digits := 0
			if groups[2] != "" {
				// Use 0 as "no digits given"; an explicit value of 0 is not
				// valid.
				digits, err = strconv.Atoi(groups[2])
				if err != nil || digits <= 0 {
					return "", ErrInvalidMacro
				}
			}
			reverse := groups[3] == "r" || groups[3] == "R"
			delimiters := groups[4]
			if delimiters == "" {
				// By default, split strings by ".".
				delimiters = "."
			}

			// Uppercase letters indicate URL escaping of the results.
			urlEscape := letter == strings.ToUpper(letter)
			letter = strings.ToLower(letter)

			str := ""
			switch letter {
			case "s":
				str = r.sender
			case "l":
				str, _ = split(r.sender)
			case "o":
				_, str = split(r.sender)
			case "d":
				str = domain
			case "i":
				str = ipToMacroStr(r.ip)
			case "p":
				// This shouldn't be used, we don't want to support it, it's
				// risky. "unknown" is a safe value.
				// https://tools.ietf.org/html/rfc7208#section-7.3
				str = "unknown"
			case "v":
				if r.ip.To4() != nil {
					str = "in-addr"
				} else {
					str = "ip6"
				}
			case "h":
				str = r.helo
			default:
				// c, r, t are allowed in exp only, and we don't expand macros
				// in exp so they are just as invalid as the rest.
				return "", ErrInvalidMacro
			}

			// Split str using the given separators.
			splitFunc := func(r rune) bool {
				return strings.ContainsRune(delimiters, r)
			}
			split := strings.FieldsFunc(str, splitFunc)

			// Reverse if requested.
			if reverse {
				reverseStrings(split)
			}

			// Leave the last $digits fields, if given.
			if digits > 0 {
				if digits > len(split) {
					digits = len(split)
				}
				split = split[len(split)-digits:]
			}

			// Join back, always with "."
			str = strings.Join(split, ".")

			// Escape if requested. Note this doesn't strictly escape ALL
			// unreserved characters, it's the closest we can get without
			// reimplmenting it ourselves.
			if urlEscape {
				str = url.QueryEscape(str)
			}

			n += str
			continue
		}
		if c == '%' {
			afterPercent = true
			continue
		}
		n += string(c)
	}

	return n, nil
}

func reverseStrings(a []string) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

func ipToMacroStr(ip net.IP) string {
	if ip.To4() != nil {
		return ip.String()
	}

	// For IPv6 addresses, the "i" macro expands to a dot-format address.
	// https://datatracker.ietf.org/doc/html/rfc7208#section-7.3
	sb := strings.Builder{}
	sb.Grow(64)
	for _, b := range ip.To16() {
		fmt.Fprintf(&sb, "%x.%x.", b>>4, b&0xf)
	}
	// Return the string without the trailing ".".
	return sb.String()[:sb.Len()-1]
}
