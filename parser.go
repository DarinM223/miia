package main

import (
	"bytes"
	"errors"
)

var (
	NumFirstIdentErr   error = errors.New("Digit is scanned as the first character in an ident")
	InvalidChErr             = errors.New("Invalid character scanned")
	InvalidTokenErr          = errors.New("Invalid token scanned")
	ExpectedStrErr           = errors.New("String different from expected")
	NumErr                   = errors.New("Error parsing number")
	StringNotClosedErr       = errors.New("String does not have an opening or closing quote")
	GotoNotStringErr         = errors.New("Goto URL is not a string type")
	BindingNotIdentErr       = errors.New("Binding statment must start with an ident")
	PosOutOfBoundsErr        = errors.New("Text index is greater than the text length")
)

type Token int

const (
	IdentToken Token = iota
	BlockToken
	RangeToken
	ForToken
	IfToken
	ElseToken
	GotoToken
	AssignToken
	AndToken
	OrToken
	NotToken
	EqualsToken
	AddToken
	SubToken
	MulToken
	DivToken
)

var keywords = map[string]Token{
	"block": BlockToken,
	"for":   ForToken,
	"if":    IfToken,
	"else":  ElseToken,
	"goto":  GotoToken,
}

var binOps = map[string]Token{
	"..":  RangeToken,
	"+":   AddToken,
	"-":   SubToken,
	"*":   MulToken,
	"/":   DivToken,
	"=":   EqualsToken,
	"or":  OrToken,
	"and": AndToken,
}

var unOps = map[string]Token{
	"not": NotToken,
}

var tokens = mergeMaps(keywords, binOps, unOps)

type Parser struct {
	pos  int
	text string
}

func NewParser(text string) *Parser {
	return &Parser{
		pos:  0,
		text: text,
	}
}

// parseIdent parses an ident from the file.
func (p *Parser) parseIdent() (string, error) {
	var ident bytes.Buffer
	for i := 0; ; i++ {
		if p.pos >= len(p.text) {
			return ident.String(), nil
		}

		ch := p.text[p.pos]
		switch {
		case isLetter(ch):
			if err := ident.WriteByte(ch); err != nil {
				return "", err
			}
			p.pos++
		case '0' <= ch && ch <= '9':
			// Idents cannot start with a digit
			if i == 0 {
				return "", NumFirstIdentErr
			}
			if err := ident.WriteByte(ch); err != nil {
				return "", err
			}
			p.pos++
		default:
			return ident.String(), nil
		}
	}
}

// parseWhitespace moves the cursor to a position where there is no
// whitespace characters.
func (p *Parser) parseWhitespace() {
	for {
		if p.pos >= len(p.text) {
			return
		}

		ch := p.text[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			p.pos++
		} else {
			return
		}
	}
}

// expectString parses from the current position and checks if it
// matches the expected string. If it doesn't it returns an error.
func (p *Parser) expectString(expected string) error {
	oldPos := p.pos
	for i := 0; i < len(expected); i++ {
		if p.pos >= len(p.text) {
			p.pos = oldPos
			return ExpectedStrErr
		}

		ch := p.text[p.pos]
		if ch == expected[i] {
			p.pos++
		} else {
			p.pos = oldPos
			return ExpectedStrErr
		}
	}
	return nil
}

// parseKeywordOrIdent retrieves a ident string and checks if it
// is a keyword or not, returning the appropriate token.
func (p *Parser) parseKeywordOrIdent() (Token, string, error) {
	ident, err := p.parseIdent()
	if err != nil {
		return -1, "", err
	}

	if token, err := lookup(ident, tokens); err == nil {
		return Token(token), ident, nil
	}

	return IdentToken, ident, nil
}

// parseNum parses an integer expression from the file.
func (p *Parser) parseNumber() (Expr, error) {
	num := -1
	isNegative := false

	for i := 0; ; i++ {
		if p.pos >= len(p.text) {
			break
		}

		ch := p.text[p.pos]
		if '0' <= ch && ch <= '9' {
			if num == -1 {
				num = 0
			}
			num *= 10
			num += int(ch) - int('0')
			p.pos++
		} else if i == 0 && ch == '-' {
			isNegative = true
			p.pos++
		} else {
			break
		}
	}
	if num == -1 {
		return IntExpr{}, NumErr
	}

	if isNegative {
		num *= -1
	}
	return IntExpr{num}, nil
}

// parseString parses a string from the file.
func (p *Parser) parseString() (Expr, error) {
	if err := p.expectString("\""); err != nil {
		return StringExpr{}, StringNotClosedErr
	}

	var str bytes.Buffer
	for {
		if p.pos >= len(p.text) {
			break
		}

		ch := p.text[p.pos]
		if ch == '"' {
			return StringExpr{str.String()}, nil
		} else {
			if err := str.WriteByte(ch); err != nil {
				return StringExpr{}, err
			}
			p.pos++
		}
	}
	return StringExpr{}, StringNotClosedErr
}

func (p *Parser) parseExpr() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, PosOutOfBoundsErr
	}

	ch := p.text[p.pos]
	switch {
	case ch == '(':
		tok, name, err := p.parseKeywordOrIdent()
		if err != nil {
			return nil, err
		}

		var expr Expr
		switch {
		case tok == IfToken:
			expr, err = p.parseIf()
		case tok == ForToken:
			expr, err = p.parseFor()
		//case tok == BlockToken:
		//    expr, err = p.parseBlock()
		case isUnaryOp(name):
			expr, err = p.parseUnOp(tok)
		case isBinaryOp(name):
			expr, err = p.parseBinOp(tok)
		}

		if err != nil {
			return nil, err
		}
		if err := p.expectString(")"); err != nil {
			return nil, err
		}

		return expr, nil
	case isLetter(ch):
		lit, err := p.parseIdent()
		if err != nil {
			return nil, err
		}

		return VarExpr{lit}, nil
	case ('0' <= ch && ch <= '9') || ch == '-':
		return p.parseNumber()
	case ch == '"':
		return p.parseString()
	}
	return nil, errors.New("Invalid expr")
}

// parseIf parses an if expression.
func (p *Parser) parseIf() (Expr, error) {
	p.parseWhitespace()
	pred, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.parseWhitespace()
	conseq, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.parseWhitespace()
	alt, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return IfExpr{
		Pred:   pred,
		Conseq: conseq,
		Alt:    alt,
	}, nil
}

// parseFor parses a for expression.
func (p *Parser) parseFor() (Expr, error) {
	p.parseWhitespace()
	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}

	p.parseWhitespace()
	collection, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.parseWhitespace()
	body, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return ForExpr{
		Collection: collection,
		Name:       name,
		Body:       body,
	}, nil
}

func (p *Parser) parseUnOp(token Token) (Expr, error) {
	p.parseWhitespace()
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return UnOp{token, expr}, nil
}

func (p *Parser) parseBinOp(token Token) (Expr, error) {
	p.parseWhitespace()
	a, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	p.parseWhitespace()
	b, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return BinOp{token, a, b}, nil
}

func (p *Parser) parseBlock() (Stmt, error) {
	if err := p.expectString("{"); err != nil {
		return SeqStmt{}, err
	}

	var stmtList []Stmt
	for {
		p.parseWhitespace()
		stmt, err := p.parseStmt()
		if err != nil {
			break
		}

		stmtList = append(stmtList, stmt)
	}

	// Generate a SeqStmt tree from the list of statements.
	var currStmt Stmt = nil
	for i := len(stmtList) - 2; i >= 0; i-- {
		if currStmt == nil {
			currStmt = SeqStmt{
				A: stmtList[i],
				B: stmtList[len(stmtList)-1],
			}
		} else {
			currStmt = SeqStmt{
				A: stmtList[i],
				B: currStmt,
			}
		}
	}

	if err := p.expectString("}"); err != nil {
		return SeqStmt{}, err
	}
	return currStmt, nil
}

func (p *Parser) parseBinding(ident string) (Stmt, error) {
	bindings := make(map[string]Expr)

	p.parseWhitespace()
	if err := p.expectString("="); err != nil {
		return BindingStmt{}, err
	}

	expr, err := p.parseExpr()
	if err != nil {
		return BindingStmt{}, err
	}
	bindings[ident] = expr

	for err := p.expectString(","); err == nil; err = p.expectString(",") {
		p.parseWhitespace()
		token, ident, err := p.parseKeywordOrIdent()
		if err != nil {
			return BindingStmt{}, err
		}

		if token != IdentToken {
			return BindingStmt{}, BindingNotIdentErr
		}

		if err := p.expectString("="); err != nil {
			return BindingStmt{}, err
		}

		expr, err := p.parseExpr()
		if err != nil {
			return BindingStmt{}, err
		}
		bindings[ident] = expr
	}

	return BindingStmt{bindings}, nil
}

func (p *Parser) parseGoto() (Stmt, error) {
	p.parseWhitespace()
	strExpr, err := p.parseString()
	if err != nil {
		return GotoStmt{}, err
	}

	if str, ok := strExpr.(StringExpr); ok {
		return GotoStmt{str.Value}, nil
	}
	return GotoStmt{}, GotoNotStringErr
}

func (p *Parser) parseStmt() (Stmt, error) {
	p.parseWhitespace()
	ch := p.text[p.pos]
	switch {
	case ch == '{':
		return p.parseBlock()
	case isLetter(ch):
		token, lit, err := p.parseKeywordOrIdent()
		if err != nil {
			return nil, err
		}

		switch token {
		case GotoToken:
			return p.parseGoto()
		case IdentToken:
			return p.parseBinding(lit)
		}
	}
	return nil, nil
}

/*
 * Utility functions
 */

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isBinaryOp(s string) bool {
	_, ok := binOps[s]
	return ok
}

func isUnaryOp(s string) bool {
	_, ok := binOps[s]
	return ok
}

func lookup(s string, dict map[string]Token) (Token, error) {
	if token, ok := dict[s]; ok {
		return token, nil
	}

	return -1, errors.New("String is not a token")
}

func mergeMaps(maps ...map[string]Token) map[string]Token {
	mergedMap := make(map[string]Token)
	for _, m := range maps {
		for k, v := range m {
			mergedMap[k] = v
		}
	}
	return mergedMap
}
