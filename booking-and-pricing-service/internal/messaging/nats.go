package messaging

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	log.Println("nats publisher connected")
	return &Publisher{conn: nc}, nil
}

func (p *Publisher) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := p.conn.Publish(subject, payload); err != nil {
		log.Printf("nats publish error on %s: %v", subject, err)
		return err
	}
	return nil
}

func (p *Publisher) Close() {
	p.conn.Close()
}

type Subscriber struct {
	conn *nats.Conn
}

func NewSubscriber(url string) (*Subscriber, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	log.Println("nats subscriber connected")
	return &Subscriber{conn: nc}, nil
}

func (s *Subscriber) Subscribe(subject string, handler func([]byte)) error {
	_, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	return err
}

func (s *Subscriber) Close() {
	s.conn.Close()
}
