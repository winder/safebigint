package safebigint_test

import (
	"testing"

	"github.com/winder/safebigint"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, safebigint.Analyzer, "testpkg")
}
