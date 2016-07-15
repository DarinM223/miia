package main

import (
	"github.com/DarinM223/miia/graph"
	"github.com/DarinM223/miia/tokens"
)

type Expr interface {
	// expr is a dummy method that "registers"
	// a struct to be an Expr.
	expr()
}

// IntExpr is an integer expression like `1`.
type IntExpr struct {
	Value int
}

// StringExpr is a string expression like "foo".
type StringExpr struct {
	Value string
}

// BoolExpr is a boolean expression like `true` or `false`.
type BoolExpr struct {
	Value bool
}

// SelectorExpr retrieves data from the current page.
type SelectorExpr struct {
	Selectors []graph.Selector
}

// VarExpr accesses a variable defined in the scope.
type VarExpr struct {
	Name string
}

// ForExpr maps values over a collection.
type ForExpr struct {
	Collection Expr
	Name       string
	Body       Expr
}

// IfExpr checks a predicate and evaluates different bodies based
// on whether the predicate is true or false.
type IfExpr struct {
	Pred   Expr // Predicate to evaluate if true/false.
	Conseq Expr // The body if predicate is true.
	Alt    Expr // The body if predicate is false.
}

// GotoExpr changes the page to the URL string
// created from evaluating the subexpression.
type GotoExpr struct {
	URL Expr
}

// BlockExpr creates a new scope and evaluates each of the
// subexprs and returns the result of the last subexpr.
type BlockExpr struct {
	Exprs []Expr
}

// BindExpr binds multiple name-expression pairs to the local scope.
type BindExpr struct {
	Bindings map[string]Expr
}

// MultOp applies an operator to a list of subexpressions.
type MultOp struct {
	Operator tokens.Token
	Exprs    []Expr
}

// BinOp applies a binary operator to two expressions.
type BinOp struct {
	Operator tokens.Token
	A        Expr
	B        Expr
}

// UnOp applies a unary operator to an expression.
type UnOp struct {
	Operator tokens.Token
	A        Expr
}

// CollectExpr collects a stream expression into an array value.
type CollectExpr struct {
	A Expr
}

func (e IntExpr) expr()      {}
func (e StringExpr) expr()   {}
func (e BoolExpr) expr()     {}
func (e SelectorExpr) expr() {}
func (e VarExpr) expr()      {}
func (e ForExpr) expr()      {}
func (e IfExpr) expr()       {}
func (e GotoExpr) expr()     {}
func (e BlockExpr) expr()    {}
func (e BindExpr) expr()     {}
func (e MultOp) expr()       {}
func (e BinOp) expr()        {}
func (e UnOp) expr()         {}
func (e CollectExpr) expr()  {}
