package safebigint

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "safebigint",
	Doc:  "warns when Uint64() is called on a *big.Int, which may truncate silently",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspector.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		checkForTruncation(pass, sel, call)
	})

	return nil, nil
}

// checkForTruncation looks for unsafe calls that may result in truncation.
func checkForTruncation(pass *analysis.Pass, sel *ast.SelectorExpr, call *ast.CallExpr) {
	// methods to flag for the silent truncate or overflow warning.
	truncatingMap := map[string]string{
		"Uint64": "Uint64()",
		"Int64":  "Int64()",
	}

	var truncMsg string
	var exists bool
	if truncMsg, exists = truncatingMap[sel.Sel.Name]; !exists {
		return
	}

	recv := pass.TypesInfo.TypeOf(sel.X)
	ptrType, ok := recv.(*types.Pointer)
	if !ok {
		return
	}

	named, ok := ptrType.Elem().(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return
	}

	if named.Obj().Pkg().Path() == "math/big" && named.Obj().Name() == "Int" {
		pass.Reportf(call.Pos(), "calling %s on *big.Int may silently truncate or overflow", truncMsg)
	}
}
