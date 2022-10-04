package glisp

import (
	"errors"
	"fmt"
	"regexp"
)

type Parser struct {
	lexer *Lexer
	env   *Environment
}

var UnexpectedEnd error = errors.New("Unexpected end of input")

const SliceDefaultCap = 10

func NewParser(l *Lexer, e *Environment) *Parser {
	return &Parser{lexer: l, env: e}
}

func ParseList(parser *Parser) (Sexp, error) {
	lexer := parser.lexer
	tok, err := lexer.PeekNextToken()
	if err != nil {
		return SexpNull, err
	}
	if tok.typ == TokenEnd {
		_, _ = lexer.GetNextToken()
		return SexpEnd, UnexpectedEnd
	}

	if tok.typ == TokenRParen {
		_, _ = lexer.GetNextToken()
		return SexpNull, nil
	}

	start := &SexpPair{}

	expr, err := ParseExpression(parser)
	if err != nil {
		return SexpNull, err
	}

	start.head = expr

	tok, err = lexer.PeekNextToken()
	if err != nil {
		return SexpNull, err
	}

	if tok.typ == TokenDot {
		// eat up the dot
		_, _ = lexer.GetNextToken()
		expr, err = ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}

		// eat up the end paren
		tok, err = lexer.GetNextToken()
		if err != nil {
			return SexpNull, err
		}
		// make sure it was actually an end paren
		if tok.typ != TokenRParen {
			return SexpNull, errors.New("extra value in dotted pair")
		}
		start.tail = expr
		return start, nil
	}

	expr, err = ParseList(parser)
	if err != nil {
		return start, err
	}
	start.tail = expr

	return start, nil
}

func ParseArray(parser *Parser) (Sexp, error) {
	lexer := parser.lexer
	arr := make([]Sexp, 0, SliceDefaultCap)

	for {
		tok, err := lexer.PeekNextToken()
		if err != nil {
			return SexpEnd, err
		}

		if tok.typ == TokenEnd {
			return SexpEnd, UnexpectedEnd
		}

		if tok.typ == TokenRSquare {
			// pop off the ]
			_, _ = lexer.GetNextToken()
			break
		}

		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		arr = append(arr, expr)
	}

	return SexpArray(arr), nil
}

func ParseHash(parser *Parser) (Sexp, error) {
	lexer := parser.lexer
	arr := make([]Sexp, 0, SliceDefaultCap)

	for {
		tok, err := lexer.PeekNextToken()
		if err != nil {
			return SexpEnd, err
		}
		if tok.typ == TokenEnd {
			return SexpEnd, UnexpectedEnd
		}
		if tok.typ == TokenRCurly {
			// pop off the }
			_, _ = lexer.GetNextToken()
			break
		}

		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		arr = append(arr, expr)
	}

	list := &SexpPair{}
	list.head = parser.env.MakeSymbol("hash")
	list.tail = MakeList(arr)

	return list, nil
}

func ParseExpression(parser *Parser) (Sexp, error) {
	lexer := parser.lexer
	env := parser.env
	tok, err := lexer.GetNextToken()
	if err != nil {
		return SexpEnd, err
	}

	switch tok.typ {
	case TokenLParen:
		return ParseList(parser)
	case TokenLSquare:
		return ParseArray(parser)
	case TokenLCurly:
		return ParseHash(parser)
	case TokenQuote:
		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("quote"), expr}), nil
	case TokenBacktick:
		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("syntax-quote"), expr}), nil
	case TokenLambda:
		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		return makeLambda(env, expr), nil
	case TokenTilde:
		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("unquote"), expr}), nil
	case TokenTildeAt:
		expr, err := ParseExpression(parser)
		if err != nil {
			return SexpNull, err
		}
		return MakeList([]Sexp{env.MakeSymbol("unquote-splicing"), expr}), nil
	case TokenSymbol:
		return env.MakeSymbol(tok.str), nil
	case TokenBool:
		return SexpBool(tok.str == "true"), nil
	case TokenDecimal:
		return NewSexpIntStrWithBase(tok.str, 10)
	case TokenHex:
		return NewSexpIntStrWithBase(tok.str, 16)
	case TokenOct:
		return NewSexpIntStrWithBase(tok.str, 8)
	case TokenBinary:
		return NewSexpIntStrWithBase(tok.str, 2)
	case TokenBinaryStream:
		return NewSexpBytesByHex(tok.str)
	case TokenChar:
		return SexpChar(tok.str[0]), nil
	case TokenString:
		return SexpStr(tok.str), nil
	case TokenFloat:
		return NewSexpFloatStr(tok.str)
	case TokenEnd:
		return SexpEnd, nil
	}
	return SexpNull, fmt.Errorf("Invalid syntax, didn't know what to do with `%v`\n%s", tok, lexer.CurLine())
}

func ParseTokens(env *Environment, lexer *Lexer) ([]Sexp, error) {
	expressions := make([]Sexp, 0, SliceDefaultCap)
	parser := Parser{lexer, env}

	for {
		expr, err := ParseExpression(&parser)
		if err != nil {
			return expressions, err
		}
		if expr == SexpEnd {
			break
		}
		expressions = append(expressions, expr)
	}
	return expressions, nil
}

var lambdaArgument = regexp.MustCompile(`^%(\d*|N)$`)

/*
(fn [& args]
 (let (foldl
         (fn [e acc] (append acc (symbol (concat "%" (string (/ (len acc) 2)))) e)) [(symbol "%N") (len args)] args) EXPR))
*/
func makeLambda(env *Environment, expr Sexp) Sexp {
	/* fix https://clojure.org/guides/learn/functions#_gotcha */
	if !IsList(expr) {
	} else if pair := expr.(*SexpPair); pair.Tail() == SexpNull {
		if IsSymbol(pair.Head()) {
			if name := pair.Head().(SexpSymbol).Name(); lambdaArgument.MatchString(name) {
				expr = pair.Head()
			}
		} else {
			expr = pair.Head()
		}
	}

	foldfn := MakeList([]Sexp{
		env.MakeSymbol("fn"),
		SexpArray{env.MakeSymbol("e"), env.MakeSymbol("acc")},
		MakeList([]Sexp{
			env.MakeSymbol("append"),
			env.MakeSymbol("acc"),
			MakeList([]Sexp{
				env.MakeSymbol("symbol"),
				MakeList([]Sexp{
					env.MakeSymbol("concat"),
					SexpStr("%"),
					MakeList([]Sexp{
						env.MakeSymbol("string"),
						MakeList([]Sexp{
							env.MakeSymbol("/"),
							MakeList([]Sexp{env.MakeSymbol("len"), env.MakeSymbol("acc")}),
							NewSexpInt(2),
						}),
					}),
				}),
			}),
			env.MakeSymbol("e"),
		}),
	})
	letArgs := MakeList([]Sexp{
		env.MakeSymbol("foldl"),
		foldfn,
		SexpArray{MakeList([]Sexp{env.MakeSymbol("symbol"), SexpStr("%N")}), MakeList([]Sexp{env.MakeSymbol("len"), env.MakeSymbol("args")})},
		env.MakeSymbol("args"),
	})
	/* set symbol % when only one argument */
	/* (append letArgs (symbol "%") (cond (empty? args) nil (car args))) */
	letArgs = MakeList([]Sexp{
		env.MakeSymbol("append"),
		letArgs,
		MakeList([]Sexp{env.MakeSymbol("symbol"), SexpStr("%")}),
		MakeList([]Sexp{
			env.MakeSymbol("cond"),
			MakeList([]Sexp{env.MakeSymbol("empty?"), env.MakeSymbol("args")}),
			SexpNull,
			MakeList([]Sexp{env.MakeSymbol("car"), env.MakeSymbol("args")}),
		}),
	})
	letExpr := MakeList([]Sexp{
		env.MakeSymbol("let"),
		letArgs,
		expr,
	})
	lambda := MakeList([]Sexp{
		env.MakeSymbol("fn"),
		SexpArray{env.MakeSymbol("&"), env.MakeSymbol("args")},
		letExpr,
	})
	return lambda
}
