package internal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
)

type TcpClient struct {
	conn net.Conn

	ctx    context.Context
	cancel context.CancelFunc
	msgCh  chan string

	ConnStatus chan struct{}
}

func InitTcpClient() (*TcpClient, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TcpClient{
		conn:       conn,
		ctx:        ctx,
		cancel:     cancel,
		msgCh:      make(chan string),
		ConnStatus: make(chan struct{}),
	}, nil
}

func (tc *TcpClient) Run() {
	// retreive input
	go tc.retreiveInput()
	// sending msg
	go tc.send()
	// receiving msg
	go tc.recv()
}

func (tc *TcpClient) retreiveInput() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-tc.ctx.Done():
			return
		default:
			if scanner.Scan() {
				msg := scanner.Text()

				if len(msg) > 0 {
					tc.msgCh <- msg
					fmt.Print("\033[1A\033[2K> ")
				}
			}
		}
	}
}

func (tc *TcpClient) send() {
	for {
		select {
		case <-tc.ctx.Done():
			return
		case msg := <-tc.msgCh:
			_, err := tc.conn.Write([]byte(msg))
			if err != nil {
				fmt.Printf("failed to send msg: %v\n", err)
				continue
			}
		}
	}
}

func (tc *TcpClient) recv() {
	for {
		select {
		case <-tc.ctx.Done():
			return
		default:
			recv := make([]byte, BUFFER_SIZE)
			n, err := tc.conn.Read(recv)
			if err != nil {
				if err == io.EOF {
					fmt.Println("disconnected from the server")
					tc.ConnStatus <- struct{}{}
					return
				} else {
					fmt.Printf("failed to read msg from server: %v\n", err)
					continue
				}
			}

			fmt.Printf("\033[2K\r%s\n> ", string(recv[:n]))
		}
	}
}

func (tc *TcpClient) cleanupConnection() {
	tc.cancel()
	tc.conn.Close()
	close(tc.msgCh)

	fmt.Println("client down")
}

func (tc *TcpClient) Stop(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		tc.cleanupConnection()
	}
}
