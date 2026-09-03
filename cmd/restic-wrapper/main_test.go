package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr swaps os.Stderr for a pipe while fn runs, since execRestic
// writes its xtrace there directly.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	orig := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	done := make(chan string, 1)
	go func() {
		out, _ := io.ReadAll(r)
		done <- string(out)
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return <-done
}

func TestStdinFilenameExtension(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			// Regression: the extension switch was not updated when the dumpdb
			// subcommand was renamed cnpg -> postgres, so a --format=custom dump
			// was stored as .sql.
			name: "postgres gets .dmp",
			args: []string{"backup", "--stdin-from-command", "--", "dumpdb", "postgres"},
			want: "--stdin-filename=myns.dmp",
		},
		{
			name: "cnpg alias still gets .dmp",
			args: []string{"backup", "--stdin-from-command", "--", "dumpdb", "cnpg"},
			want: "--stdin-filename=myns.dmp",
		},
		{
			name: "mongodb gets .dmp",
			args: []string{"backup", "--stdin-from-command", "--", "dumpdb", "mongodb"},
			want: "--stdin-filename=myns.dmp",
		},
		{
			name: "mariadb gets .sql",
			args: []string{"backup", "--stdin-from-command", "--", "dumpdb", "mariadb"},
			want: "--stdin-filename=myns.sql",
		},
		{
			name: "sqlite names the snapshot after the database file",
			args: []string{"backup", "--stdin-from-command", "--", "dumpdb", "sqlite", "/data/app.sqlite"},
			want: "--stdin-filename=/data/app.sql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RESTIC_WRAPPER_DRY_RUN", "true")
			t.Setenv("RESTIC_HOST", "myns")
			t.Setenv("RESTIC_GROUP_BY", "")

			var err error
			got := captureStderr(t, func() { err = run(tt.args) })
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected %q in %q", tt.want, got)
			}
		})
	}
}

func TestExplicitStdinFilenameIsPreserved(t *testing.T) {
	t.Setenv("RESTIC_WRAPPER_DRY_RUN", "true")
	t.Setenv("RESTIC_HOST", "myns")
	t.Setenv("RESTIC_GROUP_BY", "")

	var err error
	got := captureStderr(t, func() {
		err = run([]string{
			"backup", "--stdin-from-command", "--stdin-filename=custom.dmp",
			"--", "dumpdb", "postgres",
		})
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if strings.Contains(got, "myns.dmp") {
		t.Errorf("wrapper overrode an explicit --stdin-filename: %q", got)
	}
}
