package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const (
	SocketFile = "/run/monitroid.sock"
)

func main() {
	if err := run(); err != nil && err != context.Canceled {
		log.Fatalf("client finished with error: %s", err)
	}
}

func run() (err error) {
	c, err := net.Dial("unix", SocketFile)
	if err != nil {
		return fmt.Errorf("failed to connect to the unix socket: %w", err)
	}
	defer func() {
		if e := c.Close(); e != nil && err == nil {
			err = fmt.Errorf("failed to close the connection: %w", err)
		}
	}()

	if _, err := io.Copy(os.Stdout, c); err != nil {
		return fmt.Errorf("failed to read a response: %w", err)
	}

	return nil
}
