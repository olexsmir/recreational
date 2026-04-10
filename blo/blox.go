package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var escope = EvalScope{
	Vars: map[string]Expr{
		"my_name": {
			Type:  ExprStr,
			AsStr: "Olex",
		},
	},
	Funcs: map[string]Func{
		"let": func(ctx *EvalContext, args []Expr) (Expr, error) {
			if len(args) != 2 {
				return Expr{}, errors.New("let() expects two arguments")
			}

			if args[0].Type != ExprVar {
				return Expr{}, errors.New("First argument of let() has to be variable name")
			}

			name := args[0].AsVar
			value, err := ctx.EvalExpr(args[1])
			if err != nil {
				return Expr{}, err
			}

			ctx.TopScope().Vars[name] = value
			return Expr{}, nil
		},

		"define": func(ctx *EvalContext, args []Expr) (Expr, error) {
			if len(args) < 2 {
				return Expr{}, errors.New("define() expects at least 2 arguments")
			}

			if args[0].Type != ExprVar {
				return Expr{}, errors.New("define(): first argument must be the name of the function")
			}

			funName := args[0].AsVar
			if args[1].Type != ExprFuncall || args[1].AsFuncall.Name != "args" {
				return Expr{}, errors.New("define(): second argument must be the argument list")
			}

			funArgs := args[1].AsFuncall.Args
			for _, funArg := range funArgs {
				if funArg.Type != ExprVar {
					return Expr{}, errors.New("define(): argument list must consist of only variable names")
				}
			}

			ctx.TopScope().Funcs[funName] = func(context *EvalContext, callArgs []Expr) (Expr, error) {
				scope := EvalScope{
					Vars:  map[string]Expr{},
					Funcs: map[string]Func{},
				}

				if len(callArgs) != len(funArgs) {
					return Expr{}, errors.New(fmt.Sprintf("%s(): expected %d arguments but provided %d", funName, len(funArgs), len(args)))
				}

				for index := range callArgs {
					scope.Vars[funArgs[index].AsVar] = callArgs[index]
				}

				context.PushScope(scope)
				for _, stmt := range args[2:] {
					_, err := context.EvalExpr(stmt)
					if err != nil {
						return Expr{}, err
					}
				}
				context.PopScope()

				return Expr{}, nil
			}

			return Expr{}, nil
		},

		"say": func(context *EvalContext, args []Expr) (Expr, error) {
			for _, arg := range args {
				val, err := context.EvalExpr(arg)
				if err != nil {
					return Expr{}, err
				}

				switch val.Type {
				case ExprStr:
					fmt.Printf("%s", val.AsStr)

				case ExprInt:
					fmt.Printf("%d", val.AsInt)
				default:
					return Expr{}, errors.New("say() expects its arguments to be strings or numbers")
				}
			}
			fmt.Printf("\n")
			return Expr{}, nil
		},

		"http": func(context *EvalContext, args []Expr) (Expr, error) {
			var url strings.Builder
			for _, arg := range args {
				val, err := context.EvalExpr(arg)
				if err != nil {
					return Expr{}, err
				}
				if val.Type != ExprStr {
					return Expr{}, errors.New("http() expects its arguments to be strings")
				}
				fmt.Fprint(&url, val.AsStr)
			}

			resp, err := http.Get(url.String())
			if err != nil {
				return Expr{}, err
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return Expr{}, err
			}

			return Expr{
				Type:  ExprStr,
				AsStr: string(body),
			}, nil
		},
	},
}

func main() {
	ctx := EvalContext{}
	ctx.PushScope(escope)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "ERROR: no input is provided\n")
		os.Exit(1)
	}

	fpath := os.Args[1]
	content, err := os.ReadFile(fpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: could not read file %s: %s\n", fpath, err)
		os.Exit(1)
	}

	lexer := NewLexer([]rune(string(content)), fpath)

	exprs, err := ParseExprs(&lexer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, expr := range exprs {
		if _, err := ctx.EvalExpr(expr); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
