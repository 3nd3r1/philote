package syntax

import (
	"fmt"

	"github.com/3nd3r1/philote/internal/diag"
)

type tokKind int

const (
	tEOF tokKind = iota
	tNewline
	tIdent
	tInt
	tLBrace
	tRBrace
	tLParen
	tRParen
	tColon
	tComma
)

var tokNames = [...]string{"end of file", "end of line", "name", "number", "'{'", "'}'", "'('", "')'", "':'", "','"}

type token struct {
	kind tokKind
	text string
	pos  diag.Pos
}

func (t token) describe() string {
	if t.kind == tIdent || t.kind == tInt {
		return fmt.Sprintf("'%s'", t.text)
	}
	return tokNames[t.kind]
}

var punct = map[byte]tokKind{'{': tLBrace, '}': tRBrace, '(': tLParen, ')': tRParen, ':': tColon, ',': tComma}

func isLetter(c byte) bool { return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }

type lexer struct {
	src       string
	file      string
	i         int
	line, col int
	toks      []token
	d         *diag.List
}

func lex(file, src string, d *diag.List) []token {
	lx := &lexer{src: src, file: file, line: 1, col: 1, d: d}
	for lx.i < len(src) {
		lx.step()
	}
	lx.emit(tEOF, "", lx.pos())
	return lx.toks
}

func (lx *lexer) pos() diag.Pos { return diag.Pos{File: lx.file, Line: lx.line, Col: lx.col} }

func (lx *lexer) emit(k tokKind, text string, p diag.Pos) {
	lx.toks = append(lx.toks, token{k, text, p})
}

func (lx *lexer) advance() {
	if lx.src[lx.i] == '\n' {
		lx.line, lx.col = lx.line+1, 1
	} else {
		lx.col++
	}
	lx.i++
}

// take advances while ok holds and returns what it passed.
func (lx *lexer) take(ok func(byte) bool) string {
	start := lx.i
	for lx.i < len(lx.src) && ok(lx.src[lx.i]) {
		lx.advance()
	}
	return lx.src[start:lx.i]
}

func (lx *lexer) step() {
	c, p := lx.src[lx.i], lx.pos()
	switch {
	case c == '\n':
		lx.emit(tNewline, "", p)
		lx.advance()
	case c == ' ' || c == '\t' || c == '\r':
		lx.advance()
	case c == '/' && lx.i+1 < len(lx.src) && lx.src[lx.i+1] == '/':
		lx.take(func(c byte) bool { return c != '\n' })
	case isLetter(c):
		lx.emit(tIdent, lx.take(func(c byte) bool { return isLetter(c) || isDigit(c) }), p)
	case isDigit(c):
		lx.emit(tInt, lx.take(isDigit), p)
	default:
		if k, ok := punct[c]; ok {
			lx.emit(k, string(c), p)
		} else {
			lx.d.Errorf(p, "unexpected character %q", c)
		}
		lx.advance()
	}
}
