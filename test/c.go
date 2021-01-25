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

func (c client) server(status chan string) {
	c.listener, c.errs = net.Listen("tcp", fmt.Sprintf("%s:%s", c.hostname, c.port))
	if c.errs != nil {
		status <- "sDone"
		return
	}
	var conn net.Conn
	conn, c.errs = c.listener.Accept()
	if c.errs != nil {
		status <- "sDone"
		return
	}
	c.IOr = bufio.NewReader(conn)
	status <- "sDone"
	return
}

func (c client) client(status chan string, log chan string) {
	var conn net.Conn
	for i := 0; i <= 500; i++ {
		conn, c.errc = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", c.hostname, c.port), time.Duration(50)*time.Millisecond)
		if c.errc != nil {
			log <- fmt.Sprintf("\rTry: %d", i)
		}
	}
	if c.errc != nil {
		fmt.Println("\nExiting:\t%S", c.errc)
		os.Exit(-1)
	}
	log <- fmt.Sprintln("Connected!")
	c.IOw = bufio.NewWriter(conn)
	status <- "cDone"
	return
}

func (c client) handshake() error {
	log, status := make(chan string), make(chan string)
	go c.client(status, log)
	go c.server(status)
	i := 0
	for i < 2 {
		select {
		case status_update := <-status:
			if status_update == "cDone" {
				i++
			} else if status_update == "sDone" {
				i++
			}
		case log_update := <-log:
			fmt.Print(log_update)
		default:
			continue
		}
	}
	if c.errs != nil || c.errc != nil {
		return fmt.Errorf("Errors:\n:\tServer:\t%s\n:\tClient:\t%s\n", c.errs, c.errc)
	}
	c.IOrw = bufio.NewReadWriter(c.IOr, c.IOw)
	return nil
}

func (c client) send(s chan error) {
	var msg string
	fmt.Printf("[?>")
	_, err := fmt.Scanln(&msg)
	if err != nil {
		s <- err
	}
	if msg == "exit" {
		os.Exit(0)
	}
	n, err := c.IOrw.WriteString(msg + "\n")
	if err != nil {
		s <- err
	}
	fmt.Println("Wrote %d bytes to target Successfully", n)
	return
}

func (c client) recv(s chan error, o chan string) {
	for {
		msg, err := c.IOrw.ReadString('\n')
		if err != nil {
			s <- err
		}
		o <- fmt.Sprintln("[%s:%s> %s", c.hostname, c.port, msg)
	}
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
	status := make(chan error)
	output := make(chan string)
	go target.recv(status, output)
	for {
		target.send(status)
		select {
		case state := <-status:
			fmt.Println(state)
		case out := <-output:
			fmt.Println(out)
		}
	}
}
