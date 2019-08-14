package main

import (
	"bufio"
	"os"

	"github.com/ichekrygin/log-stats-playground/pkg/monitor"
	log "github.com/sirupsen/logrus"
)

// TODO: parameterize alert threshold value
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	span, err := monitor.NewSpan(10, 120, 100.0)
	if err != nil {
		log.Fatal(err.Error(), "failed setting up monitor")
	}

	if err := monitor.Process(scanner, span); err != nil {
		log.Fatal(err.Error(), "failed processing data")
	}
}
