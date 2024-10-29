package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

func redirect(from io.Reader, to io.Writer, doneChan chan<- bool) {
	writer := bufio.NewWriter(to)
	writer.ReadFrom(from)
	doneChan <- true
}

type config struct {
	timeout time.Duration
	host    string
	port    string
}

func getConfig() (*config, error) {
	cfg := &config{}

	flag.DurationVar(&cfg.timeout, "timeout", 10*time.Second, "timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		return nil, errors.New("Not enough arguments")
	}

	cfg.host = args[0]

	if _, err := strconv.Atoi(args[1]); err != nil {
		return nil, errors.New("Invalid port value")
	}
	cfg.port = args[1]

	return cfg, nil
}

func getConnection(addr *net.TCPAddr, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", addr.String(), timeout)
	if err == nil {
		return conn, nil
	}

	timeoutChan := time.After(timeout)
	for {
		select {
		case <-timeoutChan:
			return nil, errors.New("Timeout")
		default:
			conn, err := net.DialTimeout("tcp", addr.String(), timeout)
			if err == nil {
				return conn, nil
			}
			<-time.After(time.Millisecond)
		}
	}
}

func main() {
	cfg, err := getConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	addr, err := net.ResolveTCPAddr("tcp", cfg.host+":"+cfg.port)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := getConnection(addr, cfg.timeout)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		conn.Close()
		fmt.Println("Connection closed")
	}()

	fmt.Println("Connected")

	doneChan := make(chan bool)
	go redirect(os.Stdin, conn, doneChan)
	go redirect(conn, os.Stdout, doneChan)

	<-doneChan
}
