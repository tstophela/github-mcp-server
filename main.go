// Package main is the entry point for the GitHub MCP Server.
// It initializes and starts the Model Context Protocol server that provides
// tools for interacting with the GitHub API.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/github/github-mcp-server/pkg/server"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
	// Commit is set at build time via ldflags
	Commit = "unknown"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		transport string
		port      int
		logLevel  string
	)

	cmd := &cobra.Command{
		Use:   "github-mcp-server",
		Short: "GitHub MCP Server - Model Context Protocol server for GitHub",
		Long: `A Model Context Protocol (MCP) server that provides tools and resources
for interacting with the GitHub API. Supports repository management,
issue tracking, pull requests, and more.`,
		Version: fmt.Sprintf("%s (%s)", Version, Commit),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), transport, port, logLevel)
		},
	}

	cmd.Flags().StringVarP(&transport, "transport", "t", "stdio",
		"Transport type: stdio or http")
	// Changed default port from 8080 to 9090 to avoid conflicts with other local services
	cmd.Flags().IntVarP(&port, "port", "p", 9090,
		"Port to listen on (only used with http transport)")
	// Changed default log level from "debug" to "info" to reduce noise in normal usage
	cmd.Flags().StringVarP(&logLevel, "log-level", "l", "info",
		"Log level: debug, info, warn, error")

	return cmd
}

func runServer(ctx context.Context, transport string, port int, logLevel string) error {
	// Handle OS signals for graceful shutdown
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Read GitHub token from environment.
	// Also check GITHUB_PERSONAL_ACCESS_TOKEN as some tools set it under that name.
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		token = os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	}

	// Warn early if no token is found, since most API calls will fail without one.
	if token == "" {
		fmt.Fprintf(os.Stderr, "Warning: no GitHub token found in environment (GITHUB_TOKEN, GH_TOKEN, or GITHUB_PERSONAL_ACCESS_TOKEN)\n")
	}

	cfg := server.Config{
		Transport: transport,
		Port:      port,
		LogLevel:  logLevel,
		Token:     token,
		Version:   Version,
	}

	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	fmt.Fprintf(os.Stderr, "GitHub MCP Server %s starting (transport: %s)\n", Version, transport)

	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
