package main

import (
	"bytes"
	"errors"
	"github.com/DarinM223/miia/graph"
	"github.com/DarinM223/miia/tokens"
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
		case '0' <= ch && ch <= '9':
			// Idents cannot start with a digit
			if i == 0 {
				return "", NumFirstIdentErr
			}
			if err := ident.WriteByte(ch); err != nil {
				return "", err
			}
			p.pos++
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == ')' || ch == '(':
			return ident.String(), nil
		default:
			if err := ident.WriteByte(ch); err != nil {
				return "", err
			}
			p.pos++
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
func (p *Parser) parseKeywordOrIdent() (tokens.Token, string, error) {
	ident, err := p.parseIdent()
	if err != nil {
		return -1, "", err
	}

	if token, err := tokens.Lookup(ident, tokens.Tokens); err == nil {
		return tokens.Token(token), ident, nil
	}

	return tokens.IdentToken, ident, nil
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
		return nil, NumErr
	}

	if isNegative {
		num *= -1
	}
	return IntExpr{num}, nil
}

// parseString parses a string from the file.
func (p *Parser) parseString() (Expr, error) {
	if err := p.expectString("\""); err != nil {
		return nil, StringNotClosedErr
	}

	var str bytes.Buffer
	for {
		if p.pos >= len(p.text) {
			break
		}

		ch := p.text[p.pos]
		if ch == '"' {
			p.pos++
			return StringExpr{str.String()}, nil
		} else {
			if err := str.WriteByte(ch); err != nil {
				return nil, err
			}
			p.pos++
		}
	}
	return nil, StringNotClosedErr
}

func (p *Parser) parseExpr() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, PosOutOfBoundsErr
	}

	ch := p.text[p.pos]
	switch {
	case ch == '(':
		p.pos++
		p.parseWhitespace()
		tok, name, err := p.parseKeywordOrIdent()
		if err != nil {
			return nil, err
		}

		var expr Expr
		switch {
		case tok == tokens.IfToken:
			expr, err = p.parseIf()
		case tok == tokens.ForToken:
			expr, err = p.parseFor()
		case tok == tokens.BlockToken:
			expr, err = p.parseBlock()
		case tok == tokens.AssignToken:
			expr, err = p.parseBindings()
		case tok == tokens.GotoToken:
			expr, err = p.parseGoto()
		case tok == tokens.SelectorToken:
			expr, err = p.parseSelector()
		case tokens.IsUnaryOp(name):
			expr, err = p.parseUnOp(tok)
		case tokens.IsMultOp(name):
			expr, err = p.parseMultOp(tok)
		case tokens.IsBinaryOp(name):
			expr, err = p.parseBinOp(tok)
		default:
			panic("Invalid token type")
		}

		if err != nil {
			return nil, err
		}

		p.parseWhitespace()
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

// parseUnOp parses a unary operation like (not true).
func (p *Parser) parseUnOp(token tokens.Token) (Expr, error) {
	p.parseWhitespace()
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return UnOp{token, expr}, nil
}

// parseBinOp parses a binary operation like (and a b).
func (p *Parser) parseBinOp(token tokens.Token) (Expr, error) {
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

// parseMultOp parses an operation with more than two subexpressions
// like (+ 1 2 3 4).
func (p *Parser) parseMultOp(token tokens.Token) (Expr, error) {
	p.parseWhitespace()
	exprs, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if block, ok := exprs.(BlockExpr); ok {
		return MultOp{token, block.Exprs}, nil
	}
	return nil, errors.New("Expression not a block")
}

// parseSelector parses a selector expression.
func (p *Parser) parseSelector() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, PosOutOfBoundsErr
	}

	ch := p.text[p.pos]
	var selectorList []graph.Selector
	for ch != ')' {
		p.parseWhitespace()
		ident, err := p.parseIdent()
		if err != nil {
			return nil, err
		}

		p.parseWhitespace()
		selectorExpr, err := p.parseString()
		if err != nil {
			return nil, err
		}

		if selector, ok := selectorExpr.(StringExpr); ok {
			selectorList = append(selectorList, graph.Selector{ident, selector.Value})
		}

		p.parseWhitespace()
		if p.pos >= len(p.text) {
			return nil, PosOutOfBoundsErr
		}
		ch = p.text[p.pos]
	}
	return SelectorExpr{selectorList}, nil
}

// parseBlock parses a block of expressions.
func (p *Parser) parseBlock() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, PosOutOfBoundsErr
	}

	ch := p.text[p.pos]
	var exprList []Expr
	for ch != ')' {
		p.parseWhitespace()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		exprList = append(exprList, expr)

		p.parseWhitespace()
		if p.pos >= len(p.text) {
			return nil, PosOutOfBoundsErr
		}
		ch = p.text[p.pos]
	}
	return BlockExpr{exprList}, nil
}

func (p *Parser) parseBindings() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, PosOutOfBoundsErr
	}

	ch := p.text[p.pos]
	bindings := make(map[string]Expr)
	for ch != ')' {
		p.parseWhitespace()
		name, err := p.parseIdent()
		if err != nil {
			return nil, err
		}

		p.parseWhitespace()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		bindings[name] = expr

		p.parseWhitespace()
		if p.pos >= len(p.text) {
			return nil, PosOutOfBoundsErr
		}
		ch = p.text[p.pos]
	}
	return BindExpr{bindings}, nil
}

// parseGoto parses a goto expression.
func (p *Parser) parseGoto() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, PosOutOfBoundsErr
	}

	p.parseWhitespace()
	url, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return GotoExpr{url}, nil
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}
