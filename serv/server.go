package main

/*

#cgo CFLAGS: -I../nftp/
#cgo LDFLAGS: -L../nftp -lnftp-codec-static -Wl,-rpath,../nftp

#include <nftp.h>
#include <stdlib.h>
#include <string.h>

void set_file_done_cb(char *fname);

int
nftp_msg_type(char *msg)
{
	return (int) msg[0];
}

int
file_done_cb(void *fname)
{
	char * fn = fname;
	printf("file (%s) is done \n", fn);
	set_file_done_cb(fn);
	free(fn);
}

void
set_file_done_cb(char *fname)
{
	char *fn = strdup(fname);
	nftp_proto_register(fn, file_done_cb, fn);
}
*/
import "C"

import (
	"fmt"
	"net"
	"nftp-go/nftp"
	"unsafe"
	"time"
)

var fname_curr *C.char
var flen_curr  C.int

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

	// Set cb for a special file name 'test'
	luckfname := C.CString("test")
	C.set_file_done_cb(luckfname);
	defer C.free(unsafe.Pointer(luckfname))

	rch := make(chan []byte, 8)
	sch := make(chan []byte, 8)

	fname_curr = nil

	go handle_nftp_msg(rch, sch)
	go reply(sch, conn)
	go ask_nextid(sch)

	for {
		msg, e := nftp.ReadNftpMsg(conn)
		if e != nil {
			fmt.Println(e)
			break
		}

		rch <- msg
	}

	C.nftp_proto_fini()
}

func reply(sch chan []byte, conn net.Conn) {
	for {
		smsg := <-sch
		_, e := conn.Write(smsg)
		if e != nil {
			fmt.Println("Error in sending")
			return
		}
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
			C.nftp_proto_hello_get_fname(rmsg, rlen, &fname_curr, &flen_curr)
			fmt.Println("fname ", fname_curr)
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

func ask_nextid(sch chan []byte) {
	for {
		time.Sleep(1000 * time.Millisecond)
		if fname_curr == nil {
			continue
		}
		var blocks C.int
		var nextid C.int
		C.nftp_proto_recv_status(fname_curr, &blocks, &nextid)

		if nextid == C.int(0) {
			// No more giveme needed
			fname_curr = nil
			continue
		}
		fmt.Println(fname_curr, "Ask nextid", nextid)

		var fmsg *C.char
		var flen C.int
		C.nftp_proto_maker(fname_curr, nftp.NFTP_TYPE_GIVEME, 0, nextid, &fmsg, &flen)

		sch <- C.GoBytes(unsafe.Pointer(fmsg), flen)
	}
}

