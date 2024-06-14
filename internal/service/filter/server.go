package filter

import (
	"context"
	"easymail/internal/easylog"
	"easymail/internal/model"
	"easymail/internal/preprocessing"
	"easymail/internal/service/milter"
	"fmt"
	"net"
	"net/textproto"
	"strconv"
	"sync"
	"time"
)

type Filter struct {
	milter.Milter
	optAction   milter.OptAction
	optProtocol milter.OptProtocol
}

const timeout = 100000 * time.Microsecond

func (f *Filter) Connect(host string, addr net.IP, payload map[string]string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	feature := make([]milter.Feature, 0)

	if metrics, err := model.GetFilterMetricByStage(model.FilterStageConnect); err == nil {
		for _, metric := range metrics {
			if max(metric.PrimaryField.Stage, metric.SecondaryField.Stage) != model.FilterStageConnect {
				continue
			}
			if metric.Category == model.FilterCategoryAll {
				key := fmt.Sprintf("%s:%s", metric.MakeFilterMetricKey(), addr.String())
				if cnt, err := increaseCount(context.Background(), key, metric.MakeFilterMetricTimeout()); err == nil {
					feature = append(feature, milter.Feature{
						Name:      metric.Name,
						Value:     strconv.FormatInt(cnt, 10),
						ValueType: milter.DataTypeInt,
					})
				}
			}
		}
	}

	return milter.RespContinue, feature, nil
}

func (f *Filter) Helo(name string, payload map[string]string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	//add feature of helo argument
	feature := make([]milter.Feature, 0)
	return milter.RespContinue, feature, nil
}

func (f *Filter) MailFrom(sender string, payload map[string]string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	feature := make([]milter.Feature, 0)
	if len(sender) > 0 {
		// default feature engineer

		if d := preprocessing.GetDomain(sender); d != "" {
			//1. check sender domain exist
			if exist, err := resolver.DomainExist(d); !exist || err != nil {
				feature = append(feature, milter.Feature{
					Name:      "sender_domain_exist",
					Value:     "false",
					ValueType: milter.DataTypeBool,
				})
			}

			// 2. check spf
		}
		if metrics, err := model.GetFilterMetricByStage(model.FilterStageMailFrom); err == nil {
			for _, metric := range metrics {
				if max(metric.PrimaryField.Stage, metric.SecondaryField.Stage) != model.FilterStageMailFrom {
					continue
				}
				if metric.Category != model.FilterCategoryAll {
					continue
				}
				if metric.SecondaryFieldID == 0 {
					// no secondary field, only support count operation
					if metric.Operation == model.MetricOperationCount {
						if len(sender) > 0 {
							key := fmt.Sprintf("%s:%s", metric.MakeFilterMetricKey(), sender)
							if cnt, err := increaseCount(context.Background(), key, metric.MakeFilterMetricTimeout()); err == nil {
								feature = append(feature, milter.Feature{
									Name:      metric.Name,
									Value:     strconv.FormatInt(cnt, 10),
									ValueType: milter.DataTypeInt,
								})
							}
						}
					}
				} else {
					// has secondary field, support count and collect operation, but primary and secondary field must not be empty
					pk, okp := payload[metric.PrimaryField.Name]
					sk, oks := payload[metric.SecondaryField.Name]
					if !(okp && oks && len(pk) > 0 && len(sk) > 0) {
						continue
					}
					if metric.Operation == model.MetricOperationCount {
						key := fmt.Sprintf("%s:%s:%s", metric.MakeFilterMetricKey(), pk, sk)
						if cnt, err := increaseCount(context.Background(), key, metric.MakeFilterMetricTimeout()); err == nil {
							feature = append(feature, milter.Feature{
								Name:      metric.Name,
								Value:     strconv.FormatInt(cnt, 10),
								ValueType: milter.DataTypeInt,
							})
						}
					} else if metric.Operation == model.MetricOperationCollect {
						key := fmt.Sprintf("%s:%s", metric.MakeFilterMetricKey(), pk)
						if cnt, err := addSet(context.Background(), key, sk, metric.MakeFilterMetricTimeout()); err == nil {
							feature = append(feature, milter.Feature{
								Name:      metric.Name,
								Value:     strconv.FormatInt(cnt, 10),
								ValueType: milter.DataTypeInt,
							})
						}
					}
				}

			}
		}
	}
	return milter.RespContinue, feature, nil
}

func (f *Filter) RcptTo(receipt string, payload map[string]string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	feature := make([]milter.Feature, 0)
	if len(receipt) > 0 {
		if metrics, err := model.GetFilterMetricByStage(model.FilterStageRcptTo); err == nil {
			for _, metric := range metrics {
				if max(metric.PrimaryField.Stage, metric.SecondaryField.Stage) != model.FilterStageRcptTo {
					continue
				}
				if metric.Category != model.FilterCategoryAll {
					continue
				}
				if metric.SecondaryFieldID == 0 {
					// no secondary field, only support count operation
					if metric.Operation == model.MetricOperationCount {
						if len(receipt) > 0 {
							key := fmt.Sprintf("%s:%s", metric.MakeFilterMetricKey(), receipt)
							if cnt, err := increaseCount(context.Background(), key, metric.MakeFilterMetricTimeout()); err == nil {
								feature = append(feature, milter.Feature{
									Name:      metric.Name,
									Value:     strconv.FormatInt(cnt, 10),
									ValueType: milter.DataTypeInt,
								})
							}
						}
					}
				} else {
					// has secondary field, support count and collect operation, but primary and secondary field must not be empty
					pk, okp := payload[metric.PrimaryField.Name]
					sk, oks := payload[metric.SecondaryField.Name]
					if !(okp && oks && len(pk) > 0 && len(sk) > 0) {
						continue
					}
					if metric.Operation == model.MetricOperationCount {
						key := fmt.Sprintf("%s:%s:%s", metric.MakeFilterMetricKey(), pk, sk)
						if cnt, err := increaseCount(context.Background(), key, metric.MakeFilterMetricTimeout()); err == nil {
							feature = append(feature, milter.Feature{
								Name:      metric.Name,
								Value:     strconv.FormatInt(cnt, 10),
								ValueType: milter.DataTypeInt,
							})
						}
					} else if metric.Operation == model.MetricOperationCollect {
						key := fmt.Sprintf("%s:%s", metric.MakeFilterMetricKey(), pk)
						if cnt, err := addSet(context.Background(), key, sk, metric.MakeFilterMetricTimeout()); err == nil {
							feature = append(feature, milter.Feature{
								Name:      metric.Name,
								Value:     strconv.FormatInt(cnt, 10),
								ValueType: milter.DataTypeInt,
							})
						}
					}
				}
			}
		}
	}
	return milter.RespContinue, feature, nil
}

func (f *Filter) Header(name string, value string, payload map[string]string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	return milter.RespContinue, nil, nil
}

func (f *Filter) Headers(h textproto.MIMEHeader, payload map[string]string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	feature := make([]milter.Feature, 0)
	if h != nil {
		if metrics, err := model.GetFilterMetricByStage(model.FilterStageHeader); err == nil {
			for _, metric := range metrics {
				if max(metric.PrimaryField.Stage, metric.SecondaryField.Stage) != model.FilterStageHeader {
					continue
				}
				if metric.Category != model.FilterCategoryAll {
					continue
				}
				if metric.SecondaryFieldID == 0 {
					// no secondary field, only support count operation
					if metric.Operation == model.MetricOperationCount {
						if v, ok := payload[metric.PrimaryField.Name]; ok && len(v) > 0 {
							key := fmt.Sprintf("%s:%s", metric.MakeFilterMetricKey(), v)
							if cnt, err := increaseCount(context.Background(), key, metric.MakeFilterMetricTimeout()); err == nil {
								feature = append(feature, milter.Feature{
									Name:      metric.Name,
									Value:     strconv.FormatInt(cnt, 10),
									ValueType: milter.DataTypeInt,
								})
							}
						}
					}
				} else {
					// has secondary field, support count and collect operation, but primary and secondary field must not be empty
					pk, okp := payload[metric.PrimaryField.Name]
					sk, oks := payload[metric.SecondaryField.Name]
					if !(okp && oks && len(pk) > 0 && len(sk) > 0) {
						continue
					}
					if metric.Operation == model.MetricOperationCount {
						key := fmt.Sprintf("%s:%s:%s", metric.MakeFilterMetricKey(), pk, sk)
						if cnt, err := increaseCount(context.Background(), key, metric.MakeFilterMetricTimeout()); err == nil {
							feature = append(feature, milter.Feature{
								Name:      metric.Name,
								Value:     strconv.FormatInt(cnt, 10),
								ValueType: milter.DataTypeInt,
							})
						}
					} else if metric.Operation == model.MetricOperationCollect {
						key := fmt.Sprintf("%s:%s", metric.MakeFilterMetricKey(), pk)
						if cnt, err := addSet(context.Background(), key, sk, metric.MakeFilterMetricTimeout()); err == nil {
							feature = append(feature, milter.Feature{
								Name:      metric.Name,
								Value:     strconv.FormatInt(cnt, 10),
								ValueType: milter.DataTypeInt,
							})
						}
					}
				}
			}
		}
	}
	return milter.RespContinue, feature, nil
}

func (f *Filter) BodyChunk(chunk []byte, payload map[string]string, m *milter.Modifier) (milter.Response, []milter.Feature, error) {
	return milter.RespContinue, nil, nil
}

func (f *Filter) Body(payload map[string]string, m *milter.Modifier, macro map[string]string) (milter.Response, []milter.Feature, error) {
	if v, ok := macro["i"]; ok {
		m.AddHeader("X-Queue-ID", v)
	}
	return milter.RespContinue, nil, nil
}
func (f *Filter) Abort(m *milter.Modifier) error {
	return nil
}

type Server struct {
	name      string
	stopCh    chan struct{}
	started   bool
	lock      *sync.Mutex
	family    string
	listen    string
	debug     bool
	_log      *easylog.Logger
	html2text *preprocessing.Html2Text

	filter *Filter
}

func New(family string, listen string, _log *easylog.Logger) (server *Server) {
	server = &Server{
		name:    "filter",
		stopCh:  make(chan struct{}),
		lock:    &sync.Mutex{},
		started: false,
		family:  family,
		listen:  listen,

		filter: &Filter{
			optProtocol: milter.OptProtocol(milter.OptChangeBody | milter.OptChangeFrom | milter.OptChangeHeader | milter.OptAddHeader | milter.OptAddRcpt | milter.OptChangeFrom),
			optAction:   0,
		},
		_log:      _log,
		html2text: preprocessing.NewHtml2Text(nil),
	}
	return server
}

func (s *Server) SetDebug(debug bool) {
	s.debug = debug
}

func (s *Server) SetLogger(_log *easylog.Logger) error {
	if _log == nil {
		return fmt.Errorf("%s logger is nil", s.name)
	}
	s._log = _log
	return nil
}

func (s *Server) Start() error {
	if s._log == nil {
		return fmt.Errorf("%s logger is nil", s.name)
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.started {
		return fmt.Errorf("%s server already started", s.name)
	}

	s.started = true
	s._log.Infof("%s server started!", s.name)
	go s.run()
	return nil
}

func (s *Server) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.started {
		return fmt.Errorf("%s server not started", s.name)
	}

	s.started = false
	close(s.stopCh)
	s._log.Infof("%s server stopped!", s.name)
	return nil
}

func (s *Server) Name() string {
	return s.name
}

func (s *Server) run() (err error) {
	listener, err := net.Listen(s.family, s.listen)
	if err != nil {
		return err
	}

	for {
		client, err := listener.Accept()
		if err != nil {
			return err
		}
		session := NewSession(client, s.filter, s._log, s.html2text)
		go session.Handle()
	}
}
