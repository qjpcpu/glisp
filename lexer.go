package glisp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"unicode/utf8"
)

type TokenType int

const (
	TokenLParen TokenType = iota
	TokenRParen
	TokenLSquare
	TokenRSquare
	TokenLCurly
	TokenRCurly
	TokenDot
	TokenQuote
	TokenSharpQuote
	TokenBacktick
	TokenLambda
	TokenTilde
	TokenTildeAt
	TokenSymbol
	TokenBool
	TokenDecimal
	TokenHex
	TokenOct
	TokenBinary
	TokenBinaryStream
	TokenFloat
	TokenChar
	TokenString
	TokenEnd
)

type Token struct {
	typ TokenType
	str string
}

func (t Token) String() string {
	switch t.typ {
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenLSquare:
		return "["
	case TokenRSquare:
		return "]"
	case TokenLCurly:
		return "{"
	case TokenRCurly:
		return "}"
	case TokenDot:
		return "."
	case TokenQuote:
		return "'"
	case TokenSharpQuote:
		return "#'"
	case TokenBacktick:
		return "`"
	case TokenTilde:
		return "~"
	case TokenTildeAt:
		return "~@"
	case TokenHex:
		return "0x" + t.str
	case TokenOct:
		return "0o" + t.str
	case TokenBinary:
		return "0b" + t.str
	case TokenBinaryStream:
		return "0B" + t.str
	case TokenChar:
		quoted := strconv.Quote(t.str)
		return "#" + quoted[1:len(quoted)-1]
	case TokenLambda:
		return "#"
	}
	return t.str
}

type LexerState int

const (
	LexerNormal LexerState = iota
	LexerComment
	LexerStrLit
	LexerRawStrLit
	LexerStrEscaped
	LexerUnquote
	LexerSharp
)

type Lexer struct {
	state    LexerState
	tokens   []Token
	buffer   *bytes.Buffer
	stream   RuneReader
	finished bool
}

var (
	BoolRegex         = regexp.MustCompile("^(true|false)$")
	DecimalRegex      = regexp.MustCompile("^-?[0-9]+$")
	HexRegex          = regexp.MustCompile("^0x[0-9a-fA-F]+$")
	OctRegex          = regexp.MustCompile("^0o[0-7]+$")
	BinaryRegex       = regexp.MustCompile("^0b[01]+$")
	BinaryStreamRegex = regexp.MustCompile("^0B[0-9a-z]+$")
	SymbolRegex       = regexp.MustCompile("^[^'#]+$")
	CharRegex         = regexp.MustCompile("^#\\\\?.$")
	FloatRegex        = regexp.MustCompile("^-?([0-9]+\\.[0-9]*)|(\\.[0-9]+)|([0-9]+(\\.[0-9]*)?[eE](-?[0-9]+))$")
)

func StringToRunes(str string) []rune {
	b := []byte(str)
	runes := make([]rune, 0)

	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		runes = append(runes, r)
		b = b[size:]
	}
	return runes
}

func EscapeChar(char rune) (rune, error) {
	switch char {
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 'a':
		return '\a', nil
	case 't':
		return '\t', nil
	case '\\':
		return '\\', nil
	case '"':
		return '"', nil
	case '\'':
		return '\'', nil
	case '#':
		return '#', nil
	}
	return ' ', errors.New("invalid escape sequence")
}

func DecodeChar(atom string) (string, error) {
	switch atom {
	case `#\'`:
		atom = `#'`
	case `#\(`:
		atom = `#(`
	case `#\)`:
		atom = `#)`
	case `#\]`:
		atom = `#]`
	case `#\[`:
		atom = `#[`
	case `#\#`:
		atom = `##`
	case `#\~`:
		atom = `#~`
	case "#\\`":
		atom = "#`"
	case `#\{`:
		atom = "#{"
	case `#\}`:
		atom = "#}"
	case `#\;`:
		atom = "#;"
	case `#\\`:
		atom = `#\`
	}
	runes := StringToRunes(atom)
	if len(runes) == 3 {
		char, err := EscapeChar(runes[2])
		return string(char), err
	}

	if len(runes) == 2 {
		return string(runes[1:2]), nil
	}
	return "", errors.New("not a char literal")
}

func DecodeAtom(atom string) (Token, error) {
	if atom == "." {
		return Token{TokenDot, ""}, nil
	}
	if BoolRegex.MatchString(atom) {
		return Token{TokenBool, atom}, nil
	}
	if DecimalRegex.MatchString(atom) {
		return Token{TokenDecimal, atom}, nil
	}
	if HexRegex.MatchString(atom) {
		return Token{TokenHex, atom[2:]}, nil
	}
	if OctRegex.MatchString(atom) {
		return Token{TokenOct, atom[2:]}, nil
	}
	if BinaryRegex.MatchString(atom) {
		return Token{TokenBinary, atom[2:]}, nil
	}
	if BinaryStreamRegex.MatchString(atom) {
		return Token{TokenBinaryStream, atom[2:]}, nil
	}
	if FloatRegex.MatchString(atom) {
		return Token{TokenFloat, atom}, nil
	}
	if SymbolRegex.MatchString(atom) {
		return Token{TokenSymbol, atom}, nil
	}
	if CharRegex.MatchString(atom) {
		char, err := DecodeChar(atom)
		if err != nil {
			return Token{}, err
		}
		return Token{TokenChar, char}, nil
	}

	return Token{}, fmt.Errorf("Unrecognized atom `%s`", atom)
}

func (lexer *Lexer) dumpBuffer() error {
	if lexer.buffer.Len() <= 0 {
		return nil
	}

	tok, err := DecodeAtom(lexer.buffer.String())
	if err != nil {
		return err
	}

	lexer.buffer.Reset()
	lexer.tokens = append(lexer.tokens, tok)
	return nil
}

func (lexer *Lexer) dumpString() {
	str := lexer.buffer.String()
	lexer.buffer.Reset()
	lexer.tokens = append(lexer.tokens, Token{TokenString, str})
}

func DecodeBrace(brace rune) Token {
	switch brace {
	case '(':
		return Token{TokenLParen, ""}
	case ')':
		return Token{TokenRParen, ""}
	case '[':
		return Token{TokenLSquare, ""}
	case ']':
		return Token{TokenRSquare, ""}
	case '{':
		return Token{TokenLCurly, ""}
	case '}':
		return Token{TokenRCurly, ""}
	}
	return Token{TokenEnd, ""}
}

func (lexer *Lexer) LexNextRune(r rune) error {
	if lexer.state == LexerComment {
		if r == '\n' {
			lexer.state = LexerNormal
		}
		return nil
	}
	if lexer.state == LexerStrLit {
		if r == '\\' {
			lexer.state = LexerStrEscaped
			return nil
		}
		if r == '"' {
			lexer.dumpString()
			lexer.state = LexerNormal
			return nil
		}
		lexer.buffer.WriteRune(r)
		return nil
	}
	if lexer.state == LexerRawStrLit {
		if r == '`' {
			lexer.dumpString()
			lexer.state = LexerNormal
			return nil
		}
		lexer.buffer.WriteRune(r)
		return nil
	}
	if lexer.state == LexerStrEscaped {
		char, err := EscapeChar(r)
		if err != nil {
			return err
		}
		lexer.buffer.WriteRune(char)
		lexer.state = LexerStrLit
		return nil
	}
	if lexer.state == LexerUnquote {
		if r == '@' {
			lexer.tokens = append(
				lexer.tokens, Token{TokenTildeAt, ""})
		} else {
			lexer.tokens = append(
				lexer.tokens, Token{TokenTilde, ""})
			lexer.buffer.WriteRune(r)
		}
		lexer.state = LexerNormal
		return nil
	}
	if lexer.state == LexerSharp {
		if r == '\'' && lexer.buffer.Len() == 1 {
			lexer.buffer.Reset()
			lexer.tokens = append(lexer.tokens, Token{TokenSharpQuote, ""})
			lexer.state = LexerNormal
			return nil
		} else if r == '`' && lexer.buffer.Len() == 1 {
			lexer.buffer.Reset()
			lexer.state = LexerRawStrLit
			return nil
		} else if r == '\\' && lexer.buffer.Len() == 1 {
			_, err := lexer.buffer.WriteRune(r)
			return err
		} else if lexer.buffer.String() == `#\` {
			lexer.state = LexerNormal
			_, err := lexer.buffer.WriteRune(r)
			return err
		} else if r == '(' && lexer.buffer.Len() == 1 {
			/* lambda */
			lexer.state = LexerNormal
			lexer.buffer.Reset()
			lexer.tokens = append(lexer.tokens, Token{TokenLambda, ""})
		}
	}

	if r == '"' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected quote")
		}
		lexer.state = LexerStrLit
		return nil
	}

	if r == ';' {
		lexer.state = LexerComment
		return nil
	}

	if r == '#' && lexer.buffer.Len() == 0 {
		lexer.state = LexerSharp
		_, err := lexer.buffer.WriteRune(r)
		return err
	}
	if r == '\'' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected quote")
		}
		lexer.tokens = append(lexer.tokens, Token{TokenQuote, ""})
		return nil
	}

	if r == '`' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected backtick")
		}
		lexer.tokens = append(lexer.tokens, Token{TokenBacktick, ""})
		return nil
	}

	if r == '~' {
		if lexer.buffer.Len() > 0 {
			return errors.New("Unexpected tilde")
		}
		lexer.state = LexerUnquote
		return nil
	}

	if r == '(' || r == ')' || r == '[' || r == ']' || r == '{' || r == '}' {
		err := lexer.dumpBuffer()
		if err != nil {
			return err
		}
		lexer.tokens = append(lexer.tokens, DecodeBrace(r))
		return nil
	}
	if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
		err := lexer.dumpBuffer()
		if err != nil {
			return err
		}
		return nil
	}

	_, err := lexer.buffer.WriteRune(r)
	if err != nil {
		return err
	}
	return nil
}

func (lexer *Lexer) PeekNextToken() (Token, error) {
	if lexer.finished {
		return Token{TokenEnd, ""}, nil
	}
	for len(lexer.tokens) == 0 {
		r, _, err := lexer.stream.ReadRune()
		if err != nil {
			lexer.finished = true
			if lexer.buffer.Len() > 0 {
				lexer.dumpBuffer()
				return lexer.tokens[0], nil
			}
			return Token{TokenEnd, ""}, nil
		}

		err = lexer.LexNextRune(r)
		if err != nil {
			return Token{TokenEnd, ""}, err
		}
	}

	tok := lexer.tokens[0]
	return tok, nil
}

func (lexer *Lexer) GetNextToken() (Token, error) {
	tok, err := lexer.PeekNextToken()
	if err != nil || tok.typ == TokenEnd {
		return Token{TokenEnd, ""}, err
	}
	lexer.tokens = lexer.tokens[1:]
	return tok, nil
}

func NewLexerFromStream(stream io.RuneReader) *Lexer {
	return &Lexer{
		tokens:   make([]Token, 0, 10),
		buffer:   new(bytes.Buffer),
		state:    LexerNormal,
		stream:   NewRuneReader(stream),
		finished: false,
	}
}

func (lexer *Lexer) Linenum() int {
	n, _ := lexer.stream.Offset()
	return n
}

func (lexer *Lexer) LineOffset() int {
	_, n := lexer.stream.Offset()
	return n
}

func (lexer *Lexer) CurLine() string {
	return lexer.stream.CurLine()
}

type RuneReader interface {
	Offset() (line int, offset int)
	CurLine() string
	ReadRune() (r rune, size int, err error)
}

func NewRuneReader(r io.RuneReader) RuneReader {
	return &runeReader{r: r, linenum: 1, curline: new(bytes.Buffer)}
}

type runeReader struct {
	r               io.RuneReader
	linenum, offset int
	curline         *bytes.Buffer
	newline         bool
}

func (reader *runeReader) Offset() (int, int) {
	return reader.linenum, reader.offset
}

func (reader *runeReader) CurLine() string {
	return reader.curline.String()
}

func (reader *runeReader) ReadRune() (r rune, size int, err error) {
	r, size, err = reader.r.ReadRune()
	if reader.newline {
		reader.linenum++
		reader.offset = 0
		reader.curline.Reset()
	}
	if err == nil {
		reader.offset++
		reader.newline = r == '\n'
		reader.curline.WriteRune(r)
	}
	return
}
