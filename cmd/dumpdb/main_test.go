package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gabe565/docker-restic/internal/testutil"
)

func TestGoldenArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "postgres defaults",
			args: []string{"postgres", "--dry-run"},
			want: "+ pg_dump --format=custom --compress=0 --clean --if-exists --no-owner" +
				" --host= --port=5432 --username= --dbname=\n",
		},
		{
			name: "cnpg alias matches postgres",
			args: []string{"cnpg", "--dry-run", "-H", "db", "-d", "app", "-u", "postgres"},
			want: "+ pg_dump --format=custom --compress=0 --clean --if-exists --no-owner" +
				" --host=db --port=5432 --username=postgres --dbname=app\n",
		},
		{
			name: "mariadb",
			args: []string{"mariadb", "--dry-run", "-H", "db", "-d", "app", "-u", "root"},
			want: "+ mariadb-dump --add-drop-table --skip-dump-date --single-transaction" +
				" --host=db --port=3306 --user=root app\n",
		},
		{
			// The borgbase-operator renders exactly this for a mariadb source.
			name: "mariadb passthrough after separator",
			args: []string{"mariadb", "--dry-run", "-H", "db", "-d", "app", "-u", "root", "--", "--skip-ssl"},
			want: "+ mariadb-dump --add-drop-table --skip-dump-date --single-transaction" +
				" --host=db --port=3306 --user=root app --skip-ssl\n",
		},
		{
			name: "mongodb redacts the password",
			args: []string{"mongodb", "--dry-run", "-H", "h", "-d", "app", "-u", "u", "-p", "hunter2"},
			want: "+ mongodump --archive --authenticationDatabase=" +
				" --host=h --port=27017 --username=u --password=*** --db=app\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ClearEnv(t)
			got, err := testutil.RunGolden(t, New(), tt.args...)
			if err != nil {
				t.Fatalf("execute: %v", err)
			}
			if got != tt.want {
				t.Errorf("argv mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// Guards the mount-path contract the borgbase-operator depends on: it mounts the
// CNPG Secret at this fixed default and passes no DB_* env, so every connection
// detail has to come out of the mounted files.
func TestPostgresResolvesFromSecretMount(t *testing.T) {
	testutil.ClearEnv(t)

	dir := t.TempDir()
	for name, content := range map[string]string{
		"host":     "postgresql-rw",
		"port":     "5432",
		"dbname":   "app",
		"username": "app",
		"password": "hunter2\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := testutil.RunGolden(t, New(), "cnpg", "--dry-run", "--secret-mount", dir)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	want := "+ pg_dump --format=custom --compress=0 --clean --if-exists --no-owner" +
		" --host=postgresql-rw --port=5432 --username=app --dbname=app\n"
	if got != want {
		t.Errorf("argv mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestDefaultSecretMounts(t *testing.T) {
	// These paths are hard-coded by the borgbase-operator when it mounts the
	// database Secret. Changing them silently breaks every backup and restore.
	for _, tt := range []struct{ cmd, want string }{
		{"postgres", "/postgresql-app"},
		{"mariadb", "/mariadb"},
		{"mongodb", "/mongodb"},
	} {
		t.Run(tt.cmd, func(t *testing.T) {
			for _, c := range New().Commands() {
				if c.Name() != tt.cmd {
					continue
				}
				got := c.Flags().Lookup("secret-mount").DefValue
				if got != tt.want {
					t.Errorf("%s --secret-mount default = %q, want %q", tt.cmd, got, tt.want)
				}
				return
			}
			t.Fatalf("subcommand %q not found", tt.cmd)
		})
	}
}
