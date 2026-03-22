package main

import (
	"io"
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
	stdinName := os.Getenv("RESTIC_HOST")
	var cmd, stdinExt string
	sepIndex := len(args)

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
					case "backup-cnpg.sh", "backup-mariadb.sh":
						stdinExt = ".sql"
					case "backup-sqlite.sh":
						stdinExt = ".sql"
						if len(args) > i+2 {
							stdinName = args[i+2]
							ext := filepath.Ext(stdinName)
							switch ext {
							case ".db", ".sqlite3", ".sqlite":
								stdinName = strings.TrimSuffix(stdinName, ext)
							}
						}
					case "backup-mongodb.sh":
						stdinExt = ".dmp"
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

	if hasStdinFromCommand && !hasStdinFilename && stdinName != "" {
		finalArgs = append(finalArgs, "--stdin-filename="+stdinName+stdinExt)
	}

	finalArgs = append(finalArgs, args[sepIndex:]...)
	return execRestic(finalArgs)
}

func execRestic(args []string) error {
	path := "/usr/bin/restic"
	argv := append([]string{path}, args...)

	printCmd(os.Stderr, argv)

	if dryRunStr := os.Getenv("RESTIC_WRAPPER_DRY_RUN"); dryRunStr != "" {
		if dryRun, _ := strconv.ParseBool(dryRunStr); dryRun {
			return nil
		}
	}

	return syscall.Exec(path, argv, os.Environ())
}

func printCmd(out io.Writer, args []string) {
	var n int
	for _, arg := range args {
		n += len(arg)
	}

	var b strings.Builder
	b.Grow(n + len(args) + 3)
	b.WriteString("+ ")

	for i, arg := range args {
		if i != 0 {
			b.WriteByte(' ')
		}
		if strings.Contains(arg, " ") {
			arg = strconv.Quote(arg)
		}
		b.WriteString(arg)
	}

	b.WriteByte('\n')
	_, _ = io.WriteString(out, b.String())
}
