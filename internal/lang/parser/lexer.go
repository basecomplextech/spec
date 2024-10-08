// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

const (
	EOF   = 0
	ERROR = -1
)

var _ yyLexer = &lexer{}

type lexer struct {
	s *scanner.Scanner

	file *syntax.File // used by yyParser to return result
	err  error        // parse error
}

func newLexer(filename string, src io.Reader) *lexer {
	s := &scanner.Scanner{}
	s.Init(src)
	s.Filename = filename
	return &lexer{s: s}
}

func setLexerResult(l yyLexer, file *syntax.File) {
	l1 := l.(*lexer)
	l1.file = file
}

func (l *lexer) Lex(lval *yySymType) int {
	for {
		// Scan next token
		token := l.s.Scan()
		text := l.s.TokenText()

		// Return on eof
		if token == scanner.EOF {
			lval.yys = EOF
			return EOF
		}

		switch token {
		case scanner.Ident:
			keyword, ok := keywords[text]
			if ok {
				lval.yys = keyword
				lval.string = text

				if debugLexer {
					fmt.Printf("KEYWORD %v %v %v\n", l.s.Position, token, text)
				}

			} else {
				lval.yys = IDENT
				lval.ident = text

				if debugLexer {
					fmt.Printf("IDENT %v %v %v %v\n", l.s.Position, token, lval.yys, text)
				}
			}

			return lval.yys

		case scanner.Int:
			v, _ := strconv.ParseInt(text, 10, 64)
			lval.yys = INTEGER
			lval.integer = int(v)

			if debugLexer {
				fmt.Printf("INTEGER %v %v %v\n", l.s.Position, token, text)
			}
			return lval.yys

		case scanner.Float:
			lval.yys = int(token)
			lval.string = text

			if debugLexer {
				fmt.Printf("FLOAT %v %v %v\n", l.s.Position, token, text)
			}
			return lval.yys

		case scanner.String:
			lval.yys = STRING
			lval.string = text

			if debugLexer {
				fmt.Printf("STRING %v %v %v\n", l.s.Position, token, text)
			}
			return lval.yys

		case scanner.Comment:
			if debugLexer {
				fmt.Printf("COMMENT %v %v %v\n", l.s.Position, token, text)
			}
			continue

		default:
			lval.yys = int(token)
			lval.string = text

			if debugLexer {
				fmt.Printf("TOKEN %v %v %v\n", l.s.Position, token, text)
			}
			return lval.yys
		}
	}
}

func (l *lexer) Error(s string) {
	l.err = fmt.Errorf("%v %v", l.s.Position, s)
}

func yyLexError(l yyLexer, err error) int {
	ll := l.(*lexer)
	ll.err = fmt.Errorf("%v %w", ll.s.Position, err)
	return ERROR
}

func yyLexErrorf(l yyLexer, format string, a ...any) int {
	err := fmt.Errorf(format, a...)
	return yyLexError(l, err)
}

// util

func trimString(s string) string {
	return strings.Trim(s, "\"")
}
