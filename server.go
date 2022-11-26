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

	for {
		fmt.Println("@@")
		_msg, e := bufio.NewReader(conn).ReadString('\n')
		fmt.Println("@@")
		_msg = _msg[:len(_msg)-1]
		if e != nil {
			fmt.Println(e)
			return
		}

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
