package main

/*

#cgo CFLAGS: -I../nftp-codec/src
#cgo LDFLAGS: -L../nftp-codec/build -lnftp-codec -lhashtable -Wl,-rpath=../nftp-codec/build

#include "nftp.h"
#include <stdlib.h>

int
size2int(size_t sz)
{
	return (int) sz;
}

int
nftp_proto_handler2(char *msg, size_t len, uint8_t **rmsg, size_t *rlen)
{
	return nftp_proto_handler(msg, len, rmsg, rlen);
}

int
nftp_msg_type(char *msg)
{
	return (int) msg[0];
}

*/
import "C"

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func main() {
	smoketest()

	port := ":9999"
	listener, e := net.Listen("tcp", port)

	if e != nil {
		fmt.Println(e)
		return
	}

	defer listener.Close()

	conn, e := listener.Accept()
	if e != nil {
		fmt.Println(e)
		return
	}

	for {
		msg, e := bufio.NewReader(conn).ReadString('\n')
		if e != nil {
			fmt.Println(e)
			return
		}

		if strings.TrimSpace(msg) == "STOP" {
			fmt.Println("Exiting TCP server!")
			return
		}

		fmt.Print("-> ", msg)
		t := time.Now()
		myTime := t.Format(time.RFC3339) + "\n"
		conn.Write([]byte(myTime))
	}
}

func smoketest() {
	fmt.Println("-------------------------------")
	// C Library
	C.test()
	fmt.Println("-------------------------------")
}

func charToBytes(src *C.uchar, sz C.ulong) []byte {
	size := C.size2int(sz)
	s := make([]int, 1)
	s[0] = int(size)
	fmt.Println(s)
	return C.GoBytes(unsafe.Pointer(src), size)
}
