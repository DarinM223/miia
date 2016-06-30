package main

import (
	"errors"
	"github.com/DarinM223/http-scraper/graph"
)

type Scope struct {
	// Env is the mapping of names to Nodes in the scope.
	Env map[string]graph.Node
	// Page is the latest Goto Node in the scope.
	Page graph.Node
	// Parent is the scope's parent.
	Parent *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Env:    make(map[string]graph.Node),
		Page:   nil,
		Parent: parent,
	}
}

func (s *Scope) lookup(name string) (graph.Node, error) {
	if node, ok := s.Env[name]; ok {
		return node, nil
	}
	if s.Parent == nil {
		return nil, errors.New("Variable with name not in scope")
	}
	return s.Parent.lookup(name)
}

func (s *Scope) set(name string, value graph.Node) {
	s.Env[name] = value
}

func CompileExpr(expr Expr, scope *Scope) (graph.Node, error) {
	switch e := expr.(type) {
	case SelectorExpr:
		if scope.Page == nil {
			return nil, errors.New("Attempting to apply selector before loading a page")
		}
		return graph.NewSelectorNode(GenID(), scope.Page, e.Selectors), nil
	case VarExpr:
		return scope.lookup(e.Name)
	case ForExpr:
		collection, err := CompileExpr(e.Collection, scope)
		if err != nil {
			return nil, err
		}

		newScope := NewScope(scope)
		body, err := CompileExpr(e.Body, newScope)
		if err != nil {
			return nil, err
		}

		return graph.NewForNode(GenID(), collection, body), nil
	case IfExpr:
		pred, err := CompileExpr(e.Pred, scope)
		if err != nil {
			return nil, err
		}

		scope1, scope2 := NewScope(scope), NewScope(scope)

		conseq, err := CompileExpr(e.Conseq, scope1)
		if err != nil {
			return nil, err
		}

		alt, err := CompileExpr(e.Alt, scope2)
		if err != nil {
			return nil, err
		}

		return graph.NewIfNode(GenID(), pred, conseq, alt), nil
	case GotoExpr:
		urlNode, err := CompileExpr(e.URL, scope)
		if err != nil {
			return nil, err
		}

		gotoNode := graph.NewGotoNode(GenID(), urlNode)
		scope.Page = gotoNode
		return gotoNode, nil
	case BlockExpr:
		newScope := NewScope(scope)
		for i, expr := range e.Exprs {
			res, err := CompileExpr(expr, newScope)
			if err != nil {
				return nil, err
			}

			// Return result of last expression.
			if i == len(e.Exprs)-1 {
				return res, nil
			}
		}
	case BindExpr:
		for name, expr := range e.Bindings {
			res, err := CompileExpr(expr, scope)
			if err != nil {
				return nil, err
			}

			scope.set(name, res)
		}
		return graph.NewValueNode(GenID(), nil), nil
	case MultOp:
	case BinOp:
	case UnOp:
	}
	return nil, nil
}

var globalCounter = 0

func GenID() int {
	// TODO(DarinM223): improve this implementation
	id := globalCounter
	globalCounter++
	return id
}
