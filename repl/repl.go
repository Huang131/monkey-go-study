package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey-go/evaluator"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"strings"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment() // REPL 全局环境，变量持久化

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			// Ctrl+D 会走到这里，正常退出
			return
		}

		line := scanner.Text()
		// 识别退出命令
		cmd := strings.TrimSpace(line)
		if cmd == "exit" || cmd == "quit" {
			fmt.Fprintln(out, "bye!")
			return
		}

		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		// 调用 Evaluator 求值，输出结果
		result := evaluator.Eval(program, env)
		if result != nil {
			io.WriteString(out, result.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
