package main

import (
	"bufio"
	"os"

	"github.com/ichekrygin/partybox/pkg/monitor"
	log "github.com/sirupsen/logrus"
)

// TODO: parameterize alert threshold value
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	if err := monitor.Process(scanner, 100.0); err != nil {
		log.Fatal(err.Error(), "failed processing data")
	}
}
