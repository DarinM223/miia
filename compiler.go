package main

import (
	"errors"

	"github.com/DarinM223/miia/graph"
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
		return nil, errors.New("variable with name not in scope")
	}
	return s.Parent.lookup(name)
}

func (s *Scope) set(name string, value graph.Node) {
	s.Env[name] = value
}

func CompileExpr(g *graph.Globals, expr Expr, scope *Scope) (graph.Node, error) {
	switch e := expr.(type) {
	case IntExpr:
		return graph.NewValueNode(g, g.GenID(), e.Value), nil
	case StringExpr:
		return graph.NewValueNode(g, g.GenID(), e.Value), nil
	case BoolExpr:
		return graph.NewValueNode(g, g.GenID(), e.Value), nil
	case SelectorExpr:
		if scope.Page == nil {
			return nil, errors.New("attempting to apply selector before loading a page")
		}
		return graph.NewSelectorNode(g, g.GenID(), scope.Page, e.Selectors), nil
	case VarExpr:
		return scope.lookup(e.Name)
	case ForExpr:
		collection, err := CompileExpr(g, e.Collection, scope)
		if err != nil {
			return nil, err
		}

		newScope := NewScope(scope)
		newScope.set(e.Name, graph.NewVarNode(g, g.GenID(), e.Name))
		body, err := CompileExpr(g, e.Body, newScope)
		if err != nil {
			return nil, err
		}

		return graph.NewForNode(g, g.GenID(), e.Name, collection, body), nil
	case IfExpr:
		pred, err := CompileExpr(g, e.Pred, scope)
		if err != nil {
			return nil, err
		}

		scope1, scope2 := NewScope(scope), NewScope(scope)

		conseq, err := CompileExpr(g, e.Conseq, scope1)
		if err != nil {
			return nil, err
		}

		alt, err := CompileExpr(g, e.Alt, scope2)
		if err != nil {
			return nil, err
		}

		return graph.NewIfNode(g, g.GenID(), pred, conseq, alt), nil
	case GotoExpr:
		urlNode, err := CompileExpr(g, e.URL, scope)
		if err != nil {
			return nil, err
		}

		gotoNode := graph.NewGotoNode(g, g.GenID(), urlNode)
		scope.Page = gotoNode
		return gotoNode, nil
	case BlockExpr:
		newScope := NewScope(scope)
		var lastNode graph.Node
		for i, expr := range e.Exprs {
			res, err := CompileExpr(g, expr, newScope)
			if err != nil {
				return nil, err
			}

			// Return result of last expression.
			if i == len(e.Exprs)-1 {
				lastNode = res
				break
			}
		}
		return lastNode, nil
	case BindExpr:
		for name, expr := range e.Bindings {
			res, err := CompileExpr(g, expr, scope)
			if err != nil {
				return nil, err
			}

			scope.set(name, res)
		}
		return graph.NewValueNode(g, g.GenID(), nil), nil
	case RateLimitExpr:
		g.SetRateLimit(e.URL, e.MaxTimes, e.Duration)
		return graph.NewValueNode(g, g.GenID(), nil), nil
	case MultOp:
		nodes := make([]graph.Node, len(e.Exprs))
		for i := 0; i < len(nodes); i++ {
			node, err := CompileExpr(g, e.Exprs[i], scope)
			if err != nil {
				return nil, err
			}
			nodes[i] = node
		}
		return graph.NewMultOpNode(g, g.GenID(), e.Operator, nodes), nil
	case BinOp:
		a, err := CompileExpr(g, e.A, scope)
		if err != nil {
			return nil, err
		}

		b, err := CompileExpr(g, e.B, scope)
		if err != nil {
			return nil, err
		}

		return graph.NewBinOpNode(g, g.GenID(), e.Operator, a, b), nil
	case UnOp:
		node, err := CompileExpr(g, e.A, scope)
		if err != nil {
			return nil, err
		}

		return graph.NewUnOpNode(g, g.GenID(), e.Operator, node), nil
	case CollectExpr:
		node, err := CompileExpr(g, e.A, scope)
		if err != nil {
			return nil, err
		}
		return graph.NewCollectNode(g, g.GenID(), node), nil
	default:
		return nil, errors.New("invalid expression type")
	}
}
