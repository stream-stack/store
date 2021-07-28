package publisher

import (
	"github.com/stream-stack/publisher/pkg/proto"
	"log"
	"regexp"
)

type Subscriber struct {
	Name       string     `json:"name"`
	Pattern    string     `json:"pattern"`
	Url        string     `json:"url"`
	StartPoint StartPoint `json:"start_point"`

	regExp *regexp.Regexp
}

func (s *Subscriber) init() error {
	compile, err := regexp.Compile(s.Pattern)
	if err != nil {
		return err
	}
	s.regExp = compile
	return nil
}

func (s *Subscriber) Match(key string) bool {
	return s.regExp.MatchString(key)
}

func (s *Subscriber) Send(event proto.SaveRequest) {
	//go func() {
	//
	//}()
	log.Printf("发送数据:%+v,url:%+v", event, s.Url)
	//todo:实现发送
	//client := http.Client{}
	//client.Post(e.Url)
}
