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
nftp_proto_maker2(char *fpath, int type, uint8_t key, int n, uint8_t **rmsg, size_t *rlen)
{
	return nftp_proto_maker(fpath, type, key, n, rmsg, rlen);
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
		_, e := conn.Write(rmsgb)
		if e != nil {
			fmt.Println("Error in sending")
			return
		}

		fmt.Println("Waiting for ACK ->")
		amsg := ReadNftpMsg(conn)
		fmt.Println("Go on", amsg)

		blocks := int(C.nftp_file_blocks2(fpath))
		fmt.Print("Blocks:")
		fmt.Println(blocks)
		for i := 0; i < blocks; i++ {
			var fmsg *C.uchar
			var flen C.ulong

			C.nftp_proto_maker2(fpath, NFTP_TYPE_FILE, 0, C.int(i), &fmsg, &flen)
			defer C.free(unsafe.Pointer(fmsg))

			fmsgb := charToBytes(fmsg, flen)
			_, e := conn.Write(fmsgb)
			if e != nil {
				fmt.Println("Error in sending")
				return
			}

			fmt.Println(i+1, "/", blocks)
		}
		fmt.Println(_fpath, "has Done.")
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
	fmt.Println(int(size))
	return C.GoBytes(unsafe.Pointer(src), size)
}
