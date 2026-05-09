package messaging

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	Conn *nats.Conn
}

func NewNATSClient(url string) (*NATSClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}
	
	log.Println("connected to NATS")
	return &NATSClient{Conn: nc}, nil
}

func (n *NATSClient) Close() {
	if n.Conn != nil {
		n.Conn.Close()
	}
}

func (n *NATSClient) Publish(subject string, data []byte) error {
	return n.Conn.Publish(subject, data)
}
