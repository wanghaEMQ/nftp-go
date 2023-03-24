package nftp

/*
#cgo LDFLAGS: -lnftp-codec-static

#include "nftp.h"
#include <stdlib.h>

*/
import "C"

import (
	"fmt"
	"io"
	"net"
)

const NFTP_TYPE_HELLO = 1
const NFTP_TYPE_ACK = 2
const NFTP_TYPE_FILE = 3
const NFTP_TYPE_END = 4
const NFTP_TYPE_GIVEME = 5

func Smoketest() {
	fmt.Println("-------------------------------")
	// C Library
	C.test()
	fmt.Println("-------------------------------")
}

func ReadNftpMsg(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 5)

	_, e := io.ReadFull(conn, buf)
	if e != nil {
		return buf, e
	}

	l := toInt(buf[1:])

	bufb := make([]byte, l-5)

	_, e = io.ReadFull(conn, bufb)
	if e != nil {
		return bufb, e
	}

	buf = append(buf, bufb...)

	return buf, nil
}

func toInt(bytes []byte) int {
	result := 0
	for i := 0; i < 4; i++ {
		result = result << 8
		result += int(bytes[i])

	}

	return result
}
