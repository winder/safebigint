package safebigint_test

import (
	"testing"

	"github.com/winder/safebigint"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func mustAnalyzer() *analysis.Analyzer {
	a, err := safebigint.NewAnalyzer(safebigint.LinterSettings{
		DisableTruncationCheck: false,
		DisableMutationCheck:   false,
	})
	if err != nil {
		panic(err)
	}
	return a
}

func TestTruncation(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, mustAnalyzer(), "truncation_check")
}

func TestMutation(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, mustAnalyzer(), "mutation_check")
}

func TestHelpers(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, mustAnalyzer(), "helpers")
}
