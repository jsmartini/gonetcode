package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

type client struct {
	hostname string
	port     string
	IOr      *bufio.Reader
	IOw      *bufio.Writer
	IOrw     *bufio.ReadWriter
	listener net.Listener
	errc     error
	errs     error
}

type network interface {
	handshake() error
	server()
	client()
	send(string)
	recv()
}

func (c client) server() {
	c.listener, c.errs = net.Listen("tcp", fmt.Sprintf("%s:%s", c.hostname, c.port))
	if c.errs != nil {
		return
	}
	var conn net.Conn
	conn, c.errs = c.listener.Accept()
	if c.errs != nil {
		return
	}
	c.IOr = bufio.NewReader(conn)
}
func (c client) client() {
	var conn net.Conn
	for i := 0; i <= 500; i++ {
		conn, c.errc = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", c.hostname, c.port), time.Duration(300)*time.Millisecond)
		if c.errc != nil {
			fmt.Printf("\rTry: %d", i)
		}
	}
	fmt.Println("Connected!")
	if c.errc != nil {
		return
	}
	c.IOw = bufio.NewWriter(conn)
}

func (c client) handshake() error {
	go c.client()
	go c.server()
	if c.errs != nil || c.errc != nil {
		return fmt.Errorf("Errors:\n:\tServer:\t%s\n:\tClient:\t%s\n", c.errs, c.errc)
	}
	c.IOrw = bufio.NewReadWriter(c.IOr, c.IOw)
	return nil
}

func (c client) send() {
	var msg string
	fmt.Scanln("[?>", msg)
	if msg == "exit" {
		os.Exit(0)
	}
	n, err := c.IOrw.WriteString(msg + "\n")
	if err != nil {
		return
	}
	fmt.Println("Wrote %d bytes to target Successfully", n)
	return
}

func (c client) recv() {
	msg, err := c.IOrw.ReadString('\n')
	if err != nil {
		return
	}
	fmt.Println("[%s:%s> %s", c.hostname, c.port, msg)
	return
}

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Usage:\t ./tcp [hostname] [port]")
		os.Exit(0)
	}
	var target client
	target.hostname = args[1]
	target.port = args[2]
	err := target.handshake()
	if err != nil {
		fmt.Println("Exiting:\t%s", err)
		os.Exit(-1)
	}
	for {
		go target.recv()
		go target.send()
	}
}
