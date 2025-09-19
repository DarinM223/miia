package main

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/DarinM223/miia/graph"
	"github.com/DarinM223/miia/tokens"
)

var (
	ErrNumFirstIdent   error = errors.New("digit is scanned as the first character in an ident")
	ErrInvalidCh             = errors.New("invalid character scanned")
	ErrInvalidToken          = errors.New("invalid token scanned")
	ErrExpectedStr           = errors.New("string different from expected")
	ErrNum                   = errors.New("error parsing number")
	ErrStringNotClosed       = errors.New("string does not have an opening or closing quote")
	ErrGotoNotString         = errors.New("goto URL is not a string type")
	ErrBindingNotIdent       = errors.New("binding statment must start with an ident")
	ErrPosOutOfBounds        = errors.New("text index is greater than the text length")
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

// parseIdent parses an identifier from the file.
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
				return "", ErrNumFirstIdent
			}
			if err := ident.WriteByte(ch); err != nil {
				return "", fmt.Errorf("error writing to identifier buffer: %w", err)
			}
			p.pos++
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == ')' || ch == '(':
			return ident.String(), nil
		default:
			if err := ident.WriteByte(ch); err != nil {
				return "", fmt.Errorf("error writing to identifier buffer: %w", err)
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
			return ErrExpectedStr
		}

		ch := p.text[p.pos]
		if ch == expected[i] {
			p.pos++
		} else {
			p.pos = oldPos
			return ErrExpectedStr
		}
	}
	return nil
}

// parseKeywordOrIdent retrieves a ident string and checks if it
// is a keyword or not, returning the appropriate token.
func (p *Parser) parseKeywordOrIdent() (tokens.Token, string, error) {
	ident, err := p.parseIdent()
	if err != nil {
		return -1, "", fmt.Errorf("error parsing identifier: %w", err)
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
		return nil, ErrNum
	}

	if isNegative {
		num *= -1
	}
	return IntExpr{num}, nil
}

// parseString parses a string from the file.
func (p *Parser) parseString() (Expr, error) {
	if err := p.expectString("\""); err != nil {
		return nil, ErrStringNotClosed
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
				return nil, fmt.Errorf("error writing to string buffer: %w", err)
			}
			p.pos++
		}
	}
	return nil, ErrStringNotClosed
}

// parseExpr parses an expression from the file.
func (p *Parser) parseExpr() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, ErrPosOutOfBounds
	}

	ch := p.text[p.pos]
	switch {
	case ch == '(':
		p.pos++
		p.parseWhitespace()
		tok, name, err := p.parseKeywordOrIdent()
		if err != nil {
			return nil, fmt.Errorf("error parsing keyword or identifier: %w", err)
		}

		var expr Expr
		switch {
		case tok == tokens.IfToken:
			expr, err = p.parseIf()
		case tok == tokens.ForToken:
			expr, err = p.parseFor()
		case tok == tokens.CollectToken:
			expr, err = p.parseCollect()
		case tok == tokens.BlockToken:
			expr, err = p.parseBlock()
		case tok == tokens.RateLimitToken:
			expr, err = p.parseRateLimit()
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
			return nil, fmt.Errorf("expected expression to be closed by \")\": %w", err)
		}

		return expr, nil
	case isLetter(ch):
		lit, err := p.parseIdent()
		if err != nil {
			return nil, fmt.Errorf("error parsing variable expression: %w", err)
		}

		return VarExpr{lit}, nil
	case ('0' <= ch && ch <= '9') || ch == '-':
		return p.parseNumber()
	case ch == '"':
		return p.parseString()
	}
	return nil, errors.New("invalid expr")
}

// parseIf parses an if expression.
func (p *Parser) parseIf() (Expr, error) {
	p.parseWhitespace()
	pred, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing if predicate: %w", err)
	}

	p.parseWhitespace()
	conseq, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing if consequence: %w", err)
	}

	p.parseWhitespace()
	alt, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing if alternative: %w", err)
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
		return nil, fmt.Errorf("error parsing for binding identifier: %w", err)
	}

	p.parseWhitespace()
	collection, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing for collection: %w", err)
	}

	p.parseWhitespace()
	body, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing for body: %w", err)
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
		return nil, fmt.Errorf("error parsing unop expression: %w", err)
	}

	return UnOp{token, expr}, nil
}

// parseCollect parses a collect operation like (collect (for i nums (+ i 1))).
func (p *Parser) parseCollect() (Expr, error) {
	p.parseWhitespace()
	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing collect expression: %w", err)
	}
	return CollectExpr{expr}, nil
}

// parseBinOp parses a binary operation like (and a b).
func (p *Parser) parseBinOp(token tokens.Token) (Expr, error) {
	p.parseWhitespace()
	a, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing binop lhs: %w", err)
	}

	p.parseWhitespace()
	b, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing binop rhs: %w", err)
	}

	return BinOp{token, a, b}, nil
}

// parseMultOp parses an operation with more than two subexpressions
// like (+ 1 2 3 4).
func (p *Parser) parseMultOp(token tokens.Token) (Expr, error) {
	p.parseWhitespace()
	exprs, err := p.parseBlock()
	if err != nil {
		return nil, fmt.Errorf("error parsing multop expressions: %w", err)
	}

	if block, ok := exprs.(BlockExpr); ok {
		return MultOp{token, block.Exprs}, nil
	}
	return nil, errors.New("error parsing multop expressions: expression not a block")
}

// parseSelector parses a selector expression.
func (p *Parser) parseSelector() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, ErrPosOutOfBounds
	}

	ch := p.text[p.pos]
	var selectorList []graph.Selector
	for ch != ')' {
		p.parseWhitespace()
		ident, err := p.parseIdent()
		if err != nil {
			return nil, fmt.Errorf("error parsing selector binding: %w", err)
		}

		p.parseWhitespace()
		selectorExpr, err := p.parseString()
		if err != nil {
			return nil, fmt.Errorf("error parsing selector for binding %s: %w", ident, err)
		}

		if selector, ok := selectorExpr.(StringExpr); ok {
			selectorList = append(selectorList, graph.Selector{Name: ident, Selector: selector.Value})
		}

		p.parseWhitespace()
		if p.pos >= len(p.text) {
			return nil, ErrPosOutOfBounds
		}
		ch = p.text[p.pos]
	}
	return SelectorExpr{selectorList}, nil
}

// parseBlock parses a block of expressions.
func (p *Parser) parseBlock() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, ErrPosOutOfBounds
	}

	ch := p.text[p.pos]
	var exprList []Expr
	for ch != ')' {
		p.parseWhitespace()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, fmt.Errorf("error parsing expression at index %d: %w", len(exprList), err)
		}

		exprList = append(exprList, expr)

		p.parseWhitespace()
		if p.pos >= len(p.text) {
			return nil, ErrPosOutOfBounds
		}
		ch = p.text[p.pos]
	}
	return BlockExpr{exprList}, nil
}

// parseRateLimit parses a rate limiter expression.
func (p *Parser) parseRateLimit() (Expr, error) {
	p.parseWhitespace()
	url, err := p.parseString()
	if err != nil {
		return nil, fmt.Errorf("error parsing rate limit url: %w", err)
	}

	p.parseWhitespace()
	maxTimes, err := p.parseNumber()
	if err != nil {
		return nil, fmt.Errorf("error parsing rate limit max times: %w", err)
	}

	p.parseWhitespace()
	duration, err := p.parseNumber()
	if err != nil {
		return nil, fmt.Errorf("error parsing rate limit duration: %w", err)
	}

	return RateLimitExpr{
		url.(StringExpr).Value,
		maxTimes.(IntExpr).Value,
		time.Duration(duration.(IntExpr).Value) * time.Second,
	}, nil
}

// parseBindings parses a variable binding expression.
func (p *Parser) parseBindings() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, ErrPosOutOfBounds
	}

	ch := p.text[p.pos]
	bindings := make(map[string]Expr)
	for ch != ')' {
		p.parseWhitespace()
		name, err := p.parseIdent()
		if err != nil {
			return nil, fmt.Errorf("error parsing binding identifier: %w", err)
		}

		p.parseWhitespace()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, fmt.Errorf("error parsing expression for binding %s: %w", name, err)
		}

		bindings[name] = expr

		p.parseWhitespace()
		if p.pos >= len(p.text) {
			return nil, ErrPosOutOfBounds
		}
		ch = p.text[p.pos]
	}
	return BindExpr{bindings}, nil
}

// parseGoto parses a goto expression.
func (p *Parser) parseGoto() (Expr, error) {
	if p.pos >= len(p.text) {
		return nil, ErrPosOutOfBounds
	}

	p.parseWhitespace()
	url, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("error parsing goto expression: %w", err)
	}

	return GotoExpr{url}, nil
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}
