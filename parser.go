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
)

type Token int

const (
	IdentToken Token = iota
	LParenToken
	RParenToken
	LBraceToken
	RBraceToken
	RangeToken
	ColonToken
	ForToken
	IfToken
	ElseToken
	GotoToken
	AssignToken
	AndToken
	OrToken
	NotToken
	EqualsToken
)

var tokens = map[string]Token{
	"for":  ForToken,
	"if":   IfToken,
	"else": ElseToken,
	"goto": GotoToken,
}

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

	if token, err := lookupToken(ident); err == nil {
		return Token(token), ident, nil
	}

	return IdentToken, ident, nil
}

/*
 * Expression parsing functions
 */

// parseFor parses a for expression.
func (p *Parser) parseFor() (Expr, error) {
	p.parseWhitespace()
	name, err := p.parseIdent()
	if err != nil {
		return ForExpr{}, err
	}

	p.parseWhitespace()
	if err := p.expectString("in"); err != nil {
		return ForExpr{}, err
	}

	p.parseWhitespace()
	collection, err := p.parseExpr()
	if err != nil {
		return ForExpr{}, err
	}

	p.parseWhitespace()
	body, err := p.parseBlock()
	if err != nil {
		return ForExpr{}, err
	}

	return ForExpr{
		Collection: collection,
		Name:       name,
		Body:       body,
	}, nil
}

// parseIf parses an if expression.
func (p *Parser) parseIf() (Expr, error) {
	p.parseWhitespace()
	pred, err := p.parseExpr()
	if err != nil {
		return IfExpr{}, err
	}

	p.parseWhitespace()
	conseq, err := p.parseBlock()
	if err != nil {
		return IfExpr{}, err
	}

	p.parseWhitespace()
	if err := p.expectString("else"); err != nil {
		return IfExpr{}, err
	}

	p.parseWhitespace()
	alt, err := p.parseBlock()
	if err != nil {
		return IfExpr{}, err
	}

	return IfExpr{
		Pred:   pred,
		Conseq: conseq,
		Alt:    alt,
	}, nil
}

// parseNum parses an integer expression from the file.
func (p *Parser) parseNumber() (Expr, error) {
	num := -1

	for {
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
		} else {
			break
		}
	}
	if num == -1 {
		return IntExpr{}, NumErr
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
	p.parseWhitespace()
	ch := p.text[p.pos]
	switch {
	case isLetter(ch):
		token, lit, err := p.parseKeywordOrIdent()
		if err != nil {
			return nil, err
		}

		// TODO(DarinM223): account for operators and functions
		switch token {
		case IdentToken:
			return VarExpr{lit}, nil
		case ForToken:
			return p.parseFor()
		case IfToken:
			return p.parseIf()
		}
	case '0' <= ch && ch <= '9':
		// TODO(DarinM223): account for operators and functions
		return p.parseNumber()
	case ch == '"':
		return p.parseString()
	}
	return nil, nil
}

/*
 * Statement parsing functions
 */

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

func lookupToken(s string) (Token, error) {
	if token, ok := tokens[s]; ok {
		return token, nil
	}

	return -1, errors.New("String is not a token")
}
