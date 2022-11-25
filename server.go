package main

/*

#cgo CFLAGS: -I../nftp-codec/src
#cgo LDFLAGS: -L../nftp-codec/build -lnftp-codec -lhashtable -Wl,-rpath=../nftp-codec/build

#include "nftp.h"
#include <stdlib.h>

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
