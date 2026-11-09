package spec

import "fmt"

type tokKind int

const (
	tEOF tokKind = iota
	tNewline
	tIdent
	tString
	tInt
	tLBrace
	tRBrace
	tLParen
	tRParen
	tLBrack
	tRBrack
	tColon
	tComma
	tDot
	tSlash
	tOp
)

var tokNames = [...]string{"end of file", "end of line", "identifier", "string", "number",
	"'{'", "'}'", "'('", "')'", "'['", "']'", "':'", "','", "'.'", "'/'", "operator"}

type Token struct {
	Kind tokKind
	Text string // unquoted value for strings
	Pos  Pos
}

func (t Token) describe() string {
	switch t.Kind {
	case tIdent, tInt, tOp:
		return fmt.Sprintf("'%s'", t.Text)
	case tString:
		return fmt.Sprintf("%q", t.Text)
	}
	return tokNames[t.Kind]
}

// width is the token's length in the source, used to point just past it.
func (t Token) width() int {
	if t.Kind == tString {
		return len(t.Text) + 2
	}
	return max(1, len(t.Text))
}

func isLetter(c byte) bool { return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }

func lex(src string, d *Diags) []Token {
	var toks []Token
	line, col, i := 1, 1, 0
	adv := func() {
		if src[i] == '\n' {
			line, col = line+1, 1
		} else {
			col++
		}
		i++
	}
	emit := func(k tokKind, text string, p Pos) { toks = append(toks, Token{k, text, p}) }
	single := map[byte]tokKind{'{': tLBrace, '}': tRBrace, '(': tLParen, ')': tRParen,
		'[': tLBrack, ']': tRBrack, ':': tColon, ',': tComma, '.': tDot}

	for i < len(src) {
		c := src[i]
		p := Pos{line, col}
		switch {
		case c == '\n':
			emit(tNewline, "\n", p)
			adv()
		case c == ' ' || c == '\t' || c == '\r':
			adv()
		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			for i < len(src) && src[i] != '\n' {
				adv()
			}
		case c == '/':
			emit(tSlash, "/", p)
			adv()
		case isLetter(c):
			s := i
			for i < len(src) && (isLetter(src[i]) || isDigit(src[i])) {
				adv()
			}
			emit(tIdent, src[s:i], p)
		case isDigit(c):
			s := i
			for i < len(src) && isDigit(src[i]) {
				adv()
			}
			emit(tInt, src[s:i], p)
		case c == '"':
			adv()
			var buf []byte
			closed := false
			for i < len(src) && src[i] != '\n' {
				if src[i] == '"' {
					adv()
					closed = true
					break
				}
				if src[i] == '\\' && i+1 < len(src) && (src[i+1] == '"' || src[i+1] == '\\') {
					buf = append(buf, src[i+1])
					adv()
					adv()
					continue
				}
				buf = append(buf, src[i])
				adv()
			}
			if !closed {
				d.Errorf(p, "unterminated string")
			}
			emit(tString, string(buf), p)
		case c == '<' || c == '>' || c == '=' || c == '!':
			if i+1 < len(src) && src[i+1] == '=' {
				emit(tOp, src[i:i+2], p)
				adv()
				adv()
			} else if c == '<' || c == '>' {
				emit(tOp, src[i:i+1], p)
				adv()
			} else {
				d.Errorf(p, "unexpected '%c'", c)
				if c == '=' {
					d.Hint("comparisons use '=='")
				}
				adv()
			}
		default:
			if k, ok := single[c]; ok {
				emit(k, string(c), p)
			} else {
				d.Errorf(p, "unexpected character %q", c)
			}
			adv()
		}
	}
	emit(tEOF, "", Pos{line, col})
	return toks
}
