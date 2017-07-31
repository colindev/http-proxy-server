package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

var (
	addr string
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.StringVar(&addr, "addr", ":8000", "tcp listen on")
}

func main() {

	flag.Parse()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go pipe(conn)
	}
}

func pipe(srcConn net.Conn) {

	defer srcConn.Close()

	log.Println("connected:", srcConn.RemoteAddr())

	r := bufio.NewReader(srcConn)
	w := bufio.NewWriter(srcConn)
	defer w.Flush()
	defer log.Println("quit:", srcConn.RemoteAddr())

	req, err := http.ReadRequest(r)
	if err != nil {
		if err != io.EOF {
			log.Println(err)
		}
		return
	}

	log.Println("target:", req.RequestURI)
	targetAddr, err := net.ResolveTCPAddr("tcp", req.RequestURI)

	if err != nil {
		w.WriteString(fmt.Sprintf("HTTP/1.1 %d Bad Request", http.StatusBadRequest))
		return
	}

	destConn, err := net.DialTCP("tcp", nil, targetAddr)

	if err != nil {
		w.WriteString(fmt.Sprintf("HTTP/1.1 %d Internal Server Error", http.StatusInternalServerError))
		return
	}

	w.WriteString("HTTP/1.1 200 connection established\r\n\r\n")
	w.Flush()
	go io.Copy(destConn, r)
	io.Copy(w, destConn)

}
