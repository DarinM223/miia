package main

type Expr interface {
	// expr is a dummy method that "registers"
	// a struct to be an Expr.
	expr()
}

// An integer expression like `1`.
type IntExpr struct {
	Value int
}

// A string expression like "foo".
type StringExpr struct {
	Value string
}

// A range expression over two values like `1..2`.
type RangeExpr struct {
	Start Expr
	End   Expr
}

const (
	SelectorClass = iota // A CSS class to retrieve
	SelectorID           // A CSS id to retrieve
)

// SelectorExpr retrieves data from the current page.
type SelectorExpr struct {
	Type int
	Name string
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

// BinOp applies a binary operator to two expressions.
type BinOp struct {
	Operator Token
	A        Expr
	B        Expr
}

// UnOp applies a unary operator to an expression.
type UnOp struct {
	Operator Token
	A        Expr
}

func (e IntExpr) expr()      {}
func (e StringExpr) expr()   {}
func (e RangeExpr) expr()    {}
func (e SelectorExpr) expr() {}
func (e VarExpr) expr()      {}
func (e ForExpr) expr()      {}
func (e IfExpr) expr()       {}
func (e BinOp) expr()        {}
func (e UnOp) expr()         {}
