# Partybox
Partybox is a log parser application 

## Requirements
- Go `go1.12.5` (with `GO111MODULE=on`)

## Quick-Start
```bash
go get -u github.com/ichekrygin/partybox
partybox
```

## Building

```bash
git clone github.com/ichekrygin/partybox.git
cd partybox
make help
---
build                          build partybox binary
clean                          remove partybox binary
fmt                            format partybox
help                           print Makefile targets doc's
imports                        check partybox formatting or die
install                        install partybox
lint                           run linter on partybox
run                            run partybox app
simplify                       auot-fix format/import and lint issues whenever possible
test                           run unit tests
uninstall                      unistall partybox
vet                            vet partybox
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
partybox < sample_csv.txt
```