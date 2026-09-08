// Command mcp serves the HRIS as a read-only MCP server over stdio, so an AI
// assistant (Claude Desktop, Claude Code) can answer questions about
// attendance, leave and employees.
//
// It speaks JSON-RPC on stdin/stdout, so anything written to stdout that is
// not protocol traffic corrupts the session: all logging goes to stderr.
//
//	HRIS_API_URL=http://127.0.0.1:8080 \
//	HRIS_MCP_EMAIL=hr@perusahaan.com \
//	HRIS_MCP_PASSWORD=... \
//	  ./bin/mcp
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hris-face/api/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const version = "0.1.0"

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.SetPrefix("hris-mcp: ")

	baseURL := envOr("HRIS_API_URL", "http://127.0.0.1:8080")
	email := os.Getenv("HRIS_MCP_EMAIL")
	password := os.Getenv("HRIS_MCP_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("HRIS_MCP_EMAIL dan HRIS_MCP_PASSWORD wajib diisi " +
			"(pakai akun HR atau superadmin; server ini hanya membaca data)")
	}

	client := mcpserver.NewClient(baseURL, email, password)
	server := mcpserver.New(client, version)

	// Ctrl+C and a client disconnect both have to end the process; a stdio
	// server that outlives its client is an orphan holding a login.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("siap, menghubungi %s sebagai %s (read-only)", baseURL, email)
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		log.Fatalf("berhenti: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
