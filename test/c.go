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
	send(string) error
	recv() (string, error)
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
	conn, c.errc = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", c.hostname, c.port), time.Duration(120)*time.Second)
	if c.errc != nil {
		return
	}
	c.IOw = bufio.NewWriter(conn)
}

func (c client) handshake() error {
	go c.client()
	go c.server()
	if c.errs != nil || c.errc != nil {
		fmt.Println("Errors:\n:\tServer: %s\n:\tClient: %s", c.errs, c.errc)
		os.Exit(-1)
	}
	fmt.Println("Connected to %s:%s Successfully", c.hostname, c.port)
	c.IOrw = bufio.NewReadWriter(c.IOr, c.IOw)
	return nil
}

func (c client) send(msg string) error {
	n, err := c.IOrw.WriteString(msg + "\n")
	if err != nil {
		return err
	}
	fmt.Println("Wrote %d bytes to target Successfully", n)
	return nil
}

func (c client) recv() (string, error) {
	var msg string
	var err error
	msg, err = c.IOrw.ReadString('\n')
	if err != nil {
		return fmt.Sprintf("%s", err), err
	}
	return msg, nil
}

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Usage:\t ./tcp [hostname] [port]")
		os.Exit(0)
	}
	fmt.Println()
	var target client
	exec := true
	buffer := ""
	target.hostname = args[1]
	target.port = args[2]
	err := target.handshake()
	if err != nil {
		fmt.Println("Exiting:\t%s", err)
		os.Exit(-1)
	}
	for exec {
		fmt.Sscanln("[?>", &buffer)
		if buffer == "exit" {
			target.send("BYE!")
			exec = false
			continue
		}
		target.send(buffer)
		buffer = ""
		res, err := target.recv()
		if err != nil {
			continue
		} else {
			fmt.Println("[%s:%s> %s", target.hostname, target.port, res)
		}
	}
}
