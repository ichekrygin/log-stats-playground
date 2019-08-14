# Log-Stats-Playground
Log-Stats-Playground is a log parser application 

## Requirements
- Go `go1.12.5` (with `GO111MODULE=on`)

## Quick-Start
```bash
go get -u github.com/ichekrygin/partybox
partybox
```

## Building

```bash
git clone github.com/ichekrygin/log-stats-playground.git
cd log-stats-playground
make help
---
build                          build log-stats-playground binary
clean                          remove log-stats-playground binary
fmt                            format log-stats-playground
help                           print Makefile targets doc's
imports                        check log-stats-playground formatting or die
install                        install log-stats-playground
lint                           run linter on log-stats-playground
run                            run log-stats-playground app
simplify                       auot-fix format/import and lint issues whenever possible
test                           run unit tests
uninstall                      unistall log-stats-playground
vet                            vet log-stats-playground
```

## Assumptions
- File records are in chronological order

## Run
Using `Makefile` `run` target
```bash
make run
```

Using provided example file
```bash
log-stats-playground < sample_csv.txt
```