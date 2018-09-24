package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	goprofile "github.com/pkg/profile"

	toolkitversion "github.com/serulian/compiler/version"
	"github.com/serulian/serulian-langserver/handler"
	"github.com/serulian/serulian-langserver/version"

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
	profile                   bool
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
			if profile {
				defer goprofile.Start(goprofile.CPUProfile).Stop()
			}

			run()
		},
	}

	cmdRun.PersistentFlags().StringVar(&addr, "addr", ":4389", "The address at which the language server will be served")
	cmdRun.PersistentFlags().StringVar(&mode, "mode", stdioMode, "The communication mode under which the language server will be served (tcp|stdio)")
	cmdRun.PersistentFlags().StringVar(&entrypointSourceFile, "entrypointSourceFile", "", "The entrypoint source file for the project (optional)")
	cmdRun.PersistentFlags().StringSliceVar(&vcsDevelopmentDirectories, "vcs-dev-dir", []string{},
		"If specified, VCS packages without specification will be first checked against this path")

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Displays the version of the Serulian language server",
		Long:  `Displays the version of the Serulian language server`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Serulian Language Server\n\n")

			fmt.Printf("Toolkit Version: %s\n\n", toolkitversion.DescriptiveVersion())

			fmt.Printf("Language Server SHA: %s\n", version.GitSHA)
			fmt.Printf("Toolkit SHA: %s\n", toolkitversion.GitSHA)
		},
	}

	// Register the root command.
	var rootCmd = &cobra.Command{
		Use:   "langserver",
		Short: "Serulian Language Server",
		Long:  fmt.Sprintf("Serulian Language Server (Version %s, SHA %s)", toolkitversion.DescriptiveVersion(), version.GitSHA),
	}

	rootCmd.AddCommand(cmdRun)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "If set to true, will print debug logs")
	rootCmd.PersistentFlags().BoolVar(&profile, "profile", false, "If set to true, the language server will be profiled")
	rootCmd.Execute()
}

func run() error {
	var connOptions []jsonrpc2.ConnOpt
	if debug {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}

		log.SetOutput(os.Stderr)
		log.Println("----> Debug mode enabled <----")
		log.Printf("Serulian lauguage server binary: %s\n", ex)
		log.Printf("Toolkit Version: %s\n\n", toolkitversion.DescriptiveVersion())
		log.Printf("Language Server SHA: %s\n", version.GitSHA)
		log.Printf("Toolkit SHA: %s\n", toolkitversion.GitSHA)

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
