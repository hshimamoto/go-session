// go-session
// MIT License Copyright(c) 2019, 2020, 2021 Hiroshi Shimamoto
// vim:set sw=4 sts=4:

package session

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

type Server struct {
    listener net.Listener
    running bool
    handler func(conn net.Conn)
}

func NewServer(addr string, handler func(conn net.Conn)) (*Server, error) {
    l, err := Listen(addr)
    if err != nil {
	return nil, err
    }
    return &Server {
	listener: l,
	running: false,
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
	    if os.IsNotExist(err) == false {
		return nil, err
	    }
	}
    }
    return net.Listen(proto, addr)
}

func Dial(addr string) (net.Conn, error) {
    proto, addr := GetProtoAddr(addr)
    return net.Dial(proto, addr)
}

func HttpConnect(conn net.Conn, addr string) error {
    var err error
    _, err = conn.Write([]byte("CONNECT " + addr + " HTTP/1.1\r\n\r\n"))
    if err != nil {
	return err
    }
    buf := make([]byte, 256)
    n, err := conn.Read(buf) // discard HTTP/1.1 200 Established
    if err != nil {
	return err
    }
    headercheck := true
    for {
	if headercheck {
	    // 012345678901234
	    // HTTP/1.1 200 OK
	    if n > 12 {
		if buf[8] == ' ' && buf[9] == '2' && buf[12] == ' ' {
		    // header is okay
		    headercheck = false
		} else {
		    return fmt.Errorf("bad response: %s", string(buf[:12]))
		}
	    }
	}
	if bytes.Index(buf, []byte{13, 10, 13, 10}) > 0 {
	    break
	}
	if n >= 256 {
	    return fmt.Errorf("header too long")
	}
	r, err := conn.Read(buf[n:n+1])
	if err != nil {
	    return err
	}
	if r == 0 {
	    return fmt.Errorf("connection closed")
	}
	n += r
    }
    return err
}

func Corkscrew(proxy, addr string) (net.Conn, error) {
    conn, err := Dial(proxy)
    if err != nil {
	return nil, err
    }
    err = HttpConnect(conn, addr)
    if err != nil {
	conn.Close()
	return nil, err
    }
    return conn, nil
}

func (s *Server)Run() {
    if s.running {
	return
    }
    s.running = true
    for s.running {
	conn, err := s.listener.Accept()
	if err != nil {
	    continue
	}
	go s.handler(conn)
    }
}

func (s *Server)Stop() {
    if s.running {
	s.running = false
	s.listener.Close()
    }
}
