// go-session/examples/fwd
// MIT License Copyright(c) 2019, 2020 Hiroshi Shimamoto
// vim:set sw=4 sts=4:

package main

import (
    "log"
    "net"
    "os"
    "time"

    "github.com/hshimamoto/go-session"
    "github.com/hshimamoto/go-iorelay"
)

func main() {
    if len(os.Args) < 3 {
	log.Println("fwd listen dest")
	return
    }
    s, err := session.NewServer(os.Args[1], func(conn net.Conn) {
	defer conn.Close()
	fconn, err := session.Dial(os.Args[2])
	if err != nil {
	    log.Println("Dial %s %v\n", os.Args[2], err)
	    return
	}
	defer fconn.Close()
	iorelay.Relay(conn, fconn)
    })
    if err != nil {
	log.Printf("session.NewServer: %v\n", err)
	return
    }
    go s.Run()
    time.Sleep(time.Minute)
    s.Stop()
}
