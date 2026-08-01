package main

import (
	"fmt"
	"os"

	"github.com/nvnhan0810/ecomerce-api/internal/platform/config"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/envfile"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/postgres"
)

func main() {
	_ = envfile.Load(".env", "/app/.env")
	cfg := config.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	var err error
	switch cmd {
	case "up":
		err = postgres.Up(cfg.DatabaseURL)
	case "down":
		err = postgres.Down(cfg.DatabaseURL)
	case "refresh":
		err = postgres.Refresh(cfg.DatabaseURL)
	case "status":
		err = postgres.Status(cfg.DatabaseURL)
	case "version":
		var v int64
		v, err = postgres.Version(cfg.DatabaseURL)
		if err == nil {
			fmt.Printf("version: %d\n", v)
		}
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate %s: %v\n", cmd, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: migrate <command>

Commands:
  up        Apply all pending Up migrations
  down      Roll back the latest migration (Down)
  refresh   Down all migrations, then Up again (schema reset)
  status    Show migration status
  version   Print current DB version

Env: load from .env / DB_* / DATABASE_URL (same as API server)
`)
}
