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

	action SubscriberAction
	regExp *regexp.Regexp
}

type SubscriberAction func(subscriber *Subscriber, request proto.SaveRequest) error

//TODO:url的action如何执行？
func NewUrlSubscriber(name, pattern, url string, point StartPoint) *Subscriber {
	return &Subscriber{
		Name:       name,
		Pattern:    pattern,
		Url:        url,
		StartPoint: point,
		action: func(subscriber *Subscriber, request proto.SaveRequest) error {
			//todo：执行request请求
			return nil
		},
	}
}

func (s *Subscriber) init() error {
	compile, err := regexp.Compile(s.Pattern)
	if err != nil {
		return err
	}
	s.regExp = compile
	return nil
}

//todo:判断point是否符合
func (s *Subscriber) Match(key string) bool {
	return s.regExp.MatchString(key)
}

func (s *Subscriber) DoAction(event proto.SaveRequest) {
	s.action(s, event)
	//go func() {
	//
	//}()
	log.Printf("发送数据:%+v,url:%+v", event, s.Url)
	//todo:实现发送
	//client := http.Client{}
	//client.Post(e.Url)
}
