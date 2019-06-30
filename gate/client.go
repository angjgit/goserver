package gate

import (
	"fmt"
	"github.com/0990/goserver/network"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"reflect"
)

type Client struct {
	conn network.Conn
	gate *Gate
}

func NewClient(conn network.Conn, gate *Gate) *Client {
	return &Client{
		conn: conn,
		gate: gate,
	}
}

func (p *Client) ReadLoop() {
	for {
		data, err := p.conn.ReadMsg()
		if err != nil {
			fmt.Printf("read message: %v", err)
			break
		}

		msg, err := p.gate.Processor.Unmarshal(data)
		if err != nil {
			logrus.Debugf("unmarshal message error: %v", err)
			break
		}

		p.gate.Post(func() {
			err = p.gate.Processor.Route(msg, p)
		})
	}
}

func (p *Client) OnClose() {
	fmt.Println("client close")
	p.gate.Post(func() {
		p.gate.closeEvent(p)
	})
}

func (p *Client) WriteMsg(msg proto.Message) {
	data, err := p.gate.Processor.Marshal(msg)
	if err != nil {
		logrus.Errorf("marshal message %v error: %v", reflect.TypeOf(msg), err)
		return
	}
	err = p.conn.WriteMsg(data)
	if err != nil {
		logrus.Error("write message %v error: %v", reflect.TypeOf(msg), err)
	}
}