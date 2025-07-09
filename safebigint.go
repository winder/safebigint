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

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		selExpr, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selExpr.Sel.Name != "Uint64" {
			return
		}

		recvType := pass.TypesInfo.TypeOf(selExpr.X)
		ptrType, ok := recvType.(*types.Pointer)
		if !ok {
			return
		}

		named, ok := ptrType.Elem().(*types.Named)
		if !ok || named.Obj().Pkg() == nil {
			return
		}

		if named.Obj().Pkg().Path() == "math/big" && named.Obj().Name() == "Int" {
			pass.Reportf(call.Pos(), "calling Uint64() on *big.Int may truncate large values silently")
		}
	})

	return nil, nil
}
