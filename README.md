# Opinionated Restic Container

<!--renovate image=ghcr.io/restic/restic -->
[![Version](https://img.shields.io/badge/Version-v0.18.1-informational?style=flat)](https://github.com/gabe565/docker-restic/pkgs/container/restic)
[![Build](https://github.com/gabe565/docker-restic/actions/workflows/build.yaml/badge.svg)](https://github.com/gabe565/docker-restic/actions/workflows/build.yaml)

This repo builds container images for [Restic](https://github.com/restic/restic) with [Runitor](https://github.com/bdd/runitor) and clients for Postgres, MariaDB, and MongoDB.

Release tags are automatically updated by Renovate bot, so new releases will be available in this repository within a few hours.

## Images

- [ghcr.io/gabe565/restic](https://github.com/gabe565/docker-restic/pkgs/container/restic)

## Database tools

The image ships two companion binaries. Both resolve connection details from the
same chain — an explicit flag wins, then `DB_*` environment variables, then files
in a mounted Secret directory — so a dump can be piped straight back in.

`dumpdb` writes a dump to stdout:

```shell
dumpdb cnpg
dumpdb mariadb -- --skip-ssl
```

`restoredb` reads one from stdin and is the inverse of each `dumpdb` subcommand:

```shell
restic dump latest "$RESTIC_HOST.dmp" | restoredb cnpg
restic dump latest "$RESTIC_HOST.sql" | restoredb mariadb
```

| Subcommand | Dumped with | Restored with | Default `--secret-mount` |
| --- | --- | --- | --- |
| `postgres` (alias `cnpg`) | `pg_dump --format=custom` | `pg_restore --single-transaction` | `/postgresql-app` |
| `mariadb` | `mariadb-dump` | `mariadb` | `/mariadb` |
| `mongodb` | `mongodump --archive` | `mongorestore --archive --drop` | `/mongodb` |
| `sqlite` | `sqlite3 .dump` | `sqlite3` | n/a |

`restoredb` is destructive: every engine drops or replaces existing objects. Use
`--dry-run` to print the command without running it. `restoredb sqlite` refuses to
overwrite a non-empty database file unless `--force` is passed.
