package main

type Expr interface {
	// expr is a dummy method that "registers"
	// a struct to be an Expr.
	expr()
}

type IntExpr struct {
	Value int
}

const (
	SelectorClass = iota
	SelectorID
)

type SelectorExpr struct {
	Type int
	Name string
}

type VarExpr struct {
	Name string
}

type ForExpr struct {
	Collection Expr
	Name       string
	Body       Stmt
}

func (e IntExpr) expr()      {}
func (e SelectorExpr) expr() {}
func (e VarExpr) expr()      {}
func (e ForExpr) expr()      {}
