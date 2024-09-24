package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type HandlerFunc func(conn net.Conn)

type Server struct {
	addr     string
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

func NewServer(addr string) *Server {
	return &Server{addr: addr, handlers: make(map[string]HandlerFunc)}
}

func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		if cerr := listener.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		var err error
		if cerr := conn.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(err)
		}
	}()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)

	if err == io.EOF {
		log.Printf("%s", buf[:n])

	}
	if err != nil {

	}
	log.Printf("%s", buf[:n])

	data := buf[:n]
	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)
	if requestLineEnd == -1 {
		log.Print("invalid request")
	}

	requestLine := string(data[:requestLineEnd])
	fmt.Printf("Request line: %s\n", requestLine)
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		log.Print("invalid request")
	}

	version := parts[2]

	if version != "HTTP/1.1" {
		log.Print("invalid request")
	}

	for path, handler := range s.handlers {
		if parts[1] == path {
			handler(conn)
		}
	}
}

func GetResponse(conn net.Conn, body string) {
	_, err := conn.Write([]byte(
		"HTTP/1.1 200 OK\r\n" +
			"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
			"Content-type: text/html\r\n" +
			"Connection: close\r\n" + "\r\n" +
			body,
	))

	if err != nil {
		log.Print(err)
	}
}
