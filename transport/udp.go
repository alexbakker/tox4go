package transport

import (
	"errors"
	"net"
)

type UDPTransport struct {
	conn           *net.UDPConn
	stopChan       chan struct{}
	packetHandlers map[byte]Handler
	handler        Handler
}

func NewUDPTransport(netProto string, addr string) (*UDPTransport, error) {
	udpAddr, err := net.ResolveUDPAddr(netProto, addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP(netProto, udpAddr)
	if err != nil {
		return nil, err
	}

	return &UDPTransport{
		conn:           conn,
		stopChan:       make(chan struct{}),
		packetHandlers: map[byte]Handler{},
	}, nil
}

func (t *UDPTransport) Send(data []byte, addr *net.UDPAddr) error {
	_, err := t.conn.WriteTo(data, addr)
	return err
}

func (t *UDPTransport) Handle(handler Handler) {
	t.handler = handler
}

func (t *UDPTransport) HandlePacket(packetID byte, handler Handler) {
	t.packetHandlers[packetID] = handler
}

func (t *UDPTransport) Listen() error {
	buf := make([]byte, 2048)
	for {
		read, senderAddr, err := t.conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				close(t.stopChan)
			}
			return err
		}

		if read < 1 {
			continue
		}

		data := buf[:read]
		if handler, ok := t.packetHandlers[data[0]]; ok {
			handler(data, senderAddr)
		} else if t.handler != nil {
			t.handler(data, senderAddr)
		}
	}
}

func (t *UDPTransport) Close() error {
	err := t.conn.Close()
	<-t.stopChan
	return err
}
