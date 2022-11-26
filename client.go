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

*/
import "C"

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

	conn, e := net.Dial("tcp", port)
	if e != nil {
		fmt.Println(e)
		return
	}

	C.nftp_proto_init()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		_fpath, _ := reader.ReadString('\n')
		_fpath = _fpath[:len(_fpath)-1]

		fpath := C.CString(_fpath)
		defer C.free(unsafe.Pointer(fpath))

		fmt.Println(_fpath)

		var rmsg *C.uchar
		var rlen C.ulong

		C.nftp_proto_maker(fpath, NFTP_TYPE_HELLO, 0, 0, &rmsg, &rlen)
		defer C.free(unsafe.Pointer(rmsg))

		rmsgb := charToBytes(rmsg, rlen)
		_, e := conn.Write(append(rmsgb, byte('\n')))
		if e != nil {
			fmt.Println("Error in sending")
			return
		}
		fmt.Println("done")

		msg, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Println("-> " + msg)
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
