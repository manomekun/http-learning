package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Error(err.Error())
		return
	}
	quit := make(chan bool)
	connCh := make(chan net.Conn)

	go func() {
		for {
			c, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					logger.Info("Listener Closed")
				} else {
					logger.Error(err.Error())
				}
				return
			}
			connCh <- c
		}
	}()

	for {
		select {
		case conn := <-connCh:
			go handleConnection(conn, logger, quit)
		case <-quit:
			if err := listener.Close(); err != nil {
				logger.Error(err.Error())
			}
			return
		}
	}
}

func handleConnection(conn net.Conn, logger *slog.Logger, quit chan<- bool) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	fmt.Println(string(buf))

	if strings.HasPrefix(strings.TrimSpace(string(buf)), "quit") {
		quit <- true
		return
	}

	data := buf[:n]
	if _, err := conn.Write(data); err != nil {
		logger.Error(err.Error())
		return
	}
}
