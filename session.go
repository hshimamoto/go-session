// go-session
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:

package session

import (
	"net"
	"os"
	"strings"
)

type Server struct {
    listener net.Listener
    handler func(conn net.Conn)
}

func NewServer(addr string, handler func(conn net.Conn)) (*Server, error) {
    l, err := Listen(addr)
    if err != nil {
	return nil, err
    }
    return &Server {
	listener: l,
	handler: handler,
    }, nil
}

func GetProtoAddr(addr string) (string, string) {
    // unix:path or IP:port
    a := strings.SplitN(addr, ":", 2)
    proto := "tcp"
    if len(a) == 1 {
	proto = "unix"
    } else if a[0] == "unix" {
	proto = "unix"
	addr = a[1]
    }
    return proto, addr
}

func Listen(addr string) (net.Listener, error) {
    proto, addr := GetProtoAddr(addr)
    if proto == "unix" {
	if err := os.Remove(addr); err != nil {
	    return nil, err
	}
    }
    return net.Listen(proto, addr)
}

func Dial(addr string) (net.Conn, error) {
    proto, addr := GetProtoAddr(addr)
    return net.Dial(proto, addr)
}

func Corkscrew(proxy, addr string) (net.Conn, error) {
    conn, err := Dial(proxy)
    if err != nil {
	return nil, err
    }
    conn.Write([]byte("CONNECT " + addr + " HTTP/1.1\r\n\r\n"))
    buf := make([]byte, 256)
    conn.Read(buf) // discard HTTP/1.1 200 Established
    return conn, nil
}

func (s *Server)Run() {
    for {
	conn, _ := s.listener.Accept()
	go s.handler(conn)
    }
}
