package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type LintError struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Length   int    `json:"length"`
}

type LambdaResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func main() {
	lambda.Start(handler)
}

func handler(event Event) (LambdaResponse, error) {
	if strings.TrimSpace(event.Code) == "" {
		return jsonResponse(400, []LintError{{
			Line:     1,
			Column:   1,
			Message:  "No code provided",
			Severity: "error",
			Length:   1,
		}})
	}

	errors := lintGo(event.Code)
	return jsonResponse(200, errors)
}

func lintGo(code string) []LintError {
	fset := token.NewFileSet()
	file, parseErr := parser.ParseFile(fset, "main.go", code, parser.AllErrors)
	if parseErr != nil {
		return parseErrors(fset, parseErr)
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	var errors []LintError
	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			if typeErr, ok := err.(types.Error); ok {
				pos := fset.Position(typeErr.Pos)
				errors = append(errors, LintError{
					Line:     pos.Line,
					Column:   pos.Column,
					Message:  typeErr.Msg,
					Severity: "error",
					Length:   1,
				})
			}
		},
	}

	_, _ = conf.Check("lintpkg", fset, []*ast.File{file}, info)
	return errors
}

func parseErrors(fset *token.FileSet, err error) []LintError {
	var errors []LintError
	for _, line := range strings.Split(err.Error(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 4)
		if len(parts) == 4 {
			lineNo := 1
			column := 1
			if _, convErr := fmt.Sscanf(parts[1], "%d", &lineNo); convErr == nil {
				_, _ = fmt.Sscanf(parts[2], "%d", &column)
			}

			errors = append(errors, LintError{
				Line:     lineNo,
				Column:   column,
				Message:  strings.TrimSpace(parts[3]),
				Severity: "error",
				Length:   1,
			})
			continue
		}

		errors = append(errors, LintError{
			Line:     1,
			Column:   1,
			Message:  line,
			Severity: "error",
			Length:   1,
		})
	}

	if len(errors) == 0 {
		errors = append(errors, LintError{
			Line:     1,
			Column:   1,
			Message:  err.Error(),
			Severity: "error",
			Length:   1,
		})
	}

	return errors
}

func jsonResponse(statusCode int, errors []LintError) (LambdaResponse, error) {
	body, err := json.Marshal(map[string]any{
		"errors": errors,
	})
	if err != nil {
		return LambdaResponse{}, err
	}

	return LambdaResponse{
		StatusCode: statusCode,
		Body:       string(body),
	}, nil
}
