package main

import (
	"fmt"
	"os"

	"monkey-go/evaluator"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"monkey-go/repl"
)

// debugInput 三个调试函数共用的测试代码
// 涵盖：基础类型、函数、控制流、字符串、数组、哈希、内置函数
// 注意：暂不支持行注释 (//)，使用空行分隔不同类型的代码
const debugInput = `let five = 5;
let ten = 10;

let greeting = "Hello, " + "World!";
let name = "Monkey";

let numbers = [1, 2, 3, 4, 5];
let mixed = ["hello", 42, true];

let person = {"name": "Tom", "age": 25};

let add = fn(x, y) {
	x + y;
};

let result = add(five, ten);
!-5;
!!true;
-10 * 5;

let first = numbers[0];
let last = numbers[len(numbers) - 1];
let personName = person["name"];

let arr = [1, 2, 3];
let arrLen = len(arr);
let arrFirst = first(arr);
let arrLast = last(arr);
let arrRest = rest(arr);
let arrPushed = push(arr, 4);

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;

let fibonacci = fn(x) {
	if (x < 2) {
		return x;
	}
	return fibonacci(x - 1) + fibonacci(x - 2);
};

fibonacci(10);`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--lex":
			runLexerDebug()
			return
		case "--parse":
			runParserDebug()
			return
		case "--eval":
			runEvalDebug()
			return
		}
	}

	// 默认：启动REPL交互式终端（lexer + parser + evaluator）
	repl.Start(os.Stdin, os.Stdout)
}

// runLexerDebug 词法分析调试，打印所有token
func runLexerDebug() {
	l := lexer.New(debugInput)
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
	fmt.Println("=== 语法分析调试：Token 流 → AST ===")
	fmt.Println("\n【输入源码】")
	fmt.Println(debugInput)

	l := lexer.New(debugInput)
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

// runEvalDebug 求值调试：完整流程 Lexer → Parser → Evaluator
func runEvalDebug() {
	fmt.Println("=== 求值调试：完整流程 Lexer → Parser → Evaluator ===")
	fmt.Println("\n【输入源码】")
	fmt.Println(debugInput)

	// Step 1: Lexer（独立实例）
	fmt.Println("\n【Step 1: Lexer Token 流】(前20个)")
	l1 := lexer.New(debugInput)
	tokens := []string{}
	for {
		tok := l1.NextToken()
		tokens = append(tokens, fmt.Sprintf("%-12s %q", tok.Type, tok.Literal))
		if tok.Type == "EOF" {
			break
		}
	}
	for i := 0; i < 20 && i < len(tokens); i++ {
		fmt.Printf("  %s\n", tokens[i])
	}
	fmt.Printf("  ... (%d tokens total)\n", len(tokens))

	// Step 2: Parser（独立实例）
	fmt.Println("\n【Step 2: Parser AST】")
	l2 := lexer.New(debugInput)
	p := parser.New(l2)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("  ❌ 语法错误:")
		for _, err := range p.Errors() {
			fmt.Printf("     %s\n", err)
		}
		return
	}
	fmt.Println("  ✅ 无语法错误")
	fmt.Printf("  AST 字符串: %s\n", program.String())
	fmt.Printf("  语句数量: %d\n", len(program.Statements))

	// Step 3: Evaluator（独立环境）
	fmt.Println("\n【Step 3: Evaluator 求值】")
	env := object.NewEnvironment()

	for i, stmt := range program.Statements {
		fmt.Printf("\n  语句 [%d]: %T\n", i, stmt)
		fmt.Printf("    AST: %s\n", stmt.String())

		result := evaluator.Eval(stmt, env)

		if result != nil {
			switch result.Type() {
			case object.ERROR_OBJ:
				fmt.Printf("    ❌ 错误: %s\n", result.Inspect())
			case object.FUNCTION_OBJ:
				fmt.Printf("    ✅ 函数对象: %s\n", truncate(result.Inspect(), 50))
			case object.STRING_OBJ:
				fmt.Printf("    ✅ 字符串: %q\n", result.Inspect())
			case object.ARRAY_OBJ:
				fmt.Printf("    ✅ 数组: %s\n", result.Inspect())
			case object.HASH_OBJ:
				fmt.Printf("    ✅ 哈希: %s\n", result.Inspect())
			case object.BUILTIN_OBJ:
				fmt.Printf("    ✅ 内置函数: %s\n", result.Inspect())
			default:
				fmt.Printf("    ✅ 结果: %s\n", result.Inspect())
			}
		} else {
			fmt.Printf("    ℹ️  (let语句，无返回值)\n")
		}
	}

	// 环境变量
	fmt.Println("\n【环境变量】")
	vars := []string{
		"five", "ten", "greeting", "name",
		"numbers", "mixed", "person",
		"add", "result",
		"arr", "arrFirst", "arrLast", "arrRest", "arrPushed", "arrLen",
		"fibonacci",
	}
	for _, v := range vars {
		if val, ok := env.Get(v); ok {
			switch val.Type() {
			case object.FUNCTION_OBJ:
				fmt.Printf("  %s: fn(...)\n", v)
			case object.STRING_OBJ:
				fmt.Printf("  %s: %q\n", v, val.Inspect())
			case object.ARRAY_OBJ:
				fmt.Printf("  %s: %s\n", v, val.Inspect())
			case object.HASH_OBJ:
				fmt.Printf("  %s: %s\n", v, val.Inspect())
			default:
				fmt.Printf("  %s: %s\n", v, val.Inspect())
			}
		}
	}
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
