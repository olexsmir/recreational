package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenInvalid TokenType = iota
	TokenSym
	TokenNum
	TokenStr
	TokenOParen
	TokenCParen
	TokenComma
)

var TokenTypeName = map[TokenType]string{
	TokenInvalid: "TokenInvalid",
	TokenSym:     "TokenSym",
	TokenNum:     "TokenNum",
	TokenStr:     "TokenStr",
	TokenOParen:  "TokenOParen",
	TokenCParen:  "TokenCParen",
	TokenComma:   "TokenComma",
}

type Token struct {
	Type TokenType
	Text []rune
	Loc  Loc
}

type Loc struct {
	Filepath string
	Row, Col int
}

func (l Loc) String() string { return fmt.Sprintf("%s:%d:%d", l.Filepath, l.Row, l.Col) }

type DiagError struct {
	Loc Loc
	Err error
}

func (e *DiagError) Unwrap() error { return e.Err }
func (e *DiagError) Error() string {
	return fmt.Sprintf("%s: ERROR: %s", e.Loc, e.Err)
}

var (
	ErrLexerEOF           = errors.New("Lexer: End of file")
	ErrLexerUnclosedStr   = errors.New("Lexer: Unclosed String")
	ErrLexerInvalidEscape = errors.New("Lexer: Invalid Escape Sequence")
	ErrLexerInvalidToken  = errors.New("Lexer: Invalid Token")
)

type Lexer struct {
	Filepath string
	Content  []rune // TODO: bytes
	Row      int
	Cur      int
	Bol      int
	PeekTok  Token
	PeekErr  error
	PeekFull bool
}

func NewLexer(content []rune, filePath string) Lexer {
	return Lexer{
		Filepath: filePath,
		Content:  content,
	}
}

func (l *Lexer) ChopChar() {
	if l.Cur < len(l.Content) {
		x := l.Content[l.Cur]
		l.Cur += 1
		if x == '\n' {
			l.Bol = l.Cur
			l.Row += 1
		}
	}
}

func (l *Lexer) TrimLeft() {
	for l.Cur < len(l.Content) && unicode.IsSpace(l.Content[l.Cur]) {
		l.ChopChar()
	}
}

func (l Lexer) StartsWith(prefix []rune) bool {
	if l.Cur+len(prefix) > len(l.Content) {
		return false
	}

	for i := range prefix {
		if prefix[i] != l.Content[l.Cur+i] {
			return false
		}
	}

	return true
}

func (l *Lexer) Loc() Loc {
	return Loc{
		Filepath: l.Filepath,
		Row:      l.Row + 1,
		Col:      l.Cur - l.Bol + 1,
	}
}

func (l *Lexer) ChopWhile(p func(rune) bool) (result []rune) {
	for l.Cur < len(l.Content) && p(l.Content[l.Cur]) {
		result = append(result, l.Content[l.Cur])
		l.ChopChar()
	}
	return
}

func (l *Lexer) ChopToken() (Token, error) {
	t := Token{}

	l.TrimLeft()
	for l.Cur < len(l.Content) && l.StartsWith([]rune("--")) {
		for l.Cur < len(l.Content) && l.Content[l.Cur] != '\n' {
			l.ChopChar()
		}
		l.TrimLeft()
	}

	t.Loc = l.Loc()

	if l.Cur >= len(l.Content) {
		return t, ErrLexerEOF
	}

	first := l.Content[l.Cur]

	ps := []rune("(),")
	ts := []TokenType{TokenOParen, TokenCParen, TokenComma}
	for i := range ps {
		if first == ps[i] {
			t.Type = ts[i]
			t.Text = []rune{ps[i]}
			l.ChopChar()
			return t, nil
		}
	}

	if unicode.IsDigit(first) {
		t.Type = TokenNum
		t.Text = l.ChopWhile(unicode.IsDigit)
		return t, nil
	}

	if unicode.IsLetter(first) || first == '_' {
		t.Type = TokenSym
		t.Text = l.ChopWhile(func(x rune) bool {
			return unicode.IsLetter(x) || unicode.IsDigit(x) || x == '_'
		})
		return t, nil
	}

	if first == '"' {
		l.ChopChar()

		t.Type = TokenStr

		for l.Cur < len(l.Content) && l.Content[l.Cur] != '"' {
			if l.Content[l.Cur] == '\\' {
				l.ChopChar()
				if l.Cur >= len(l.Content) {
					return t, &DiagError{
						Loc: l.Loc(),
						Err: ErrLexerUnclosedStr,
					}
				}
				if l.Content[l.Cur] == '"' {
					t.Text = append(t.Text, '"')
					l.ChopChar()
				} else {
					loc := l.Loc()
					l.ChopChar()
					return t, &DiagError{
						Loc: loc,
						Err: fmt.Errorf("%w: %c", ErrLexerInvalidEscape, l.Content[l.Cur]),
					}
				}
			} else {
				t.Text = append(t.Text, l.Content[l.Cur])
				l.ChopChar()
			}
		}

		if l.Cur >= len(l.Content) {
			return t, &DiagError{
				Loc: l.Loc(),
				Err: ErrLexerUnclosedStr,
			}
		}

		l.ChopChar()
		return t, nil
	}

	l.ChopChar()
	return t, &DiagError{
		Loc: l.Loc(),
		Err: fmt.Errorf("%w: %c", ErrLexerInvalidToken, first),
	}
}

func (l *Lexer) Peek() (Token, error) {
	if !l.PeekFull {
		l.PeekTok, l.PeekErr = l.ChopToken()
		l.PeekFull = true
	}
	return l.PeekTok, l.PeekErr
}

func (l *Lexer) Next() (Token, error) {
	if l.PeekFull {
		l.PeekFull = false
	} else {
		l.PeekTok, l.PeekErr = l.ChopToken()
	}
	return l.PeekTok, l.PeekErr
}

type ExprType int

const (
	ExprVoid ExprType = iota
	ExprInt
	ExprStr
	ExprVar
	ExprFuncall
)

type Expr struct {
	Type ExprType
	Loc  Loc

	AsInt     int
	AsStr     string
	AsVar     string
	AsFuncall Funcall
}

type Funcall struct {
	Name string
	Args []Expr
}

func (funcall *Funcall) String() string {
	var result strings.Builder
	fmt.Fprintf(&result, "%s(", funcall.Name)
	for i, arg := range funcall.Args {
		if i > 0 {
			fmt.Fprintf(&result, ", ")
		}
		fmt.Fprintf(&result, "%s", arg.String())
	}
	fmt.Fprintf(&result, ")")
	return result.String()
}

func (expr *Expr) Dump(level int) {
	for i := 0; i < level; i += 1 {
		fmt.Printf("  ")
	}

	switch expr.Type {
	case ExprVoid:
		fmt.Printf("Void\n")
	case ExprInt:
		fmt.Printf("Int: %d\n", expr.AsInt)
	case ExprStr:
		fmt.Printf("Str: \"%s\"\n", expr.AsStr) // TODO: escape strings
	case ExprVar:
		fmt.Printf("Var: %s\n", expr.AsVar)
	case ExprFuncall:
		fmt.Printf("Funcall: %s\n", expr.AsFuncall.Name)
		for _, arg := range expr.AsFuncall.Args {
			arg.Dump(level + 1)
		}
	}
}

func (expr *Expr) String() string {
	switch expr.Type {
	case ExprVoid:
		return ""
	case ExprInt:
		return fmt.Sprintf("%d", expr.AsInt)
	case ExprStr:
		return fmt.Sprintf("\"%s\"", expr.AsStr) // TODO: escape string
	case ExprVar:
		return expr.AsVar
	case ExprFuncall:
		return expr.AsFuncall.String()
	default:
		panic("unreachable")
	}
}

var (
	ErrParserUnexpectedToken = errors.New("Parser: Unexpected Token")
	ErrParserUnclosedFuncall = errors.New("Parser: Unclosed Funcall")
)

func ParseExpr(l *Lexer) (Expr, error) {
	t, err := l.Next()
	if err != nil {
		return Expr{}, err
	}

	switch t.Type {
	case TokenSym:
		oparen, err := l.Peek()
		if err != nil || oparen.Type != TokenOParen {
			return Expr{
				Type:  ExprVar,
				Loc:   t.Loc,
				AsVar: string(t.Text),
			}, nil
		} else {
			l.Next()
			cparen, err := l.Peek()

			if err == nil && cparen.Type == TokenCParen {
				l.Next()
				return Expr{
					Type:      ExprFuncall,
					Loc:       t.Loc,
					AsFuncall: Funcall{Name: string(t.Text)},
				}, nil
			}

			arg, err := ParseExpr(l)
			if err != nil {
				return Expr{}, err
			}

			args := []Expr{arg}

			comma, err := l.Next()
			for err == nil && comma.Type == TokenComma {
				arg, err = ParseExpr(l)
				if err != nil {
					return Expr{}, err
				}
				args = append(args, arg)
				comma, err = l.Next()
			}

			if err == ErrLexerEOF || comma.Type != TokenCParen {
				return Expr{}, &DiagError{Loc: comma.Loc, Err: ErrParserUnclosedFuncall}
			}

			return Expr{
				Type: ExprFuncall,
				Loc:  t.Loc,
				AsFuncall: Funcall{
					Name: string(t.Text),
					Args: args,
				},
			}, nil
		}

	case TokenNum:
		s := string(t.Text)
		x, err := strconv.Atoi(s)
		if err != nil {
			return Expr{}, &DiagError{Loc: t.Loc, Err: err}
		}
		return Expr{
			Type:  ExprInt,
			Loc:   t.Loc,
			AsInt: x,
		}, nil

	case TokenStr:
		return Expr{
			Type:  ExprStr,
			Loc:   t.Loc,
			AsStr: string(t.Text),
		}, nil
	}

	return Expr{}, &DiagError{
		Loc: t.Loc,
		Err: fmt.Errorf("%w: %s", ErrParserUnexpectedToken, TokenTypeName[t.Type]),
	}
}

func ParseExprs(l *Lexer) ([]Expr, error) {
	exprs := []Expr{}
	for {
		expr, err := ParseExpr(l)
		if err != nil {
			if errors.Is(err, ErrLexerEOF) {
				err = nil
			}
			return exprs, err
		}
		exprs = append(exprs, expr)
	}
}

type (
	Func = func(context *EvalContext, args []Expr) (Expr, error)

	EvalContext struct{ Scopes []EvalScope }
	EvalScope   struct {
		Vars  map[string]Expr
		Funcs map[string]Func
	}
)

func (e EvalContext) LookupVar(name string) (Expr, bool) {
	scopes := e.Scopes
	for len(scopes) > 0 {
		n := len(scopes)
		varr, ok := scopes[n-1].Vars[name]
		if ok {
			return varr, true
		}
		scopes = scopes[:n-1]
	}
	return Expr{}, false
}

func (e EvalContext) LookupFunc(name string) (Func, bool) {
	scopes := e.Scopes
	for len(scopes) > 0 {
		n := len(scopes)
		fun, ok := scopes[n-1].Funcs[name]
		if ok {
			return fun, true
		}
		scopes = scopes[:n-1]
	}
	return nil, false
}

func (e *EvalContext) PushScope(scope EvalScope) { e.Scopes = append(e.Scopes, scope) }
func (e *EvalContext) PopScope()                 { e.Scopes = e.Scopes[0 : len(e.Scopes)-1] }

func (e *EvalContext) TopScope() *EvalScope {
	length := len(e.Scopes)
	if length <= 0 {
		panic("No scopes found")
	}
	return &e.Scopes[length-1]
}

var (
	ErrRuntimeUnknownVar = errors.New("unknown runtime variable")
	ErrRuntimeUnknownFun = errors.New("unknown runtime function")
)

func (e *EvalContext) EvalExpr(expr Expr) (Expr, error) {
	switch expr.Type {
	case ExprVoid, ExprInt, ExprStr:
		return expr, nil

	case ExprVar:
		v, ok := e.LookupVar(expr.AsVar)
		if !ok {
			return Expr{}, fmt.Errorf("%s: ERROR: %w: %s", expr.Loc, ErrRuntimeUnknownVar, expr.AsVar)
		}
		return e.EvalExpr(v)

	case ExprFuncall:
		fn, ok := e.LookupFunc(expr.AsFuncall.Name)
		if !ok {
			return Expr{}, fmt.Errorf("%s: ERROR: %w: %s", expr.Loc, ErrRuntimeUnknownFun, expr.AsFuncall.Name)
		}
		return fn(e, expr.AsFuncall.Args)

	default:
		panic("unreachable")
	}
}
