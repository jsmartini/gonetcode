package test

import (
	"net"
	"fmt"
	"time"
	"bufio"
	"os"
)

type client struct {
	hostname string
	port string
	conn net.Conn
	IO *bufio.ReadWriter
	err error
}

type network interface {
	handshake() error
	send(string) error
	recv() (string, error)
}

func (c client) handshake() error {
	c.conn, c.err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", c.hostname, c.port), time.Duration(50)*time.Millisecond)
	if c.err != nil {
		return fmt.Errorf("Could not Connect:\t%s", c.err)
	}
	fmt.Println("Connected to %s:%s Successfully", c.hostname, c.port)
	c.IO = bufio.NewReadWriter(bufio.NewReader(c.conn), bufio.NewWriter(c.conn))
	return nil
}

func (c client) send(msg string) error {
	n, err := c.IO.WriteString(msg+ "\n")
	if err != nil {
		return err
	}
	fmt.Println("Wrote %d bytes to target Successfully", n)
	return nil
}

func (c client) recv() (string, error) {
	var msg string
	var err error
	msg, err = c.IO.ReadString('\n')
	if err != nil {
		return fmt.Sprintf("%s",err), err
	}
	return msg, nil
}

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Usage:\t ./tcp [hostname] [port]")
		os.Exit(0)
	}
	var target client
	exec := true
	buffer := ""
	target.hostname = args[1]
	target.port 	= args[2]
	err :=	target.handshake()
	if err != nil{
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
		}else{
			fmt.Println("[%s:%s> %s", target.hostname, target.port, res)
		}
	}
}


