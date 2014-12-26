package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// parseRequestLine parses "GET /foo HTTP/1.1" into its three parts.
func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

func handleConnection(conn net.Conn) (err error) {
	defer func() {
		conn.Close()
	}()

	fmt.Println("=== begin ===")

	buf := make([]byte, 4<<10)

	var n int
	if n, err = conn.Read(buf); err != nil {
		fmt.Println("frist read err: ", err.Error())
		return
	}

	var line []byte
	if i := bytes.IndexByte(buf[:n], '\n'); i >= 0 {
		line = buf[:i+1]
	}

	method, requestURI, proto, ok := parseRequestLine(string(line))
	if !ok {
		fmt.Println("parseRequestLine err")
		return errors.New("parseRequestLine err")
	}
	fmt.Printf("%s %s %s \n", method, requestURI, proto)

	url, err := url.ParseRequestURI(requestURI)
	if err != nil {
		fmt.Println("ParseRequestURI err: " + err.Error())
		return
	}

	fmt.Println("url {")
	fmt.Println("Scheme", url.Scheme)
	fmt.Println("Opaque", url.Opaque)
	fmt.Println("Host", url.Host)
	fmt.Println("Path", url.Path)
	fmt.Println("RawQuery", url.RawQuery)
	fmt.Println("Fragment", url.Scheme)
	fmt.Println("}")

	var c2 net.Conn
	c2, err = net.Dial("tcp", "localhost:8888")

	c2.Write(buf[:n])

	go PipeThenClose(conn, c2)
	PipeThenClose(c2, conn)

	fmt.Println("=== end ===")

	return err
}

func main() {
	const (
		lnAddr = "0.0.0.0:10000"
	)
	fmt.Println("server start, listen to: " + lnAddr)

	ln, err := net.Listen("tcp", lnAddr)
	if err != nil {
		fmt.Println("net listen err: " + err.Error())
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept err: " + err.Error())
			return
		}
		fmt.Printf("new accept conn: %s \n", conn.RemoteAddr())
		go handleConnection(conn)
	}

	fmt.Println("server end")
}
