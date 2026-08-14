package main

import (
	"fmt"
	"os"

	"monkey-go/lexer"
	"monkey-go/parser"
	"monkey-go/repl"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--lex":
			runLexerDebug()
			return
		case "--parse":
			runParserDebug()
			return
		}
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
!-5;
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
	fmt.Println("=== 词法分析调试：Token 流 ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %-12s Literal: %q\n", tok.Type, tok.Literal)
		if tok.Type == "EOF" {
			break
		}
	}
}

// runParserDebug 语法分析调试，打印 AST
func runParserDebug() {
	input := `let five = 5;
let ten = 10;

let add = fn(x, y) {
	x + y;
};

let result = add(five, ten);
!-5;
!!true;
-10 * 5;
5 < 10 > 5;
a + b * c - d / e;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
`

	fmt.Println("=== 语法分析调试：Token 流 → AST ===")
	fmt.Println("\n【输入源码】")
	fmt.Println(input)

	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()

	fmt.Println("【语法错误】")
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Printf("  ❌ %s\n", err)
		}
	} else {
		fmt.Println("  ✅ 无语法错误")
	}

	fmt.Println("\n【生成的 AST】")
	fmt.Println(program.String())

	fmt.Println("【AST 节点详情】")
	for i, stmt := range program.Statements {
		fmt.Printf("  [%d] %T: %s\n", i, stmt, stmt.String())
	}
}
