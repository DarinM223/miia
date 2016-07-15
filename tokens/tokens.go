package tokens

import "errors"

type Token int

const (
	IdentToken Token = iota
	BlockToken
	SelectorToken
	RangeToken
	ListToken
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

var Keywords = map[string]Token{
	"block": BlockToken,
	"for":   ForToken,
	"if":    IfToken,
	"else":  ElseToken,
	"goto":  GotoToken,
	"sel":   SelectorToken,
	"set":   AssignToken,
}

var BinOps = map[string]Token{
	"..":  RangeToken,
	"=":   EqualsToken,
	"or":  OrToken,
	"and": AndToken,
}

var MultOps = map[string]Token{
	"+":    AddToken,
	"-":    SubToken,
	"*":    MulToken,
	"/":    DivToken,
	"list": ListToken,
}

var UnOps = map[string]Token{
	"not": NotToken,
}

var Tokens = mergeMaps(Keywords, BinOps, UnOps, MultOps)

func IsBinaryOp(s string) bool {
	_, ok := BinOps[s]
	return ok
}

func IsUnaryOp(s string) bool {
	_, ok := UnOps[s]
	return ok
}

func IsMultOp(s string) bool {
	_, ok := MultOps[s]
	return ok
}

func Lookup(s string, dict map[string]Token) (Token, error) {
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
