# :robot: CIHunter

[![Build](https://github.com/ZJU-SEC/CIHunter/actions/workflows/build.yml/badge.svg)](https://github.com/ZJU-SEC/CIHunter/actions/workflows/build.yml)

## :gear: Prerequisite

- Docker
- Golang
- PostgreSQL

Prepare yourself a `config.ini` configuration according to `config.ini.tmpl`.

### :bulb: Dockerized PostgreSQL

To run a dockerized PostgreSQL, check [this](https://hub.docker.com/_/postgres).

Start a postgres container:

```bash
$ docker run \
  --name postgres -d \
  --restart unless-stopped \
  -e POSTGRES_USER=ZJU-SEC \
  -e POSTGRES_PASSWORD=<YOUR DB PASSWORD> \
  -e POSTGRES_DB=CIHunter \
  -p 5432:5432 postgres
```


## :hammer_and_wrench: Build

```bash
$ go build CIHunter
```

## :rocket: Run


```bash
$ ./CIHunter
```
