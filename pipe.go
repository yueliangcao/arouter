package main

import (
	"net"
)

func PipeThenClose(dst, src net.Conn) {
	defer func() {
		src.Close()
		dst.Close()
	}()

	var buf [4 << 10]byte
	for {
		n, err := src.Read(buf[:])
		if err != nil {
			//fmt.Printf("pipe src %s read err: %s \n", src.RemoteAddr(), err.Error())
			return
		}

		n, err = dst.Write(buf[:n])
		if err != nil {
			//fmt.Printf("pipe dst %s write err: %s \n", dst.RemoteAddr(), err.Error())
			return
		}
	}
}
