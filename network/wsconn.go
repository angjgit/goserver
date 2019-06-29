package network

import (
	"github.com/gorilla/websocket"
	"net"
	"sync"
)

type WSConn struct {
	sync.Mutex
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	closeFlag bool
}

func NewWSConn(conn *websocket.Conn) *WSConn {
	wsc := new(WSConn)
	wsc.send = make(chan []byte, 256)
	wsc.conn = conn
	return wsc
}

//func(c *WSConn)Run(){
//	go c.writePump()
//}

func (p *WSConn) writePump() {
	for data := range p.send {
		if data == nil {
			break
		}
		err := p.conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			break
		}
	}
	p.conn.Close()
	p.Lock()
	p.closeFlag = true
	p.Unlock()
}

func (p *WSConn) WriteMsg(args ...[]byte) error {
	p.Lock()
	defer p.Unlock()

	if p.closeFlag {
		return nil
	}

	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	if len(args) == 1 {
		return p.doWrite(args[0])
	}

	msg := make([]byte, msgLen)
	l := 0
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}
	return p.doWrite(msg)
}

func (p *WSConn) doWrite(data []byte) error {
	//TODO send chan 堵满情况处理
	p.send <- data
	return nil
}

func (p *WSConn) ReadMsg() ([]byte, error) {
	_, data, err := p.conn.ReadMessage()
	return data, err
}

func (p *WSConn) Close() {
	p.Lock()
	defer p.Unlock()
	if p.closeFlag {
		return
	}
	p.doWrite(nil)
	p.closeFlag = true
}

func (p *WSConn) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *WSConn) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}
