package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey-go/lexer"
	"monkey-go/parser"
	"strings"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

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

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
