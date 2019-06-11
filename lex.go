// Lute - A structural markdown engine.
// Copyright (C) 2019-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package lute

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token returned from the scanner.
type item struct {
	typ  itemType // the type of this item
	pos  int      // the starting position, in bytes, of this item in the input string
	val  string   // the value of this item, aka lexeme
	line int      // the line number at the start of this item
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}

	return fmt.Sprintf("%q", i.val)
}

// https://github.github.com/gfm/#whitespace-character
func (i item) isWhitespace() bool {
	return itemSpace == i.typ || itemTab == i.typ || itemNewline == i.typ // TODO(D): line tabulation (U+000B), form feed (U+000C), or carriage return (U+000D)
}

// itemType identifies the type of lex items.
type itemType int

// Make the types pretty print.
var itemName = map[itemType]string{
	itemEOF:          "EOF",
	itemStr:          "str",
	itemBacktick:     "`",
	itemTilde:        "~",
	itemExclamation:  "!",
	itemCrosshatch:   "#",
	itemAsterisk:     "*",
	itemOpenParen:    "(",
	itemCloseParen:   ")",
	itemHyphen:       "-",
	itemPlus:         "+",
	itemTab:          "tab",
	itemOpenBracket:  "[",
	itemCloseBracket: "]",
	itemDoublequote:  "\"",
	itemSinglequote:  "'",
	itemGreater:      ">",
	itemSpace:        "space",
	itemNewline:      "newline",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}

	return s
}

const (
	itemEOF          itemType = iota // EOF
	itemStr                          // plain text
	itemBacktick                     // `
	itemTilde                        // ~
	itemExclamation                  // !
	itemCrosshatch                   // #
	itemAsterisk                     // *
	itemOpenParen                    // (
	itemCloseParen                   // )
	itemHyphen                       // -
	itemPlus                         // +
	itemTab                          // \t
	itemOpenBracket                  // [
	itemCloseBracket                 // ]
	itemDoublequote                  // "
	itemSinglequote                  // '
	itemGreater                      // >
	itemSpace                        // space
	itemNewline                      // \n
)

const (
	end = -1
)

type lexer struct {
	items   [][]item
	lastPos int
	line    int
}

type scanner struct {
	input   string // the string being scanned
	pos     int    // current position in the input
	start   int    // start position of this item
	width   int    // width of last rune read from input
	lastPos int    // position of most recent item returned by nextItem
	items   []item // scanned items
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
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
	ret := &lexer{items: [][]item{}}
	if "" == input {
		ret.items = append(ret.items, []item{})
		ret.items[ret.line] = append(ret.items[ret.line], item{typ: itemEOF})

		return ret
	}

	lines := strings.Split(input, "\n")
	multipleLines := 1 < len(lines)
	for _, line := range lines {
		if "" == line {
			break
		}

		if multipleLines {
			line += "\n"
		}
		s := &scanner{
			input: line,
			items: []item{},
		}
		s.run()

		ret.items = append(ret.items, s.items)
	}

	ret.items[ret.line] = append(ret.items[ret.line], item{typ: itemEOF})

	return ret
}

func (s *scanner) run() {
	for {
		r := s.next()
		switch {
		case '`' == r:
			s.newItem(itemBacktick)
		case '!' == r:
			s.newItem(itemExclamation)
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
		case '+' == r:
			s.newItem(itemPlus)
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
		case '>' == r:
			s.newItem(itemGreater)
		case ' ' == r:
			s.newItem(itemSpace)
		case '\n' == r:
			s.newItem(itemNewline)
		case end == r:
			return
		default:
		str:
			for {
				switch {
				case unicode.IsLetter(r), unicode.IsNumber(r):
					// absorb
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
	s.items = append(s.items, item{t, s.start, s.input[s.start:s.pos], 1})
	s.start = s.pos
}
