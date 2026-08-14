package main

import (
	"fmt"
	"os"

	"monkey-go/lexer"
	"monkey-go/repl"
)

func main() {
	// 如果命令行传 --lex，执行词法分析调试
	if len(os.Args) > 1 && os.Args[1] == "--lex" {
		runLexerDebug()
		return
	}

	// 默认：启动REPL交互式终端（lexer + parser）
	repl.Start(os.Stdin, os.Stdout)
}

// runLexerDebug 词法分析调试，打印所有token
func runLexerDebug() {
	input := `let five = 5;
let ten = 10;

let add = fn(x, y) {
	x + y;
};

let result = add(five, ten);
!-/*5;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
`

	l := lexer.New(input)
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %-10s Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == "EOF" {
			break
		}
	}
}
