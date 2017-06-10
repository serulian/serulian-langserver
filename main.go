package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/serulian/serulian-langserver/handler"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"
)

const (
	tcpMode   = "tcp"
	stdioMode = "stdio"
)

var (
	mode                      string
	debug                     bool
	addr                      string
	entrypointSourceFile      string
	vcsDevelopmentDirectories []string
)

func main() {
	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "Runs the Serulian language server",
		Long:  `Runs the Serulian language server`,
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	cmdRun.PersistentFlags().StringVar(&addr, "addr", ":4389", "The address at which the language server will be served")
	cmdRun.PersistentFlags().StringVar(&mode, "mode", stdioMode, "The communication mode under which the language server will be served (tcp|stdio)")
	cmdRun.PersistentFlags().StringVar(&entrypointSourceFile, "entrypointSourceFile", "", "The entrypoint source file for the project (optional)")
	cmdRun.PersistentFlags().StringSliceVar(&vcsDevelopmentDirectories, "vcs-dev-dir", []string{},
		"If specified, VCS packages without specification will be first checked against this path")

	// Register the root command.
	var rootCmd = &cobra.Command{
		Use:   "langserver",
		Short: "Serulian Language Server",
		Long:  "Serulian Language Server",
	}

	rootCmd.AddCommand(cmdRun)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "If set to true, will print debug logs")
	rootCmd.Execute()
}

func run() error {
	var connOptions []jsonrpc2.ConnOpt
	if debug {
		log.SetOutput(os.Stderr)
		log.Println("Debug mode enabled")
		connOptions = append(connOptions, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
	}

	handler := handler.NewHandler(entrypointSourceFile, vcsDevelopmentDirectories)
	if mode == stdioMode {
		log.Printf("Serulian language server running under STDIO mode\n")
		<-jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}), handler, connOptions...).DisconnectNotify()
		return nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		if debug {
			log.Printf("Got error when trying to start: %v\n", err)
		}
		return err
	}
	defer listener.Close()

	log.Printf("Serulian language server running under TCP at %s\n", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), handler, connOptions...)
	}
}

// Note: Based on https://github.com/sourcegraph/go-langserver/blob/master/main.go
type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
