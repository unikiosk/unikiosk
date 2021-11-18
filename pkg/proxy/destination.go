package proxy

import (
	"fmt"
	"log"
	"net"

	"github.com/cybozu-go/transocks"
)

type TCPListener struct {
	net.Listener
}

type TCPConn struct {
	*net.TCPConn
	OrigAddr string // ip:port
}

func NewTCPListener(listenAddress string) (*TCPListener, error) {
	l, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, err
	}
	return &TCPListener{
		Listener: l,
	}, nil
}

func (l *TCPListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return c, err
	}

	tc, ok := c.(*net.TCPConn)
	if !ok {
		return c, fmt.Errorf("Accepted non-TCP connection - %v", c)
	}

	origAddr, err := transocks.GetOriginalDST(tc)
	if err != nil {
		return c, fmt.Errorf("GetOriginalDST failed - %s", err.Error())
	}

	log.Printf("debug: OriginalDST: %s", origAddr)
	log.Printf("debug: LocalAddr: %s", tc.LocalAddr().String())
	log.Printf("debug: RemoteAddr: %s", tc.RemoteAddr().String())

	return &TCPConn{tc, origAddr.String()}, nil
}

func ListenTCP(listenAddress string, handler func(tc *TCPConn)) {
	l, err := NewTCPListener(listenAddress)
	if err != nil {
		log.Fatalf("alert: Error listening for tcp connections - %s", err.Error())
	}

	for {
		conn, err := l.Accept() // wait here
		if err != nil {
			log.Printf("warn: Error accepting new connection - %s", err.Error())
			return
		}

		log.Printf("debug: Accepted new connection")

		go func(conn net.Conn) {
			defer func() {
				conn.Close()
			}()

			tc, _ := conn.(*TCPConn)

			handler(tc)
		}(conn)
	}
}
