# :robot: CIAnalyser

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
  -e POSTGRES_DB=CIAnalyser \
  -p 5432:5432 postgres
```


## :hammer_and_wrench: Build

```bash
$ go build CIAnalyser
```

## :rocket: Run


```bash
$ ./CIAnalyser

These are common CIAnalyser commands used in various situations:

working with official action scripts:
  official-actions    get official actions from GitHub marketplace
  official-repos      use the actions got to fetch their repositories
  
download information & clone repositoreis
  dependents          get all the dependents from the action repositories via 'Insight' page
  recover             recover the failed process in the `dependents` stage
  migrate             migerate the repos found in `dependents` stage to list repo
  clone-repo          download all the repositories
  
other options provided by CIAnalyser
  crawl-script
  extract-script
  clone-script
  clone-contributor
  extract-credential
  label-usage
  parse-use
```
