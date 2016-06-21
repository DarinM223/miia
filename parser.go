package main

import (
	"bytes"
	"errors"
)

var (
	NumFirstIdentErr error = errors.New("Digit is scanned as the first character in an ident")
	InvalidChErr           = errors.New("Invalid character scanned")
	InvalidTokenErr        = errors.New("Invalid token scanned")
	ExpectedStrErr         = errors.New("String different from expected")
	NumRangeErr            = errors.New("Error parsing number or range")
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

// expectString parses from the current position and checks if it
// matches the expected string. If it doesn't it returns an error.
func (p *Parser) expectString(expected string) error {
	for i := 0; i < len(expected); i++ {
		ch := p.text[p.pos]
		if ch == expected[i] {
			p.pos++
		} else {
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

	return IdentToken, ident, InvalidTokenErr
}

// parseFor parses a for expression.
func (p *Parser) parseFor() (Expr, error) {
	p.parseWhitespace()
	name, err := p.parseIdent()
	if err != nil {
		return ForExpr{}, err
	}

	p.parseWhitespace()
	p.expectString("in")

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
	p.expectString("else")

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

// parseNumOrRange either returns a integer expression
// or a range expression between two integers.
func (p *Parser) parseNumOrRange() (Expr, error) {
	isRange := false
	num1, num2 := -1, -1
	for {
		ch := p.text[p.pos]
		switch {
		case '0' <= ch && ch <= '9':
			if isRange {
				if num2 == -1 {
					num2 = 0
				}
				num2 *= 10
				num2 += int(ch) - int('0')
			} else {
				if num1 == -1 {
					num1 = 0
				}
				num1 *= 10
				num1 += int(ch) - int('0')
			}
			p.pos++
		case ch == '.':
			isRange = true
			p.pos++
			p.expectString(".")
		case ch == ' ' || ch == '\t' || ch == '\n':
			break
		default:
			return nil, NumRangeErr
		}
	}
	if isRange && num1 != -1 && num2 != -1 {
		return RangeExpr{IntExpr{num1}, IntExpr{num2}}, nil
	} else if !isRange && num1 != 1 {
		return IntExpr{num1}, nil
	}
	return nil, NumRangeErr
}

// parseString parses a string from the file.
func (p *Parser) parseString() (Expr, error) {
	p.expectString("\"")
	var str bytes.Buffer
	for {
		ch := p.text[p.pos]
		switch {
		case ch == '"':
			return StringExpr{str.String()}, nil
		default:
			if err := str.WriteByte(ch); err != nil {
				return StringExpr{}, err
			}
			p.pos++
		}
	}
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

func (p *Parser) parseStmt() (Stmt, error) {
	// TODO(DarinM223): implement this
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
