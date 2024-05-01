package main

import (
	"fmt"
	"log"
	"net"
)

type Message struct {
	from string
	payload []byte
}

type Server struct {
	listenAddr 	string
	ln        	net.Listener
	quitch	 	chan struct{}
	msgchan   	chan Message
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch: make(chan struct{}),
		msgchan: make(chan Message, 100),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln
	
	go s.acceptLoop()

	//wait for the quit channel
	<- s.quitch
	//close message channel
	close(s.msgchan)
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		fmt.Printf("new connection: %v", conn.RemoteAddr())
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error:", err)
			return
		}
		s.msgchan <- Message{
			from: conn.RemoteAddr().String(), 
			payload: buf[:n],
		}

		conn.Write([]byte("Hello, I am server!"))
	}
}

func (s *Server) Stop() {
	close(s.quitch)
}

func main() {
    server := NewServer(":4001")
	go func() {
		for msg := range server.msgchan {
			fmt.Printf("Msg From %s:\t%s\n ", msg.from, string(msg.payload))
		}
	}()

	log.Fatal(server.Start())
}