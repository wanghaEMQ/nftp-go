package main

/*

#cgo CFLAGS: -I../nftp/
#cgo LDFLAGS: -L../nftp -lnftp-codec-static -Wl,-rpath,../nftp

#include <nftp.h>
#include <stdlib.h>

int
nftp_msg_type(char *msg)
{
	return (int) msg[0];
}

int
nftp_file_blocks2(char *fpath)
{
	size_t sz;
	if (nftp_file_blocks(fpath, &sz) != 0)
		return 0;
	return (int) sz;
}

*/
import "C"

import (
	"bufio"
	"fmt"
	"net"
	"nftp-go/nftp"
	"os"
	"unsafe"
	"math/rand"
)

func main() {
	nftp.Smoketest()

	port := ":9999"

	conn, e := net.Dial("tcp", port)
	if e != nil {
		fmt.Println(e)
		return
	}

	sch := make(chan []byte, 8)
	ack := make(chan []byte, 8)

	go sender(sch, conn)
	go handle_giveme(ack, sch, conn)

	C.nftp_proto_init()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("/path/to/file=>> ")
		_fpath, _ := reader.ReadString('\n')
		_fpath = _fpath[:len(_fpath)-1]

		fpath := C.CString(_fpath)
		defer C.free(unsafe.Pointer(fpath))

		fmt.Println(_fpath)

		var rmsg *C.char
		var rlen C.int

		rv := C.nftp_proto_maker(fpath, nftp.NFTP_TYPE_HELLO, 0, 0, &rmsg, &rlen)
		defer C.free(unsafe.Pointer(rmsg))

		if rv != 0 {
			continue
		}

		_, e := conn.Write(C.GoBytes(unsafe.Pointer(rmsg), rlen))
		if e != nil {
			fmt.Println("Error in sending")
			return
		}

		fmt.Println("Waiting for ACK ->")
		amsg := <- ack

		fmt.Println("Go on", amsg)

		blocks := int(C.nftp_file_blocks2(fpath))
		fmt.Print("Blocks:")
		fmt.Println(blocks)
		for i := 0; i < blocks; i++ {
			var fmsg *C.char
			var flen C.int

			if i == blocks-1 {
				C.nftp_proto_maker(fpath, nftp.NFTP_TYPE_END, 0, C.int(i), &fmsg, &flen)
			} else {
				C.nftp_proto_maker(fpath, nftp.NFTP_TYPE_FILE, 0, C.int(i), &fmsg, &flen)
			}
			defer C.free(unsafe.Pointer(fmsg))

			// Simulate unstable network
			if rand.Int()%100 < 20 {
				continue
			}

			sch <- C.GoBytes(unsafe.Pointer(fmsg), flen)

			fmt.Println(i+1, "/", blocks)
		}
		fmt.Println(_fpath, "has Done.")
	}

	C.nftp_proto_fini()
}

func handle_giveme(ack chan []byte, sch chan []byte, conn net.Conn) {
	for {
		_rmsg, e := nftp.ReadNftpMsg(conn)
		if e != nil {
			fmt.Println(e)
			break
		}

		rmsg := C.CString(string(_rmsg))
		rlen := C.int(len(_rmsg))
		defer C.free(unsafe.Pointer(rmsg))

		if C.nftp_msg_type(rmsg) == nftp.NFTP_TYPE_ACK {
			ack <- _rmsg
			continue
		}
		if C.nftp_msg_type(rmsg) != nftp.NFTP_TYPE_GIVEME {
			fmt.Println("NOT GIVEME???", C.nftp_msg_type(rmsg))
			continue
		}

		// Reply FILE/END packet
		var smsg *C.char
		var slen C.int

		C.nftp_proto_handler(rmsg, rlen, &smsg, &slen)
		defer C.free(unsafe.Pointer(smsg))

		sch <- C.GoBytes(unsafe.Pointer(smsg), slen)
	}
}

func sender(sch chan []byte, conn net.Conn) {
	for {
		smsg := <- sch
		_, e := conn.Write(smsg)
		if e != nil {
			fmt.Println("Error in sending")
			return
		}
	}
}

