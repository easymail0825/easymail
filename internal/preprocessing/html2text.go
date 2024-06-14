package preprocessing

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
	"sync"
	"unicode"
)

type Html2Text struct {
	htmlTagMapper map[string]struct{}
}

var onceHtml2Text sync.Once

func NewHtml2Text(tags []string) *Html2Text {
	var h *Html2Text
	var tagMapper = make(map[string]struct{})
	if len(tags) == 0 {
		tags = defaultHtmlTags
	}
	for _, tag := range tags {
		tagMapper[tag] = struct{}{}
	}
	onceHtml2Text.Do(func() {
		h = &Html2Text{
			htmlTagMapper: tagMapper,
		}
	})
	return h
}

func (h *Html2Text) matchHtmlProperty(src, key, value string) bool {
	src = strings.ToLower(src)
	keyIndex := strings.Index(src, key+":")
	valueIndex := strings.Index(src, value)
	if keyIndex < 0 || valueIndex < 0 {
		return false
	}
	for i := keyIndex + len(key) + 1; i < valueIndex; i++ {
		if !unicode.IsSpace(rune(src[i])) {
			return false
		}
	}
	return true
}

func (h *Html2Text) mapHtmlProperty(style string) map[string]string {
	var mapper = make(map[string]string)
	style = strings.ToLower(style)
	styles := strings.Split(style, ";")
	for _, s := range styles {
		d := strings.SplitN(s, ":", 2)
		if len(d) == 2 {
			mapper[strings.Trim(d[0], " ")] = strings.Trim(d[1], " ")
		}
	}
	return mapper
}

func (h *Html2Text) Parse(src string) (text, url []string) {
	tagReplace := map[string]string{
		"<h1>":  "<span>",
		"<h2>":  "<span>",
		"<h3>":  "<span>",
		"<h4>":  "<span>",
		"<h5>":  "<span>",
		"<h6>":  "<span>",
		"<h7>":  "<span>",
		"</h1>": "</span>",
		"</h2>": "</span>",
		"</h3>": "</span>",
		"</h4>": "</span>",
		"</h5>": "</span>",
		"</h6>": "</span>",
		"</h7>": "</span>",
		"<H1>":  "<span>",
		"<H2>":  "<span>",
		"<H3>":  "<span>",
		"<H4>":  "<span>",
		"<H5>":  "<span>",
		"<H6>":  "<span>",
		"<H7>":  "<span>",
		"</H1>": "</span>",
		"</H2>": "</span>",
		"</H3>": "</span>",
		"</H4>": "</span>",
		"</H5>": "</span>",
		"</H6>": "</span>",
		"</H7>": "</span>",
	}
	for oldTag, newTag := range tagReplace {
		src = strings.ReplaceAll(src, oldTag, newTag)
	}
	document, err := goquery.NewDocumentFromReader(strings.NewReader(src))
	if err != nil {
		return nil, nil
	}
	document.Find("noscript").Remove()
	document.Find("script").Remove()
	document.Find("style").Remove()
	document.Find(`span[style*="display:none"]`).Remove()
	document.Find(`div[style*="display:none"]`).Remove()

	// data save the text in array style
	title := document.Find("title").Contents().Text()
	text = append(text, title)

	// drop hidden text by special style
	document.Find("body *").Each(func(i int, s *goquery.Selection) {
		node := s.Get(0)
		name := node.Data
		// ignore unknown html tags
		if _, ok := h.htmlTagMapper[name]; !ok {
			s.Empty()
		}

		if style, ok := s.Attr("style"); ok {
			m := h.mapHtmlProperty(style)
			width := m["width"]
			height := m["height"]
			font := m["font"]
			fontSize := m["font-size"]
			if strings.HasPrefix(width, "0") || strings.HasPrefix(width, "-") {
				s.Remove()
			}
			if strings.HasPrefix(height, "0") || strings.HasPrefix(height, "-") {
				s.Remove()
			}
			if strings.HasPrefix(font, "0") || strings.HasPrefix(font, "-") {
				s.Remove()
			}
			if strings.HasPrefix(fontSize, "0") || strings.HasPrefix(fontSize, "-") {
				s.Remove()
			}
			if m["display"] == "none" {
				s.Remove()
			}
		}
	})

	// select link text
	document.Find("body *").Each(func(i int, s *goquery.Selection) {
		if link, ok := s.Attr("href"); ok {
			linkText := s.Text()
			text = append(text, linkText)
			url = append(url, link)
		}
		if imgSrc, ok := s.Attr("src"); ok {
			imgAlt, _ := s.Attr("alt")
			if strings.Index(strings.ToLower(imgSrc), "data:") > 0 {
				url = append(url, imgSrc)
			}
			if len(imgAlt) > 0 {
				text = append(text, imgAlt)
			}
		}
	})

	// select meta text
	document.Find("meta").Each(func(i int, s *goquery.Selection) {
		if meta, ok := s.Attr("http-equiv"); ok {
			if strings.ToLower(meta) == "refresh" {
				if c, ok := s.Attr("text"); ok {
					d := strings.SplitN(c, "=", 2)
					if len(d) == 2 {
						url = append(url, d[1])
					}
				}
			}
		}
	})

	// select video text
	document.Find("video").Each(func(i int, s *goquery.Selection) {
		if poster, ok := s.Attr("poster"); ok {
			poster = strings.ReplaceAll(poster, "\r", "")
			poster = strings.ReplaceAll(poster, "\n", "")
			url = append(url, poster)
		}
	})

	// select innerText from tags in body
	document.Find("body").Each(func(i int, s *goquery.Selection) {
		innerText := s.Text()
		if len(innerText) > 0 {
			text = append(text, innerText)
		}
	})

	// select innerText from tags in input control
	document.Find("input").Each(func(i int, s *goquery.Selection) {
		innerText, _ := s.Attr("value")
		if len(innerText) > 0 {
			text = append(text, innerText)
		}
	})

	return
}
