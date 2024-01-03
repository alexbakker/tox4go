package transport

import (
	"errors"
	"net"
)

type UDPTransport struct {
	conn     *net.UDPConn
	stopChan chan struct{}
	handler  PacketHandler
}

func NewUDPTransport(netProto string, addr string, handler PacketHandler) (*UDPTransport, error) {
	udpAddr, err := net.ResolveUDPAddr(netProto, addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP(netProto, udpAddr)
	if err != nil {
		return nil, err
	}

	return &UDPTransport{
		conn:     conn,
		stopChan: make(chan struct{}),
		handler:  handler,
	}, nil
}

func (t *UDPTransport) SendPacket(data []byte, addr *net.UDPAddr) error {
	_, err := t.conn.WriteTo(data, addr)
	return err
}

func (t *UDPTransport) HandlePacket(data []byte, addr *net.UDPAddr) {
	t.handler(data, addr)
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

		t.HandlePacket(buf[:read], senderAddr)
	}
}

func (t *UDPTransport) Close() error {
	err := t.conn.Close()
	<-t.stopChan
	return err
}
