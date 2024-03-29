// Code generated by goyacc -l -v grammar.out -o grammar.go grammar.y. DO NOT EDIT.
package parser

import __yyfmt__ "fmt"

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/ast"
)

type yySymType struct {
	yys int
	// Tokens
	ident   string
	integer int
	string  string

	// Type
	type_ *ast.Type

	// Import
	import_ *ast.Import
	imports []*ast.Import

	// Option
	option  *ast.Option
	options []*ast.Option

	// Definition
	definition  *ast.Definition
	definitions []*ast.Definition

	// Enum
	enum_value  *ast.EnumValue
	enum_values []*ast.EnumValue

	// Field
	field  *ast.Field
	fields ast.Fields

	// Struct
	struct_field  *ast.StructField
	struct_fields []*ast.StructField

	// Service
	service        *ast.Service
	method         *ast.Method
	methods        []*ast.Method
	method_input   ast.MethodInput
	method_output  ast.MethodOutput
	method_channel *ast.MethodChannel
	method_field   *ast.Field
	method_fields  ast.Fields
}

const ANY = 57346
const ENUM = 57347
const IMPORT = 57348
const MESSAGE = 57349
const OPTIONS = 57350
const STRUCT = 57351
const SERVICE = 57352
const SUBSERVICE = 57353
const IDENT = 57354
const INTEGER = 57355
const STRING = 57356
const METHOD_OUTPUT = 57357

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"ANY",
	"ENUM",
	"IMPORT",
	"MESSAGE",
	"OPTIONS",
	"STRUCT",
	"SERVICE",
	"SUBSERVICE",
	"IDENT",
	"INTEGER",
	"STRING",
	"METHOD_OUTPUT",
	"'('",
	"')'",
	"'='",
	"'['",
	"']'",
	"'.'",
	"'{'",
	"'}'",
	"';'",
	"'<'",
	"'-'",
	"'>'",
	"','",
}

var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

var yyExca = [...]int8{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 95,
	17, 24,
	-2, 1,
	-1, 96,
	17, 26,
	-2, 3,
	-1, 97,
	17, 27,
	-2, 5,
}

const yyPrivate = 57344

const yyLast = 180

var yyAct = [...]uint8{
	63, 99, 108, 64, 106, 90, 100, 43, 96, 55,
	48, 97, 50, 51, 52, 53, 95, 47, 114, 48,
	49, 50, 51, 52, 53, 45, 123, 122, 119, 109,
	110, 92, 116, 102, 101, 77, 75, 110, 88, 89,
	62, 39, 38, 44, 37, 36, 35, 81, 60, 56,
	80, 76, 47, 40, 48, 49, 50, 51, 52, 53,
	45, 71, 74, 74, 126, 125, 121, 118, 117, 44,
	78, 72, 82, 47, 112, 48, 49, 50, 51, 52,
	53, 45, 111, 67, 86, 25, 68, 24, 93, 94,
	23, 66, 69, 104, 33, 57, 105, 103, 65, 31,
	84, 115, 8, 6, 34, 85, 79, 16, 105, 17,
	120, 18, 19, 20, 67, 87, 124, 68, 30, 29,
	28, 27, 66, 127, 128, 47, 26, 48, 49, 50,
	51, 52, 53, 45, 47, 5, 48, 49, 50, 51,
	52, 53, 45, 96, 58, 48, 97, 50, 51, 52,
	53, 95, 3, 113, 61, 1, 98, 107, 91, 83,
	73, 15, 14, 54, 70, 13, 42, 12, 41, 59,
	11, 7, 10, 4, 21, 32, 2, 9, 22, 46,
}

var yyPact = [...]int16{
	146, -1000, 127, 87, -1000, 86, -1000, 102, -1000, 73,
	-1000, -1000, -1000, -1000, -1000, -1000, 114, 109, 108, 107,
	106, 82, -1000, -1000, -1000, 90, 24, 23, 22, 20,
	19, -1000, -1000, 35, -1000, -1000, 130, -1000, -1000, -1000,
	81, 121, 16, -1000, 79, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 69, 48, 13, -1000, -1000, -1000,
	33, 12, 130, 93, -1000, 30, 26, -1000, -1000, -1000,
	-1000, 79, -1000, -1000, 84, -1000, 92, -1000, -1000, -1000,
	110, 103, 14, 15, 139, 10, -1000, -1000, -1000, -1000,
	9, 77, 4, 65, 57, 26, -1000, -1000, -10, -1000,
	79, -1000, -1000, 8, 139, 51, 50, 11, 49, 1,
	-1, -1000, -1000, -1000, 130, 52, -1000, -1000, -1000, -1000,
	47, -1000, 79, 79, -1000, -1000, -1000, -1000, -1000,
}

var yyPgo = [...]uint8{
	0, 179, 6, 178, 177, 176, 175, 174, 173, 0,
	3, 172, 171, 170, 169, 168, 167, 7, 166, 165,
	164, 163, 162, 161, 9, 160, 159, 5, 158, 157,
	2, 1, 156, 4, 155, 154, 153,
}

var yyR1 = [...]int8{
	0, 2, 2, 1, 1, 1, 1, 1, 1, 1,
	34, 3, 3, 4, 4, 5, 5, 8, 8, 7,
	7, 6, 9, 9, 10, 10, 10, 10, 11, 11,
	11, 11, 11, 12, 12, 13, 14, 15, 15, 16,
	17, 18, 18, 18, 19, 20, 21, 21, 22, 23,
	24, 24, 25, 25, 25, 26, 26, 27, 27, 28,
	28, 28, 29, 30, 33, 32, 32, 32, 31, 36,
	36, 35, 35,
}

var yyR2 = [...]int8{
	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	3, 1, 2, 0, 2, 0, 4, 0, 4, 0,
	2, 3, 1, 3, 1, 3, 1, 1, 1, 1,
	1, 1, 1, 0, 2, 5, 4, 0, 2, 6,
	3, 0, 1, 3, 5, 3, 0, 2, 5, 5,
	0, 2, 3, 4, 5, 3, 3, 3, 3, 3,
	3, 4, 3, 3, 2, 0, 1, 3, 3, 0,
	1, 0, 1,
}

var yyChk = [...]int16{
	-1000, -34, -5, 6, -8, 8, 16, -12, 16, -4,
	-11, -13, -16, -19, -22, -23, 5, 7, 9, 10,
	11, -7, -3, 17, 14, 12, 12, 12, 12, 12,
	12, 17, -6, 12, 14, 22, 22, 22, 22, 22,
	18, -15, -18, -17, -2, 12, -1, 4, 6, 7,
	8, 9, 10, 11, -21, -24, -24, 14, 23, -14,
	-2, -35, 24, -9, -10, 19, 12, 4, 7, 23,
	-20, -2, 23, -25, -2, 23, 18, 23, -17, 13,
	20, 21, -9, -26, 16, 13, -10, 12, 24, 24,
	-27, -28, 16, -10, -33, 12, 4, 7, -32, -31,
	-2, 24, 24, -27, 16, -10, -33, -29, -30, 25,
	26, 17, 17, -36, 28, -9, 24, 17, 17, 17,
	-30, 17, 26, 27, -31, 13, 17, -9, -9,
}

var yyDef = [...]int8{
	15, -2, 17, 0, 33, 0, 13, 10, 19, 0,
	34, 28, 29, 30, 31, 32, 0, 0, 0, 0,
	0, 0, 14, 16, 11, 0, 0, 0, 0, 0,
	0, 18, 20, 0, 12, 37, 41, 46, 50, 50,
	0, 0, 71, 42, 0, 1, 2, 3, 4, 5,
	6, 7, 8, 9, 0, 0, 0, 21, 35, 38,
	0, 0, 72, 0, 22, 0, 24, 26, 27, 44,
	47, 0, 48, 51, 0, 49, 0, 39, 43, 40,
	0, 0, 0, 0, 65, 0, 23, 25, 45, 52,
	0, 0, 65, 0, 0, -2, -2, -2, 69, 66,
	0, 36, 53, 0, 65, 0, 0, 0, 0, 0,
	0, 55, 56, 64, 70, 0, 54, 57, 58, 59,
	0, 60, 0, 0, 67, 68, 61, 62, 63,
}

var yyTok1 = [...]int8{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	16, 17, 3, 3, 28, 26, 21, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 24,
	25, 18, 27, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 19, 3, 20, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 22, 3, 23,
}

var yyTok2 = [...]int8{
	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15,
}

var yyTok3 = [...]int8{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := int(yyPact[state])
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && int(yyChk[int(yyAct[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || int(yyExca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := int(yyExca[i])
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = int(yyTok1[0])
		goto out
	}
	if char < len(yyTok1) {
		token = int(yyTok1[char])
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = int(yyTok2[char-yyPrivate])
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = int(yyTok3[i+0])
		if token == char {
			token = int(yyTok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int(yyTok2[1]) /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = int(yyPact[yystate])
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = int(yyAct[yyn])
	if int(yyChk[yyn]) == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = int(yyDef[yystate])
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && int(yyExca[xi+1]) == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = int(yyExca[xi+0])
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = int(yyExca[xi+1])
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = int(yyPact[yyS[yyp].yys]) + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = int(yyAct[yyn]) /* simulate a shift of "error" */
					if int(yyChk[yystate]) == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= int(yyR2[yyn])
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = int(yyR1[yyn])
	yyg := int(yyPgo[yyn])
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = int(yyAct[yyg])
	} else {
		yystate = int(yyAct[yyj])
		if int(yyChk[yystate]) != -yyn {
			yystate = int(yyAct[yyg])
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = yyDollar[1].ident
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = yyDollar[1].ident
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = "any"
		}
	case 4:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = "import"
		}
	case 5:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = "message"
		}
	case 6:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = "options"
		}
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = "struct"
		}
	case 8:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = "service"
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			yyVAL.ident = "subservice"
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			file := &ast.File{
				Imports:     yyDollar[1].imports,
				Options:     yyDollar[2].options,
				Definitions: yyDollar[3].definitions,
			}
			setLexerResult(yylex, file)
		}
	case 11:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			if debugParser {
				fmt.Println("import ", yyDollar[1].string)
			}
			yyVAL.import_ = &ast.Import{
				ID: trimString(yyDollar[1].string),
			}
		}
	case 12:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			if debugParser {
				fmt.Println("import ", yyDollar[1].ident, yyDollar[2].string)
			}
			yyVAL.import_ = &ast.Import{
				Alias: yyDollar[1].ident,
				ID:    trimString(yyDollar[2].string),
			}
		}
	case 13:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.imports = nil
		}
	case 14:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			if debugParser {
				fmt.Println("import_list", yyDollar[1].imports, yyDollar[2].import_)
			}
			yyVAL.imports = append(yyVAL.imports, yyDollar[2].import_)
		}
	case 15:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.imports = nil
		}
	case 16:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			if debugParser {
				fmt.Println("imports", yyDollar[3].imports)
			}
			yyVAL.imports = append(yyVAL.imports, yyDollar[3].imports...)
		}
	case 17:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.options = nil
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			if debugParser {
				fmt.Println("options", yyDollar[3].options)
			}
			yyVAL.options = append(yyVAL.options, yyDollar[3].options...)
		}
	case 19:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.options = nil
		}
	case 20:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			if debugParser {
				fmt.Println("option_list", yyDollar[1].options, yyDollar[2].option)
			}
			yyVAL.options = append(yyVAL.options, yyDollar[2].option)
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("option ", yyDollar[1].ident, yyDollar[3].string)
			}
			yyVAL.option = &ast.Option{
				Name:  yyDollar[1].ident,
				Value: trimString(yyDollar[3].string),
			}
		}
	case 22:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			if debugParser {
				fmt.Printf("type *%v\n", yyDollar[1].type_)
			}
			yyVAL.type_ = yyDollar[1].type_
		}
	case 23:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Printf("type []%v\n", yyDollar[3].type_)
			}
			yyVAL.type_ = &ast.Type{
				Kind:    ast.KindList,
				Element: yyDollar[3].type_,
			}
		}
	case 24:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			if debugParser {
				fmt.Println("base type", yyDollar[1].ident)
			}
			yyVAL.type_ = &ast.Type{
				Kind: ast.GetKind(yyDollar[1].ident),
				Name: yyDollar[1].ident,
			}
		}
	case 25:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Printf("base type %v.%v\n", yyDollar[1].ident, yyDollar[3].ident)
			}
			yyVAL.type_ = &ast.Type{
				Kind:   ast.KindReference,
				Name:   yyDollar[3].ident,
				Import: yyDollar[1].ident,
			}
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			if debugParser {
				fmt.Println("base type", "any")
			}
			yyVAL.type_ = &ast.Type{
				Kind: ast.KindAny,
				Name: "any",
			}
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			if debugParser {
				fmt.Println("base type", "message")
			}
			yyVAL.type_ = &ast.Type{
				Kind: ast.KindAnyMessage,
				Name: "message",
			}
		}
	case 33:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.definitions = nil
		}
	case 34:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			if debugParser {
				fmt.Println("definitions", yyDollar[1].definitions, yyDollar[2].definition)
			}
			yyVAL.definitions = append(yyVAL.definitions, yyDollar[2].definition)
		}
	case 35:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			if debugParser {
				fmt.Println("enum", yyDollar[2].ident, yyDollar[4].enum_values)
			}
			yyVAL.definition = &ast.Definition{
				Type: ast.DefinitionEnum,
				Name: yyDollar[2].ident,

				Enum: &ast.Enum{
					Values: yyDollar[4].enum_values,
				},
			}
		}
	case 36:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			if debugParser {
				fmt.Println("enum value", yyDollar[1].ident, yyDollar[3].integer)
			}
			yyVAL.enum_value = &ast.EnumValue{
				Name:  yyDollar[1].ident,
				Value: yyDollar[3].integer,
			}
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.enum_values = nil
		}
	case 38:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			if debugParser {
				fmt.Println("enum values", yyDollar[1].enum_values, yyDollar[2].enum_value)
			}
			yyVAL.enum_values = append(yyVAL.enum_values, yyDollar[2].enum_value)
		}
	case 39:
		yyDollar = yyS[yypt-6 : yypt+1]
		{
			if debugParser {
				fmt.Println("message", yyDollar[2].ident, yyDollar[4].fields)
			}
			yyVAL.definition = &ast.Definition{
				Type: ast.DefinitionMessage,
				Name: yyDollar[2].ident,

				Message: &ast.Message{
					Fields: yyDollar[4].fields,
				},
			}
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("message field", yyDollar[1].ident, yyDollar[2].type_, yyDollar[3].integer)
			}
			yyVAL.field = &ast.Field{
				Name: yyDollar[1].ident,
				Type: yyDollar[2].type_,
				Tag:  yyDollar[3].integer,
			}
		}
	case 41:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.fields = nil
		}
	case 42:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			if debugParser {
				fmt.Println("message fields", yyDollar[1].field)
			}
			yyVAL.fields = []*ast.Field{yyDollar[1].field}
		}
	case 43:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("message fields", yyDollar[1].fields, yyDollar[3].field)
			}
			yyVAL.fields = append(yyVAL.fields, yyDollar[3].field)
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			if debugParser {
				fmt.Println("struct", yyDollar[2].ident, yyDollar[4].struct_fields)
			}
			yyVAL.definition = &ast.Definition{
				Type: ast.DefinitionStruct,
				Name: yyDollar[2].ident,

				Struct: &ast.Struct{
					Fields: yyDollar[4].struct_fields,
				},
			}
		}
	case 45:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("struct field", yyDollar[1].ident, yyDollar[2].type_)
			}
			yyVAL.struct_field = &ast.StructField{
				Name: yyDollar[1].ident,
				Type: yyDollar[2].type_,
			}
		}
	case 46:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.struct_fields = nil
		}
	case 47:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			if debugParser {
				fmt.Println("struct fields", yyDollar[1].struct_fields, yyDollar[2].struct_field)
			}
			yyVAL.struct_fields = append(yyVAL.struct_fields, yyDollar[2].struct_field)
		}
	case 48:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			if debugParser {
				fmt.Println("service", yyDollar[2].ident, yyDollar[4].methods)
			}
			yyVAL.definition = &ast.Definition{
				Type: ast.DefinitionService,
				Name: yyDollar[2].ident,

				Service: &ast.Service{
					Methods: yyDollar[4].methods,
				},
			}
		}
	case 49:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			if debugParser {
				fmt.Println("subservice", yyDollar[2].ident, yyDollar[4].methods)
			}
			yyVAL.definition = &ast.Definition{
				Type: ast.DefinitionService,
				Name: yyDollar[2].ident,

				Service: &ast.Service{
					Sub:     true,
					Methods: yyDollar[4].methods,
				},
			}
		}
	case 50:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.methods = nil
		}
	case 51:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			yyVAL.methods = append(yyDollar[1].methods, yyDollar[2].method)
		}
	case 52:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method", yyDollar[1].ident, yyDollar[2].method_input)
			}
			yyVAL.method = &ast.Method{
				Name:  yyDollar[1].ident,
				Input: yyDollar[2].method_input,
			}
		}
	case 53:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			if debugParser {
				fmt.Println("method", yyDollar[1].ident, yyDollar[2].method_input, yyDollar[3].method_output)
			}
			yyVAL.method = &ast.Method{
				Name:   yyDollar[1].ident,
				Input:  yyDollar[2].method_input,
				Output: yyDollar[3].method_output,
			}
		}
	case 54:
		yyDollar = yyS[yypt-5 : yypt+1]
		{
			if debugParser {
				fmt.Println("method", yyDollar[1].ident, yyDollar[2].method_input, yyDollar[3].method_channel, yyDollar[4].method_output)
			}
			yyVAL.method = &ast.Method{
				Name:    yyDollar[1].ident,
				Input:   yyDollar[2].method_input,
				Channel: yyDollar[3].method_channel,
				Output:  yyDollar[4].method_output,
			}
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method input", yyDollar[2].type_)
			}
			yyVAL.method_input = yyDollar[2].type_
		}
	case 56:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method input", yyDollar[2].fields)
			}
			yyVAL.method_input = yyDollar[2].fields
		}
	case 57:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method output", yyDollar[2].type_)
			}
			yyVAL.method_output = yyDollar[2].type_
		}
	case 58:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method output", yyDollar[2].fields)
			}
			yyVAL.method_output = yyDollar[2].fields
		}
	case 59:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method channel", yyDollar[2].type_)
			}

			yyVAL.method_channel = &ast.MethodChannel{
				In: yyDollar[2].type_,
			}
		}
	case 60:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method channel", yyDollar[2].type_)
			}

			yyVAL.method_channel = &ast.MethodChannel{
				Out: yyDollar[2].type_,
			}
		}
	case 61:
		yyDollar = yyS[yypt-4 : yypt+1]
		{
			if debugParser {
				fmt.Println("method channel", yyDollar[2].type_, yyDollar[3].type_)
			}

			yyVAL.method_channel = &ast.MethodChannel{
				In:  yyDollar[2].type_,
				Out: yyDollar[3].type_,
			}
		}
	case 62:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.type_ = yyDollar[3].type_
		}
	case 63:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			yyVAL.type_ = yyDollar[3].type_
		}
	case 64:
		yyDollar = yyS[yypt-2 : yypt+1]
		{
			if debugParser {
				fmt.Println("method field list", yyDollar[1].fields)
			}
			yyVAL.fields = yyDollar[1].fields
		}
	case 65:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
			yyVAL.fields = nil
		}
	case 66:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
			if debugParser {
				fmt.Println("method fields", yyDollar[1].field)
			}
			yyVAL.fields = []*ast.Field{yyDollar[1].field}
		}
	case 67:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method fields", yyDollar[1].fields, yyDollar[3].field)
			}
			yyVAL.fields = append(yyDollar[1].fields, yyDollar[3].field)
		}
	case 68:
		yyDollar = yyS[yypt-3 : yypt+1]
		{
			if debugParser {
				fmt.Println("method field", yyDollar[1].ident, yyDollar[2].type_, yyDollar[3].integer)
			}
			yyVAL.field = &ast.Field{
				Name: yyDollar[1].ident,
				Type: yyDollar[2].type_,
				Tag:  yyDollar[3].integer,
			}
		}
	case 69:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
		}
	case 70:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
		}
	case 71:
		yyDollar = yyS[yypt-0 : yypt+1]
		{
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		{
		}
	}
	goto yystack /* stack new state and value */
}
