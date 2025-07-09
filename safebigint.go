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
		checkForMutation(pass, sel, call)
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

// isBigIntPointer reports whether t is *big.Int.
func isBigIntPointer(t types.Type) bool {
	ptrType, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptrType.Elem().(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == "math/big" && named.Obj().Name() == "Int"
}

// getReferencedObject extracts the types.Object for an identifier or selector.
func getReferencedObject(pass *analysis.Pass, expr ast.Expr) types.Object {
	switch e := expr.(type) {
	case *ast.Ident:
		return pass.TypesInfo.Uses[e]
	case *ast.SelectorExpr:
		return pass.TypesInfo.Uses[e.Sel]
	default:
		return nil
	}
}

// checkForMutation checks for unsafe patterns where the receiver is also passed
// as an argument to a mutating method, which can lead to unexpected shared-object
// mutation issues.
func checkForMutation(pass *analysis.Pass, sel *ast.SelectorExpr, call *ast.CallExpr) {
	// List of big.Int methods that mutate the receiver and take one or more input big.Ints.
	mutatingMethods := map[string]bool{
		"Add": true, "Sub": true, "Mul": true, Div": true, "Mod": true, "Rem": true,
		"And": true, "Or": true, "Xor": true, "Lsh": true, "Rsh": true,
		"Exp": true, "Quo": true,
	}

	if !mutatingMethods[sel.Sel.Name] {
		return
	}

	// Ensure receiver is *big.Int
	if !isBigIntPointer(pass.TypesInfo.TypeOf(sel.X)) {
		return
	}

	recvObj := getReferencedObject(pass, sel.X)
	if recvObj == nil {
		return
	}

	for _, arg := range call.Args {
		argObj := getReferencedObject(pass, arg)
		if argObj != nil && argObj == recvObj {
			pass.Reportf(call.Pos(),
				"shared-object mutation: calling %s with receiver also passed as argument (e.g., x.%s(x, ...)) can be unsafe",
				sel.Sel.Name, sel.Sel.Name)
			break
		}
	}
}
