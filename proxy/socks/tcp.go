package socks

import (
	"io"
	"io/ioutil"
	"net"

	adapters "github.com/renmingxu/clash/adapters/inbound"
	"github.com/renmingxu/clash/component/socks5"
	C "github.com/renmingxu/clash/constant"
	"github.com/renmingxu/clash/log"
	authStore "github.com/renmingxu/clash/proxy/auth"
	"github.com/renmingxu/clash/tunnel"
)

type SockListener struct {
	net.Listener
	address string
	closed  bool
}

func NewSocksProxy(addr string) (*SockListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	sl := &SockListener{l, addr, false}
	go func() {
		log.Infoln("SOCKS proxy listening at: %s", addr)
		for {
			c, err := l.Accept()
			if err != nil {
				if sl.closed {
					break
				}
				continue
			}
			go HandleSocks(c)
		}
	}()

	return sl, nil
}

func (l *SockListener) Close() {
	l.closed = true
	_ = l.Listener.Close()
}

func (l *SockListener) Address() string {
	return l.address
}

func HandleSocks(conn net.Conn) {
	target, command, err := socks5.ServerHandshake(conn, authStore.Authenticator())
	if err != nil {
		_ = conn.Close()
		return
	}
	if c, ok := conn.(*net.TCPConn); ok {
		c.SetKeepAlive(true)
	}
	if command == socks5.CmdUDPAssociate {
		defer conn.Close()
		_, _ = io.Copy(ioutil.Discard, conn)
		return
	}
	tunnel.Add(adapters.NewSocket(target, conn, C.SOCKS))
}
