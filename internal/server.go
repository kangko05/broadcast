package internal

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
)

type TcpServer struct {
	listener net.Listener
	clients  map[net.Conn]bool

	mu     sync.RWMutex
	stopCh chan struct{}
}

func InitTcpServer() (*TcpServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		return nil, err
	}

	return &TcpServer{
		listener: listener,
		clients:  make(map[net.Conn]bool, 0),
		mu:       sync.RWMutex{},
		stopCh:   make(chan struct{}, 1),
	}, nil
}

func (tc *TcpServer) Run() error {
	for {
		select {
		case <-tc.stopCh:
			return nil
		default:
			conn, err := tc.listener.Accept()
			if err != nil {
				return err
			}

			tc.mu.Lock()
			_, ok := tc.clients[conn]
			if !ok {
				tc.clients[conn] = true
			}
			tc.mu.Unlock()

			go tc.handleConnection(conn)
		}
	}
}

func (tc *TcpServer) handleConnection(conn net.Conn) {
	defer tc.cleanupConnection(conn)

	for {
		select {
		case <-tc.stopCh:
			return
		default:
			// read message
			recv := make([]byte, BUFFER_SIZE)
			n, err := conn.Read(recv)
			if err != nil {
				if err == io.EOF {
					fmt.Println(conn.RemoteAddr(), "disconnected")
				} else {
					fmt.Println("reading msg:", err)
				}

				return
			}

			msg := string(recv[:n])

			if msg == "exit" {
				fmt.Println(conn.RemoteAddr(), "disconnected")
				return
			}

			tc.broadcast(conn, recv[:n])
		}
	}
}

func (tc *TcpServer) broadcast(conn net.Conn, packet []byte) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	for connection, connected := range tc.clients {
		if connected {

			name := conn.RemoteAddr().String()

			if connection == conn {
				name = "me"
			}

			msg := []byte(name + ": ")
			msg = append(msg, packet...)

			n, err := connection.Write(msg)
			if err != nil {
				fmt.Printf("write to %v failed\n", connection.RemoteAddr())
				continue
			}

			if n < len(msg) {
				fmt.Printf("write to %v failed\n", connection.RemoteAddr())
				continue
			}

			fmt.Printf("wrtie to %v sucess\n", connection.RemoteAddr())
		}
	}
}

func (tc *TcpServer) cleanupConnection(conn net.Conn) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	_, ok := tc.clients[conn]
	if !ok {
		return
	}

	delete(tc.clients, conn)

	conn.Close()
}

func (tc *TcpServer) Stop(ctx context.Context) {
	select {
	case <-ctx.Done():
		return

	default:
		tc.stopCh <- struct{}{}

		tc.mu.Lock()
		defer tc.mu.Unlock()

		for conn, connected := range tc.clients {
			if connected {
				conn.Close()
			}
		}

		tc.listener.Close()

		fmt.Println("server down")

		return
	}
}
