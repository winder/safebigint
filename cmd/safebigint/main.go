package main

import (
	"flag"
	"log"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/winder/safebigint"
)

func main() {
	var config safebigint.LinterSettings
	flag.BoolVar(&config.DisableTruncationCheck, "disable-truncation-check", false, "Disable checks for truncating conversions")
	flag.BoolVar(&config.DisableMutationCheck, "disable-mutation-check", false, "Disable checks for shared object mutation")
	flag.Parse()

	analyzer, err := safebigint.NewAnalyzer(config)
	if err != nil {
		log.Fatalf("failed to create analyzer: %v", err)
	}

	singlechecker.Main(analyzer)
}
