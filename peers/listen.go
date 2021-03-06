package peers

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/BjornGudmundsson/p2pBackup/files"
)

var fileLock *sync.Mutex

const localhost = "127.0.0.1"

//ListenUDP sets up a UDP server
//on the given port.
func ListenUDP(port string) {
	serverAddr, e := net.ResolveUDPAddr("udp", port)
	if e != nil {
		panic(e)
	}
	conn, e := net.ListenUDP("udp", serverAddr)
	if e != nil {
		fmt.Println(conn)
		fmt.Println(e.Error())
	}
	defer conn.Close()
	buf := make([]byte, 4096)
	for {
		n, addr, e := conn.ReadFromUDP(buf)
		fmt.Println("Got packet from: ", addr.String())
		fmt.Println("Message: ", string(buf[:n]))
		if e != nil {
			fmt.Println(e.Error())
		}
	}
}

//ListenTCP starts a new tcp server on the given port
func ListenTCP(port string, backupFile string) {
	l, e := net.Listen("tcp4", port)
	if e != nil {
		panic(e)
	}
	defer l.Close()
	handler := createHandler(backupFile)
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		} else {

			go handler(c)
		}
	}
}

//verifyData takes in bytes and verifies that
//this data was sent by a known peer
func verifyData(data []byte) bool {
	//TODO: Write the function in such a way that it compares the data against all public keys.
	return true
}

//BackupData takes in a file
//descriptor structure and
//data and appends it to the file
func BackupData(f file, data []byte) error {
	verified := verifyData(data)
	if !verified {
		return NotVerifiedError()
	}
	e := files.AppendToFile(files.File(f), data)
	if e != nil {
		return e
	}
	return nil
}

func createHandler(fileName string) func(net.Conn) {
	f := func(c net.Conn) {
		fd, e := files.GetFile(fileName)
		if e != nil {
			panic(e)
		}
		fl := file(*fd)
		reader := bufio.NewReader(c)
		s, e := reader.ReadString(';')
		if e == io.EOF {
			fmt.Println("Could not read the data from the buffer")
		} else {
			if e == io.ErrUnexpectedEOF {
				err := BackupData(fl, []byte(s))
				if err != nil {
					fmt.Println(err)
				}
			} else {
				e = BackupData(fl, []byte(s))
				if e != nil {
					fmt.Println(e.Error())
				}
			}
			_, e = c.Write([]byte("Message received \n"))
			if e != nil {
				fmt.Println(e.Error())
			}
			e = c.Close()
			if e != nil {
				fmt.Println(e.Error())
			}
		}
	}
	return f
}

//SendTCPData takes in a slice of bytes
//and sends it the given peer.
func SendTCPData(d []byte, p *Peer) error {
	conn, e := net.Dial("tcp", p.Addr.String()+":"+strconv.Itoa(p.Port))
	if e != nil {
		return e
	}
	fmt.Fprintf(conn, string(d))
	//message, e := bufio.NewReader(conn).ReadString('\n')
	//if e != nil {
	//	return e
	//}
	//fmt.Println("Received: ", message)
	e = conn.Close()
	return e
}
