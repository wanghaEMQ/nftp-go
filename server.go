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
	"fmt"
	"io"
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
	defer listener.Close()
	if e != nil {
		fmt.Println(e)
		return
	}

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

	rch := make(chan []byte, 8)
	sch := make(chan []byte, 8)

	go handle_nftp_msg(rch, sch)
	go reply(sch, conn)

	for {
		fmt.Println("@@")
		msg, e := ReadNftpMsg(conn)
		if e != nil {
			fmt.Println(e)
			break
		}
		fmt.Println("@@")

		rch <- msg
	}

	C.nftp_proto_fini()
}

func reply(sch chan []byte, conn net.Conn) {
	for {
		smsg := <-sch
		conn.Write(smsg)
	}
}

func handle_nftp_msg(rch chan []byte, sch chan []byte) {
	for {
		_rmsg := <-rch

		rmsg := C.CString(string(_rmsg))
		rlen := C.ulong(len(_rmsg))
		defer C.free(unsafe.Pointer(rmsg))
		fmt.Println("-> ", rlen, "msg")

		var smsg *C.uchar
		var slen C.ulong

		C.nftp_proto_handler2(rmsg, rlen, &smsg, &slen)
		defer C.free(unsafe.Pointer(smsg))

		switch tp := C.nftp_msg_type(rmsg); tp {
		case NFTP_TYPE_HELLO:
			sch <- charToBytes(smsg, slen)
			fmt.Println("Receive Hello msg and Reply ACK.")
		case NFTP_TYPE_ACK:
			fmt.Println("Receive ACK msg. Skip.")
		case NFTP_TYPE_END:
			fmt.Println("Received file.")
		default:
		}
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
	return C.GoBytes(unsafe.Pointer(src), size)
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
