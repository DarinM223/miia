package main

import (
	"bytes"
	"errors"
)

var (
	NumFirstIdentErr error = errors.New("Digit is scanned as the first character in an ident")
	InvalidChErr           = errors.New("Invalid character scanned")
	InvalidTokenErr        = errors.New("Invalid token scanned")
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
	AssignToken
	AndToken
	OrToken
	NotToken
	EqualsToken
)

var tokens = [...]string{
	IdentToken:  "IDENT",
	LParenToken: "(",
	RParenToken: ")",
	LBraceToken: "{",
	RBraceToken: "}",
	RangeToken:  "..",
	ColonToken:  ":",
	ForToken:    "for",
	IfToken:     "if",
	AssignToken: "=",
	AndToken:    "&&",
	OrToken:     "||",
	NotToken:    "!",
	EqualsToken: "==",
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
		case ch == ' ', ch == '\t', ch == '\n':
			return ident.String(), nil
		default:
			return "", InvalidChErr
		}
	}
}

// parseWhitespace moves the cursor to a position where there is no
// whitespace characters.
func (p *Parser) parseWhitespace() {
	for {
		ch := p.text[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' {
			p.pos++
		} else {
			return
		}
	}
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

	return IdentToken, ident, InvalidTokenErr
}

func (p *Parser) parseFor() (Expr, error) {
	p.parseWhitespace()
	name, err := p.parseIdent()
	if err != nil {
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

func (p *Parser) parseIf() (Expr, error) {
	// TODO(DarinM223): implement this
	return IfExpr{}, nil
}

func (p *Parser) parseNumOrRange() (Expr, error) {
	// TODO(DarinM223): implement this
	return nil, nil
}

func (p *Parser) parseString() (Expr, error) {
	// TODO(DarinM223): implement this
	return nil, nil
}

func (p *Parser) parseBlock() (Stmt, error) {
	// TODO(DarinM223): implement this
	return nil, nil
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

		switch token {
		case IdentToken:
			return VarExpr{lit}, nil
		case ForToken:
			return p.parseFor()
		case IfToken:
			return p.parseIf()
		}
	case '0' <= ch && ch <= '9':
		return p.parseNumOrRange()
	case ch == '"':
		return p.parseString()
	}
	return nil, nil
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func lookupToken(s string) (Token, error) {
	for tok, tokLit := range tokens {
		if tokLit == s {
			return Token(tok), nil
		}
	}

	return -1, errors.New("String is not a token")
}
