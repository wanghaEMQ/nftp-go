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
	"unsafe"
)

const NFTP_TYPE_HELLO = 1
const NFTP_TYPE_ACK = 2
const NFTP_TYPE_FILE = 3
const NFTP_TYPE_END = 4
const NFTP_TYPE_GIVEME = 5

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

	// Init NFTP protocol.
	C.nftp_proto_init()

	// Set recvdir
	recvdir := C.CString("./")
	C.nftp_set_recvdir(recvdir)
	defer C.free(unsafe.Pointer(recvdir))

	for {
		fmt.Println("@@")
		_msg := ReadNftpMsg(conn)
		fmt.Println("@@")
		_msg = _msg[:len(_msg)-1]

		var smsg *C.uchar
		var slen C.ulong

		rmsg := C.CString(_msg)
		rlen := C.ulong(len(_msg))
		defer C.free(unsafe.Pointer(rmsg))

		fmt.Println(rlen)
		fmt.Println("-> ", _msg)

		C.nftp_proto_handler2(rmsg, rlen, &smsg, &slen)
		defer C.free(unsafe.Pointer(smsg))

		if C.nftp_msg_type(rmsg) == NFTP_TYPE_HELLO {
			smsgb := charToBytes(smsg, slen)
			conn.Write(append(smsgb, byte('\n')))
		} else if C.nftp_msg_type(rmsg) == NFTP_TYPE_ACK {
			fmt.Println("Receive ACK msg. Skip.")
		} else if C.nftp_msg_type(rmsg) == NFTP_TYPE_END {
			fmt.Println("Received file.")
			break
		}
	}

	C.nftp_proto_fini()
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

func ReadNftpMsg(conn net.Conn) string {
	buf := make([]byte, 5)

	fmt.Println("here")
	reader := bufio.NewReader(conn)
	_, e := reader.Read(buf)
	if e != nil {
		fmt.Println(e)
		return string("")
	}

	l := int(buf[1]<<24) + int(buf[2]<<16) + int(buf[3]<<8) + int(buf[4])
	fmt.Println(l)

	bufb := make([]byte, l-5)

	_, e = reader.Read(bufb)
	if e != nil {
		fmt.Println(e)
		return string("")
	}
	fmt.Println("here2")

	buf = append(buf, bufb...)

	return string(buf)
}
