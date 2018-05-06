// copy https://github.com/zyfworks/AnotherSteamCommunityFix
package netlib

import (
	"io"
	"log"
	"net"
	"time"
)

func handleConn(conn net.Conn, remote string) {
	defer conn.Close()

	remoteConn, err := net.DialTimeout("tcp", remote, 15*time.Second)
	fatalErr(err)
	defer remoteConn.Close()

	go io.Copy(conn, remoteConn)
	io.Copy(remoteConn, conn)
}

func StartServingTCPProxy(local, remote string) {
	listener, err := net.Listen("tcp4", local)
	fatalErr(err)

	for {
		conn, err := listener.Accept()
		fatalErr(err)
		go handleConn(conn, remote)
	}
}

func fatalErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
