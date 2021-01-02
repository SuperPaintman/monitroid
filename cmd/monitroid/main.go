package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/SuperPaintman/monitroid/gatherers"
	"github.com/SuperPaintman/monitroid/supervisor"
)

const (
	PIDFile    = "/run/monitroid.pid"
	SocketFile = "/run/monitroid.sock"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Handle signals.
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	if err := run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("daemon finished with error: %s", err)
	}
}

func run(ctx context.Context) (err error) {
	// Acquire the pid file.
	if err := acquirePidFile(); err != nil {
		return err
	}
	defer func() {
		if e := os.Remove(PIDFile); e != nil && err == nil {
			err = fmt.Errorf("failed to remove PID file: %w", e)
		}
	}()

	// Create a supervisor.
	spv := &supervisor.Supervisor{}
	spv.Register("cpu", 1*time.Second, &gatherers.CPU{})
	spv.Register("ram", 2*time.Second, &gatherers.RAM{})
	spv.Register("disk", 10*time.Second, &gatherers.Disk{})
	defer spv.Stop()

	// Create a TCP server.
	if err := os.Remove(SocketFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove the unix socket: %w", err)
	}

	l, err := net.Listen("unix", SocketFile)
	if err != nil {
		return fmt.Errorf("failed to create TCP server: %w", err)
	}
	defer func() {
		if e := l.Close(); e != nil && err == nil {
			err = fmt.Errorf("failed to close TCP server: %w", err)
		}
	}()

	if err := os.Chmod(SocketFile, 0666); err != nil {
		return fmt.Errorf("failed to change the mode of the unix socket: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := l.Accept()
			if err != nil {
				log.Printf("Failed to accept the connection: %s", err)
				continue
			}

			go handle(conn, spv)
		}
	}()

	<-ctx.Done()
	return ctx.Err()
}

func handle(conn net.Conn, spv *supervisor.Supervisor) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close the connection: %s", err)
		}
	}()

	if err := spv.DumpJSON(conn); err != nil {
		log.Printf("Failed to dump json into the connecion: %s", err)
	}
}

func acquirePidFile() error {
	// Read the PID file.
	b, err := ioutil.ReadFile(PIDFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	// Look for the PID in the process list.
	if !os.IsNotExist(err) {
		oldpid, err := strconv.Atoi(string(b))
		if err != nil {
			log.Printf("Failed to parse PID file: %s", err)
		} else {
			isRunning, err := isProcessRunning(oldpid)
			if err != nil {
				log.Printf("Failed to check old PID: %s", err)
			} else if isRunning {
				return fmt.Errorf("pid is already running: %d", oldpid)
			}
		}
	}

	// Write new PID into the file.
	pid := os.Getpid()

	if err := ioutil.WriteFile(PIDFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

func isProcessRunning(pid int) (bool, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false, fmt.Errorf("failed to find process: %w", err)
	}

	if err := p.Signal(syscall.Signal(0)); err == nil {
		return true, nil
	}

	return false, nil
}
