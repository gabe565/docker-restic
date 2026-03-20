package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	var hasGroupBy, hasStdinFromCommand, hasStdinFilename bool
	var cmd, stdinExtension string
	var sepIndex int

outer:
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") {
			key, _, _ := strings.Cut(arg, "=")
			switch {
			case key == "--group-by":
				hasGroupBy = true
			case key == "--stdin-from-command":
				hasStdinFromCommand = true
			case key == "--stdin-filename":
				hasStdinFilename = true
			case arg == "--":
				sepIndex = i
				if len(args) > i+1 {
					next := args[i+1]
					switch filepath.Base(next) {
					case "backup-cnpg.sh", "backup-mariadb.sh", "backup-sqlite.sh":
						stdinExtension = ".sql"
					case "backup-mongodb.sh":
						stdinExtension = ".dmp"
					}
				}
				break outer
			}
		} else if cmd == "" {
			cmd = arg
		}
	}

	switch cmd {
	case "backup", "forget", "snapshots":
	default:
		return execRestic(args)
	}

	finalArgs := slices.Grow(slices.Clone(args[:sepIndex]), 2)

	if !hasGroupBy {
		if groupBy := os.Getenv("RESTIC_GROUP_BY"); groupBy != "" {
			finalArgs = append(finalArgs, "--group-by="+groupBy)
		}
	}

	if hasStdinFromCommand && !hasStdinFilename {
		if host := os.Getenv("RESTIC_HOST"); host != "" {
			finalArgs = append(finalArgs, "--stdin-filename="+host+stdinExtension)
		}
	}

	finalArgs = append(finalArgs, args[sepIndex:]...)
	return execRestic(finalArgs)
}

func execRestic(args []string) error {
	path := "/usr/bin/restic"
	argv := append([]string{path}, args...)
	if err := syscall.Exec(path, argv, os.Environ()); err != nil {
		return fmt.Errorf("exec %s: %w", path, err)
	}
	return nil
}
