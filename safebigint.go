package safebigint

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type Config struct {
	EnableTruncationCheck bool `mapstructure:"enable-truncation-check"`
	EnableMutationCheck   bool `mapstructure:"enable-mutation-check"`
}

var config Config

func init() {
	Analyzer.Flags.BoolVar(&config.EnableTruncationCheck, "enable-truncation-check", true, "Enable checks for truncating conversions")
	Analyzer.Flags.BoolVar(&config.EnableMutationCheck, "enable-mutation-check", true, "Enable checks for shared object mutation")
}

var Analyzer = &analysis.Analyzer{
	Name: "safebigint",
	Doc:  "warns when Uint64() is called on a *big.Int, which may truncate silently",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: run,
}

// truncatingMethods are methods to flag for the silent truncate or overflow warning.
var truncatingMethods = map[string]struct{}{
	"Uint64": {},
	"Int64":  {},
}

// mutatingMethods are methods to flag for the mutating methods warning.
var mutatingMethods = map[string]struct{}{
	"Add": {}, "Sub": {}, "Mul": {}, "Div": {}, "Mod": {}, "Rem": {},
	"And": {}, "Or": {}, "Xor": {}, "Lsh": {}, "Rsh": {}, "Exp": {}, "Quo": {},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	checks := []func(*analysis.Pass, *ast.SelectorExpr, *ast.CallExpr){}
	if config.EnableTruncationCheck {
		checks = append(checks, checkForTruncation)
	}
	if config.EnableMutationCheck {
		checks = append(checks, checkForMutation)
	}

	inspector.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		for _, check := range checks {
			check(pass, sel, call)
		}
	})

	return nil, nil
}

// getBigIntReceiver returns the types.Object of the receiver if it's a *big.Int method call.
func getBigIntReceiver(pass *analysis.Pass, sel *ast.SelectorExpr) (types.Object, bool) {
	t := pass.TypesInfo.TypeOf(sel.X)
	ptrType, ok := t.(*types.Pointer)
	if !ok {
		return nil, false
	}

	named, ok := ptrType.Elem().(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return nil, false
	}

	if named.Obj().Pkg().Path() != "math/big" || named.Obj().Name() != "Int" {
		return nil, false
	}

	obj := getReferencedObject(pass, sel.X)
	if obj == nil {
		return nil, false
	}

	return obj, true
}

// checkForTruncation looks for unsafe calls that may result in truncation.
func checkForTruncation(pass *analysis.Pass, sel *ast.SelectorExpr, call *ast.CallExpr) {

	_, exists := truncatingMethods[sel.Sel.Name]
	if !exists {
		return
	}

	if _, ok := getBigIntReceiver(pass, sel); ok {
		pass.Reportf(call.Pos(), "calling %s on *big.Int may silently truncate or overflow", sel.Sel.Name)
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
		"Add": true, "Sub": true, "Mul": true, "Div": true, "Mod": true, "Rem": true,
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
