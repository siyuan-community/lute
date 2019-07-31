// Lute - A structured markdown engine.
// Copyright (C) 2019-present, b3log.org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lute

import (
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	items   []items
	lastPos int
	line    int
}

type scanner struct {
	input   string // the string being scanned
	pos     int    // current position in the input
	start   int    // start position of this item
	width   int    // width of last rune read from input
	lastPos int    // position of most recent item returned by nextItem
	items   items  // scanned items
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() *item {
	if len(l.items) <= l.line || len(l.items[l.line]) <= l.lastPos {
		return &item{typ: itemEOF}
	}

	item := l.items[l.line][l.lastPos]
	if itemNewline == item.typ {
		l.line++
		l.lastPos = 0
	} else {
		l.lastPos++
	}

	return item
}

// lex creates a new lexer for the input string.
func lex(name, input string) *lexer {
	ret := &lexer{items: []items{}}
	if "" == input {
		ret.items = append(ret.items, items{})
		ret.items[ret.line] = append(ret.items[ret.line], &item{typ: itemEOF})

		return ret
	}

	var lines []string
	var line string
	length := len(input)
	for i, c := range input {
		char := string(c)
		line += char
		if "\n" == char || i == length-1 {
			lines = append(lines, line)
			line = ""
		}
	}

	for _, line := range lines {
		if "" == line {
			ret.items = append(ret.items, items{{typ: itemNewline, val: "\n"}})

			continue
		}

		s := &scanner{
			input: line,
			items: items{},
		}
		s.run()

		ret.items = append(ret.items, s.items)
	}

	lastLine := len(ret.items) - 1
	ret.items[lastLine] = append(ret.items[lastLine], &item{typ: itemEOF})

	return ret
}

func (s *scanner) run() {
	for {
		r := s.next()
		switch {
		case '`' == r:
			s.newItem(itemBacktick)
		case '~' == r:
			s.newItem(itemTilde)
		case '!' == r:
			s.newItem(itemBang)
		case '#' == r:
			s.newItem(itemCrosshatch)
		case '*' == r:
			s.newItem(itemAsterisk)
		case '(' == r:
			s.newItem(itemOpenParen)
		case ')' == r:
			s.newItem(itemCloseParen)
		case '-' == r:
			s.newItem(itemHyphen)
		case '_' == r:
			s.newItem(itemUnderscore)
		case '+' == r:
			s.newItem(itemPlus)
		case '=' == r:
			s.newItem(itemEqual)
		case '\t' == r:
			s.newItem(itemTab)
		case '[' == r:
			s.newItem(itemOpenBracket)
		case ']' == r:
			s.newItem(itemCloseBracket)
		case '"' == r:
			s.newItem(itemDoublequote)
		case '\'' == r:
			s.newItem(itemSinglequote)
		case '<' == r:
			s.newItem(itemLess)
		case '>' == r:
			s.newItem(itemGreater)
		case ' ' == r:
			s.newItem(itemSpace)
		case '\n' == r:
			s.newItem(itemNewline)
		case '\\' == r:
			s.newItem(itemBackslash)
		case '/' == r:
			s.newItem(itemSlash)
		case '.' == r:
			s.newItem(itemDot)
		case ':' == r:
			s.newItem(itemColon)
		case '?' == r:
			s.newItem(itemQuestion)
		case '&' == r:
			s.newItem(itemAmpersand)
		case ';' == r:
			s.newItem(itemSemicolon)
		case unicode.IsSymbol(r), unicode.IsPunct(r), unicode.IsSpace(r):
			s.newItem(itemStr)
		case unicode.IsControl(r):
			s.newItem(itemControl)
		case end == r:
			return
		default:
		str:
			for {
				switch {
				case unicode.IsLetter(r), unicode.IsNumber(r):
					r = s.next()
				default:
					s.backup()
					s.newItem(itemStr)
					break str
				}
			}
		}
	}
}

// next returns the next rune in the input.
func (s *scanner) next() rune {
	if int(s.pos) >= len(s.input) {
		s.width = 0

		return end
	}

	r, w := utf8.DecodeRuneInString(s.input[s.pos:])
	s.width = w
	s.pos += s.width

	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (s *scanner) backup() {
	s.pos -= s.width
}

// newItem creates an item with the specified item type.
func (s *scanner) newItem(t itemType) {
	s.items = append(s.items, &item{t, s.start, s.input[s.start:s.pos], 1})
	s.start = s.pos
}
