package main

import (
	"bufio"
	_ "bytes"
	_ "errors"
	"fmt"
	"net"
	"net/http"
	_ "net/url"
	"path"
	_ "strings"
	"sync"
	"time"
)

var cfg *Config

func getSerAddr(req *http.Request) string {

	if req.URL.Path == "/" {
		return cfg.StaticSer
	}

	switch path.Ext(req.URL.Path) {
	case "":
		return cfg.HandleSer
	default:
		return cfg.StaticSer
	}
}

func setRequest(req *http.Request) {
	if req.URL.Path == "/" {
		req.URL.Path = cfg.IndexPage
	}
}

type Router struct {
	Conn    net.Conn
	ConnMap map[string]net.Conn
	Timeout time.Duration
}

func NewRouter(c net.Conn) Router {
	return Router{c, map[string]net.Conn{}, time.Duration(1 * time.Hour)}
}

func (router Router) getConn(addr string) net.Conn {
	return router.ConnMap[addr]
}

func (router Router) setConn(addr string, c net.Conn) {
	router.ConnMap[addr] = c
}

func (router Router) delConn(addr string) {
	if c, ok := router.ConnMap[addr]; ok {
		c.Close()
		delete(router.ConnMap, addr)
	}
}

func (router Router) pipe(src net.Conn) {
	var dst = router.Conn
	var buf [4 << 10]byte
	for {
		//src.SetDeadline(time.Now().Add(router.Timeout))
		n, err := src.Read(buf[:])
		if err != nil {
			return
		}

		//dst.SetDeadline(time.Now().Add(router.Timeout))
		n, err = dst.Write(buf[:n])
		if err != nil {
			return
		}
	}
}

func (router Router) Run() (err error) {

	var wg sync.WaitGroup
	var req *http.Request

	for {
		router.Conn.SetDeadline(time.Now().Add(router.Timeout))
		if req, err = http.ReadRequest(bufio.NewReader(router.Conn)); err != nil {
			fmt.Println("read request err: " + err.Error())
			break
		}

		setRequest(req)

		addr := getSerAddr(req)

		fmt.Printf("%s %s at %s\n", req.Method, req.RequestURI, addr)

		c := router.getConn(addr)
		if c == nil {
			if c, err = net.Dial("tcp", addr); err != nil {
				fmt.Println("dial err: " + err.Error())
				fmt.Fprintf(router.Conn, "HTTP/1.1 400 dial %s err \r\n\r\n", addr)
				break
			}

			router.setConn(addr, c)

			wg.Add(1)
			go func(c net.Conn, k string) {
				defer wg.Done()
				router.pipe(c)
				router.delConn(addr)
			}(c, addr)
		}

		if err = req.Write(c); err != nil {
			fmt.Println("req write err: " + err.Error())
			break
		}
	}

	wg.Wait()
	router.close()

	return
}

func (router Router) close() {
	router.Conn.Close()
	for k, v := range router.ConnMap {
		v.Close()
		delete(router.ConnMap, k)
	}
}

func handleConnection(conn net.Conn) (err error) {

	router := NewRouter(conn)
	router.Timeout = time.Duration(30 * time.Second)

	//fmt.Println("===begin===")

	router.Run()

	//fmt.Println("===end===")

	return nil
}

func main() {
	var err error
	if cfg, err = ParseConfig("config.json"); err != nil {
		fmt.Println("load config err: " + err.Error())
		return
	}

	fmt.Println("server start, listen to: " + cfg.LnAddr)

	ln, err := net.Listen("tcp", cfg.LnAddr)
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
