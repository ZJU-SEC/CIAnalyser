# :robot: CIAnalyser

[![Build](https://github.com/ZJU-SEC/CIAnalyser/actions/workflows/build.yml/badge.svg)](https://github.com/ZJU-SEC/CIAnalyser/actions/workflows/build.yml)


> `CIAnalyser` is a tool developed for our paper: _Understanding Security Threats in Open Source Software CI/CD Scripts_. It is intended to crawl repositories with OSS CI configured and analyze the meta information.

For the latest release and the dataset, check [here](https://github.com/ZJU-SEC/CIAnalyser/releases/tag/v3).

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

```
$ ./CIAnalyser <stage-code>

These are common stage code used in various situations:

crawl data:
  index-repo            crawl repos via GitHub API
  clone-repo            Git clone the crawled repos
  clone-script          Git clone the CI scripts
  crawl-verified        crawl the verified CI scripts
  
prepare for analysis: 
  extract-script        extract the CI scripts dependency
  categorize-script     categorize CI scripts to find 
  parse-using           get runtime environment of each CI script
  label-usage           count the reference type of the script usage
  label-lag             calculate reference lag of the script usage
  extract-credential    extract credential usage in repos
  
generate analysis report:
  analyze
```
