package main

import (
	"bytes"
	"fmt"
	"syscall/js"

	"github.com/ankalang/anka/evaluator"
	"github.com/ankalang/anka/lexer"
	"github.com/ankalang/anka/object"
	"github.com/ankalang/anka/parser"
)


var Version = "dev"


func runCode(this js.Value, i []js.Value) interface{} {
	m := make(map[string]interface{})
	var buf bytes.Buffer
	
	code := i[0].String()
	env := object.NewEnvironment(&buf, "", Version)
	lex := lexer.New(code)
	p := parser.New(lex)

	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors(), buf)
		m["out"] = buf.String()
		return js.ValueOf(m)
	}

	result := evaluator.BeginEval(program, env, lex)
	m["out"] = buf.String()
	m["result"] = result.Inspect()

	return js.ValueOf(m)
}

func printParserErrors(errors []string, buf bytes.Buffer) {
	fmt.Fprintf(&buf, "%s", " Ayrıştırıcı hatası:\n")
	for _, msg := range errors {
		fmt.Fprintf(&buf, "%s", "\t"+msg+"\n")
	}
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("ank_run_code", js.FuncOf(runCode))
	<-c
}
