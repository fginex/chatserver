package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultListenPort = 8085
)

func main() {

	listenPort := defaultListenPort

	if len(os.Args) > 1 {

		//server listening port
		p, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		listenPort = p
	}

	var chatServer ChatServer

	chatServer.start(listenPort)

}
