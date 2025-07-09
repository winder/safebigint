package main

import (
	"github.com/winder/safebigint"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(safebigint.Analyzer)
}
