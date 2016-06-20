package main

type Stmt interface {
	// stmt is a dummy method that "registers"
	// a struct to be a Stmt.
	stmt()
}

type BindingStmt struct {
	Bindings map[string]Expr
}

type GotoStmt struct {
	URL string
}

type SeqStmt struct {
	A Stmt
	B Stmt
}

func (s BindingStmt) stmt() {}
func (s GotoStmt) stmt()    {}
func (s SeqStmt) stmt()     {}
