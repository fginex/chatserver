package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const (
	msgDelimiter     = '\n'
	outputLineTerm   = "\r\n"
	cmdSpecifier     = '/'
	nickMsgDelimiter = ": "
	minNickLen       = 4
)

// ChatServer type
type ChatServer struct {
	port                   int
	allUsers               UserSyncMap //use sync.Map for thread-safety
	registeredNicks        RegMap
	incomingMessageChannel chan MsgInfo
}

func (server *ChatServer) start(port int) {
	server.port = port
	server.registeredNicks = make(RegMap)
	server.incomingMessageChannel = make(chan MsgInfo)

	log.Println("Server Started. Listening on port:", port)
	ctr := 0

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("Error:", err)
		os.Exit(1)
	}

	//only one of these threads per app instance - handles dispatching of incoming messages to all clients
	go server.outputHandler()

	//loop - runs on main thread and waits for incoming client connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		ctr++
		log.Println("Connection Established")

		//one of these threads to per client to read input
		go server.userInputHandler(server.allUsers.createNewUser(ctr, conn))
	}
}

func (server *ChatServer) outputHandler() {

	//process all incoming messages coming thru the message channel
	//execute incoming commands or dispatch messages to other clients in the chat

	for {
		select {
		case m := <-server.incomingMessageChannel:
			if len(m.msg) > 0 && m.msg[0] == cmdSpecifier {
				go server.processCommand(m)
			} else {
				server.allUsers.dispatchUserMessage(m)
			}
		}
	}
}

func (server *ChatServer) userInputHandler(user *UserInfo) {
	//handles reading from a client as well as cleanup once the client disconnects

	defer func() {
		user.conn.Close()

		_, ok := server.allUsers.Load(user.uid)
		if ok {
			server.allUsers.Delete(user.uid)
		}

		log.Println(user.ToString(), "disconnected.")
	}()

	user.swriteln(server.allUsers.onlineUserList())
	user.swriteln("Welcome! Your nickname is: " + user.nickname)

	//cast to net.TCPConn using type assertion
	tcpconn, ok := user.conn.(*net.TCPConn)
	if ok {
		//set tcp connection attributes
		tcpconn.SetNoDelay(false) //send accumulated data - better for this application
	}

	bufferedReader := bufio.NewReader(user.conn)
	for {
		message, err := bufferedReader.ReadBytes(msgDelimiter)
		if err == io.EOF {
			break
		}

		//strip trailing white spaces related to line termination
		message = bytes.TrimRight(message, outputLineTerm)

		msglen := len(message)
		if msglen > 0 {
			server.incomingMessageChannel <- MsgInfo{uid: user.uid, msg: message}
		}
	}
}

func (server *ChatServer) processCommand(m MsgInfo) {

	//TODO: if i had more time i would make this better by breaking out each command into
	// a seperate function then using interfaces and a map to run a given command
	// dynamically by pulling from the map and executing the corresponding func then
	// returning the result. This would eliminate having to use a switch.

	if m.msg[0] != cmdSpecifier || len(m.msg) <= 1 {
		return
	}
	smsg := string(m.msg[1:])

	//make sure the user is online
	userI, isOnline := server.allUsers.Load(m.uid)
	user := userI.(*UserInfo)

	if !isOnline {
		log.Println("uid:", user.uid, "Attempted to execute command [", smsg, "] but is not online")
		return
	}

	switch v := strings.Split(smsg, " "); v[0] {
	case "nick":
		if len(v) != 2 {
			log.Println("uid:", user.uid, "Invalid command [", smsg, "]")
			return
		}
		newNick := v[1]

		//do some validation of the nickname
		if !isValidNickname(newNick) {
			user.swriteln("Invalid Nickname.")
			log.Println("uid:", user.uid, "nickname not changed. invalid length.")
			return
		}

		if server.registeredNicks.isAlreadyRegistered(newNick) {
			user.swriteln("Nickname not changed because it is reserved.")
			log.Println("uid:", user.uid, "nickname not changed. already registered.")
			return
		}

		user.updateNickname(newNick)
		log.Println("uid:", user.uid, "changed nickname to", user.nickname)

	case "register":
		if len(v) != 3 {
			log.Println("uid:", user.uid, "Invalid command [", smsg, "]")
			return
		}
		newNick := v[1]

		if !isValidNickname(newNick) {
			user.swriteln("Invalid Nickname.")
			log.Println("uid:", user.uid, "nickname not changed. invalid length.")
			return
		}

		//check for existing registration
		if reg, regexists := server.registeredNicks[newNick]; regexists {

			if reg.checkPassword(v[2]) {
				//ok to assign the nickname to this user
				user.updateNickname(newNick)
				log.Println("uid:", user.uid, "changed nickname to", user.nickname)
			} else {
				log.Println("uid:", user.uid, "nickname not changed. invalid password.")
			}
		} else {
			//add as new registration
			if reg, regsuccess := server.registeredNicks.createNewReg(newNick, v[2]); regsuccess {

				//disconnect any existing users with this nickname now that it is registered
				server.allUsers.disconnectUsersWithNickname(newNick)

				//ok to assign the nickname to this user
				user.updateNickname(reg.nickname)
				log.Println("uid:", user.uid, "changed nickname to", user.nickname)

				//TODO: persist the registered nicks if we want it to be reloaded after a server restart
				//...

			} else {
				log.Println("uid:", user.uid, "nickname not registered.")
			}
		}

	default:
		user.swriteln("Invalid Command.")
		log.Println("uid:", user.uid, "Invalid command [", smsg, "]")
	}
}
