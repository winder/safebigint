package safebigint_test

import (
	"testing"

	"github.com/winder/safebigint"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestTruncation(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, safebigint.Analyzer, "truncation_check")
}

func TestMutation(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, safebigint.Analyzer, "mutation_check")
}
