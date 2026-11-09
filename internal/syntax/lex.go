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
	tString
	tArrow
)

var tokNames = [...]string{"end of file", "end of line", "name", "number", "'{'", "'}'", "'('", "')'", "':'", "','", "string", "'->'"}

type token struct {
	kind tokKind
	text string
	pos  diag.Pos
}

func (t token) describe() string {
	switch t.kind {
	case tIdent, tInt:
		return fmt.Sprintf("'%s'", t.text)
	case tString:
		return fmt.Sprintf("%q", t.text)
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
	case c == '"':
		lx.str(p)
	case c == '-' && lx.i+1 < len(lx.src) && lx.src[lx.i+1] == '>':
		lx.emit(tArrow, "->", p)
		lx.advance()
		lx.advance()
	default:
		if k, ok := punct[c]; ok {
			lx.emit(k, string(c), p)
		} else {
			lx.d.Errorf(p, "unexpected character %q", c)
		}
		lx.advance()
	}
}

// str lexes a one-line string with \" and \\ escapes.
func (lx *lexer) str(p diag.Pos) {
	lx.advance()
	var buf []byte
	for lx.i < len(lx.src) && lx.src[lx.i] != '\n' {
		c := lx.src[lx.i]
		lx.advance()
		switch {
		case c == '"':
			lx.emit(tString, string(buf), p)
			return
		case c == '\\' && lx.i < len(lx.src) && (lx.src[lx.i] == '"' || lx.src[lx.i] == '\\'):
			buf = append(buf, lx.src[lx.i])
			lx.advance()
		default:
			buf = append(buf, c)
		}
	}
	lx.d.Errorf(p, "this string is never closed")
	lx.emit(tString, string(buf), p)
}
