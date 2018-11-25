package main

import (
	"bufio"
	"fmt"
	"net"
	"testing"
)

func TestChatServer(t *testing.T) {

	var server ChatServer

	port := 8085

	go server.start(port)

	//try and connect to the server
	conn, e1 := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if e1 != nil {
		t.Errorf("Unable to connect to server")
		return
	}

	reader := bufio.NewReader(conn)
	s2, e2 := reader.ReadString('\n')
	if e2 != nil || len(s2) <= 0 {
		t.Errorf("Expecting list of online users but it was not received.")
		return

	}

	//try the /nick command
	writer := bufio.NewWriter(conn)
	_, e3 := writer.WriteString("/nick frank\r\n")
	if e3 != nil {
		t.Errorf("/nick command failed %v", e3)
		return

	}

	s4, e4 := reader.ReadString('\n')
	if e4 != nil || len(s4) <= 0 {
		t.Errorf("Expecting response from /nick command but it was not received.")
		return

	}

	//try the /register command
	_, e5 := writer.WriteString("/register frank frank123\r\n")
	if e5 != nil {
		t.Errorf("/register command failed %v", e5)
		return

	}

	s6, e6 := reader.ReadString('\n')
	if e6 != nil || len(s6) <= 0 {
		t.Errorf("Expecting response from /nick command but it was not received.")
		return

	}
}
