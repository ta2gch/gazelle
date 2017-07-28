package tokenizer

import (
	"io"
	"regexp"
	"strings"
)

// TokenReader interface type is the interface
// for reading string with every token
// Reader is like bufio.Reader but has PeekRune
// which returns a rune without advancing pointer
type TokenReader struct {
	err error
	ru  rune
	sz  int
	rr  io.RuneReader
}

// NewReader creates interal reader from io.RuneReader
func NewTokenReader(r io.RuneReader) *TokenReader {
	b := new(TokenReader)
	b.rr = r
	b.ru, b.sz, b.err = r.ReadRune()
	return b
}

// PeekRune returns a rune without advancing pointer
func (r *TokenReader) PeekRune() (rune, int, error) {
	return r.ru, r.sz, r.err
}

// ReadRune returns a rune with advancing pointer
func (r *TokenReader) ReadRune() (rune, int, error) {
	ru := r.ru
	sz := r.sz
	err := r.err
	r.ru, r.sz, r.err = r.rr.ReadRune()
	return ru, sz, err
}

func concatMatcher(src ...string) *regexp.Regexp {
	return regexp.MustCompile("^(?:" + strings.Join(src, ")$|^(?:") + ")$")
}

var macro = strings.Join([]string{"#(?:[[:digit:]]+[aA]?)?", ",@?", "'", "`"}, "|")
var integer = strings.Join([]string{"[[:digit:]]+", "[+-][[:digit:]]*", "#(?:[bB][+-]?[01]*)?", "#(?:[oO][+-]?[0-7]*)?", "#(?:[xX][+-]?[[:xdigit:]]*)?"}, "|")
var float = strings.Join([]string{"[[:digit:]]+(?:\\.?[[:digit:]]*(?:[eE](?:[-+]?[[:digit:]]*)?)?)?", "[+-](?:[[:digit:]]+(?:\\.?[[:digit:]]*(?:[eE](?:[-+]?[[:digit:]]*)?)?)?)?"}, "|")
var character = strings.Join([]string{"#(?:\\\\[[:alpha:]]*)?", "#(?:\\\\[[:graph:]]?)?"}, "|")
var str = strings.Join([]string{"\"(?:\\\\\"|[^\"])*\"?"}, "|")
var symbol = strings.Join([]string{"[<>/*=?_!$%[\\]^{}~a-zA-Z][<>/*=?_!$%[\\]^{}~0-9a-zA-Z]*", "\\|(?:\\\\\\||[^|])*\\|?"}, "|")
var parentheses = strings.Join([]string{"\\.", "\\(", "\\)"}, "|")

var token = concatMatcher(
	macro,
	integer,
	float,
	character,
	str,
	symbol,
	parentheses)

// ReadToken returns error or string as token
func (r *TokenReader) ReadToken() (string, error) {
	buf := ""
	for {
		if buf == "" {
			p, _, err := r.PeekRune()
			if err != nil {
				return "", err
			}
			if token.MatchString(string(p)) {
				buf = string(p)
			}
		} else {
			p, _, err := r.PeekRune()
			if err != nil {
				return buf, nil
			}
			if !token.MatchString(buf + string(p)) {
				return buf, nil
			}
			buf += string(p)
		}
		r.ReadRune()
	}
}