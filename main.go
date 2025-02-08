package main

import (
	"broadcast/internal"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	arg, err := getArg(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	switch arg {
	case "client":
		if err := runClient(sigCh); err != nil {
			fmt.Println("failed to start the client:", err)
			return
		}
	case "server":
		if err := runServer(sigCh); err != nil {
			fmt.Println("failed to start the server:", err)
			return
		}
	}
}

func getArg(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("not enoough arg: need either [server|client]")
	}

	if args[0] != "server" && args[0] != "client" {
		return "", fmt.Errorf("invalid arg: %s", args[0])
	}

	return args[0], nil
}

func runClient(sigCh chan os.Signal) error {
	client, err := internal.InitTcpClient()
	if err != nil {
		return err
	}

	go client.Run()

	select {
	case <-client.ConnStatus:
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client.Stop(ctx)
	}

	return nil
}

func runServer(sigCh chan os.Signal) error {
	server, err := internal.InitTcpServer()
	if err != nil {
		return err
	}

	go server.Run()

	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Stop(ctx)

	return nil
}
