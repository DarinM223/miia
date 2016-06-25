package main

import "github.com/DarinM223/http-scraper/graph"

type Scope struct {
	// Env is the mapping of names to Nodes in the scope.
	Env map[string]graph.Node
	// Page is the latest Goto Node in the scope.
	Page graph.Node
	// Parent is the scope's parent.
	Parent *Scope
}

func CompileExpr(expr Expr) ([]graph.Node, error) {
	switch expr.(type) {
	case SelectorExpr:
	case VarExpr:
	case ForExpr:
	case IfExpr:
	case GotoExpr:
	case BlockExpr:
	case BindExpr:
	case MultOp:
	case BinOp:
	case UnOp:
	}
	return nil, nil
}
