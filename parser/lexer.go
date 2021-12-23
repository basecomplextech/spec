package parser

import (
	"errors"
	"unicode"
)

const (
	yyEOF = 0
)

var (
	yyEOFstr = []byte("eof")
)

var _ yyLexer = &yyLexerImpl{}

type yyLexerImpl struct {
	b   []byte
	pos int

	err   error
	stmts []interface{}
}

func yyNewLexer(b []byte) *yyLexerImpl {
	return &yyLexerImpl{
		b:   b,
		pos: 0,
	}
}

func yyLexAppendStmts(l yyLexer, stmt ...interface{}) {
	ll := l.(*yyLexerImpl)
	ll.stmts = append(ll.stmts, stmt...)
}

func (l *yyLexerImpl) Lex(lval *yySymType) int {
	l.skipWhitespace()

	for {
		ch := l.peek()

		switch {
		case ch == yyEOF:
			return l.scanEOF(lval)

		case isIdentStart(ch):
			return l.scanIdent(lval)

		default:
			l.pos++
			lval.yys = ch
			return ch
		}
	}
}

func (l *yyLexerImpl) Error(s string) {
	l.err = errors.New(s)
}

func (l *yyLexerImpl) scanEOF(lval *yySymType) int {
	lval.yys = yyEOF
	// lval.str = yyEOFstr
	return lval.yys
}

func (l *yyLexerImpl) scanIdent(lval *yySymType) int {
	// start := l.pos

	for {
		ch := l.peek()
		if isIdentMiddle(ch) {
			l.pos++
			continue
		}
		break
	}

	// lval.ident = l.b[start:l.pos]
	// lval.yys = yyGetKeywordID(lval.ident)
	return lval.yys
}

func (l *yyLexerImpl) skipWhitespace() {
	for {
		ch := l.peek()

		if unicode.IsSpace(rune(ch)) {
			l.pos++
			continue
		}

		break
	}
}

func (l *yyLexerImpl) peek() int {
	if l.pos >= len(l.b) {
		return yyEOF
	}

	return int(l.b[l.pos])
}

// isIdentStart returns true if the character is valid at the start of an identifier.
func isIdentStart(ch int) bool {
	return (ch >= 'A' && ch <= 'Z') ||
		(ch >= 'a' && ch <= 'z') ||
		(ch >= 128 && ch <= 255) ||
		(ch == '_')
}

// isIdentMiddle returns true if the character is valid inside an identifier.
func isIdentMiddle(ch int) bool {
	return isIdentStart(ch) || unicode.IsDigit(rune(ch))
}
