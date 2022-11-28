package main

/*

#cgo CFLAGS: -I../nftp/
#cgo LDFLAGS: -L../nftp -lnftp-codec -lhashtable -Wl,-rpath=../nftp

#include <nftp.h>
#include <stdlib.h>

int
nftp_msg_type(char *msg)
{
	return (int) msg[0];
}

*/
import "C"

import (
	"fmt"
	"net"
	"nftp-go/nftp"
	"unsafe"
)

func main() {
	nftp.Smoketest()

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
		msg, e := nftp.ReadNftpMsg(conn)
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
		rlen := C.int(len(_rmsg))
		defer C.free(unsafe.Pointer(rmsg))
		fmt.Println("-> ", rlen, "msg")

		var smsg *C.char
		var slen C.int

		C.nftp_proto_handler(rmsg, rlen, &smsg, &slen)
		defer C.free(unsafe.Pointer(smsg))

		switch tp := C.nftp_msg_type(rmsg); tp {
		case nftp.NFTP_TYPE_HELLO:
			sch <- C.GoBytes(unsafe.Pointer(smsg), slen)
			fmt.Println("Receive Hello msg and Reply ACK.")
		case nftp.NFTP_TYPE_ACK:
			fmt.Println("Receive ACK msg. Skip.")
		case nftp.NFTP_TYPE_END:
			fmt.Println("Received file.")
		default:
		}
	}
}
