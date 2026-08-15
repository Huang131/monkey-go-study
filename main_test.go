package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"monkey-go/evaluator"
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
)

// TestMain 提供测试setup/teardown
func TestMain(m *testing.M) {
	// 运行所有测试
	code := m.Run()
	os.Exit(code)
}

// testEval 辅助函数：解析并求值一段代码
func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return evaluator.Eval(program, env)
}

// TestRunLexerDebug 测试词法分析调试输出
func TestRunLexerDebug(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runLexerDebug()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := buf.String()

	// 验证输出包含预期的Token类型
	expectedTokens := []string{"LET", "IDENT", "INT", "FUNCTION", "("}
	for _, token := range expectedTokens {
		if !strings.Contains(output, token) {
			t.Errorf("expected output to contain token %q", token)
		}
	}
}

// TestRunParserDebug 测试语法分析调试输出
func TestRunParserDebug(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runParserDebug()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := buf.String()

	// 验证输出包含关键信息
	if !strings.Contains(output, "=== 语法分析调试") {
		t.Error("expected output to contain debug header")
	}
	if !strings.Contains(output, "无语法错误") {
		t.Error("expected output to contain '无语法错误'")
	}
	if !strings.Contains(output, "AST 节点详情") {
		t.Error("expected output to contain 'AST 节点详情'")
	}
}

// TestRunEvalDebug 测试求值调试输出
func TestRunEvalDebug(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runEvalDebug()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := buf.String()

	// 验证输出包含关键信息
	if !strings.Contains(output, "=== 求值调试") {
		t.Error("expected output to contain debug header")
	}
	if !strings.Contains(output, "Step 1: Lexer") {
		t.Error("expected output to contain 'Step 1: Lexer'")
	}
	if !strings.Contains(output, "Step 2: Parser") {
		t.Error("expected output to contain 'Step 2: Parser'")
	}
	if !strings.Contains(output, "Step 3: Evaluator") {
		t.Error("expected output to contain 'Step 3: Evaluator'")
	}
	// 验证环境变量输出
	if !strings.Contains(output, "five: 5") {
		t.Error("expected output to contain 'five: 5'")
	}
	if !strings.Contains(output, "ten: 10") {
		t.Error("expected output to contain 'ten: 10'")
	}
}

// TestEvalBasicExpressions 测试基础表达式求值
func TestEvalBasicExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5", "5"},
		{"10", "10"},
		{"-5", "-5"},
		{"5 + 5", "10"},
		{"10 - 5", "5"},
		{"5 * 5", "25"},
		{"20 / 4", "5"},
		{"50 / 10 + 5", "10"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestEvalBooleanExpressions 测试布尔表达式求值
func TestEvalBooleanExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"true", "true"},
		{"false", "false"},
		{"!true", "false"},
		{"!false", "true"},
		{"!!true", "true"},
		{"5 == 5", "true"},
		{"5 != 5", "false"},
		{"5 < 10", "true"},
		{"5 > 10", "false"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestEvalLetStatements 测试变量绑定
func TestEvalLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let a = 5; a", "5"},
		{"let a = 5 * 5; a", "25"},
		{"let a = 5; let b = a; b", "5"},
		{"let a = 5; let b = a + 1; b", "6"},
		{"let a = 5; let b = 6; a + b", "11"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestEvalFunctionLiterals 测试函数字面量
func TestEvalFunctionLiterals(t *testing.T) {
	input := "fn(x) { x + 1 }"
	obj := testEval(input)

	if obj.Type() != object.FUNCTION_OBJ {
		t.Errorf("expected FUNCTION, got %s", obj.Type())
	}

	fn := obj.(*object.Function)
	if len(fn.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(fn.Parameters))
	}
	if fn.Parameters[0].Value != "x" {
		t.Errorf("expected parameter 'x', got %q", fn.Parameters[0].Value)
	}
}

// TestEvalFunctionCalls 测试函数调用
func TestEvalFunctionCalls(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let identity = fn(x) { x }; identity(5)", "5"},
		{"let double = fn(x) { x * 2 }; double(5)", "10"},
		{"let add = fn(x, y) { x + y }; add(5, 5)", "10"},
		{"let add = fn(x, y) { x + y }; add(5 + 5, add(5, 5))", "20"},
		{"(fn(x) { x })(5)", "5"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestEvalReturnStatements 测试return语句
func TestEvalReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"return 5;", "5"},
		{"return 10; 9;", "10"}, // 9不会被执行
		{"let add = fn(x) { return x; x + 1 }; add(5)", "5"},
		{"if (true) { return 5; }", "5"},
		{"if (false) { return 5; } else { return 10; }", "10"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestEvalIfExpressions 测试if表达式
func TestEvalIfExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"if (true) { 10 }", "10"},
		{"if (false) { 10 }", "null"},
		{"if (1) { 10 }", "10"}, // 数字都为真
		{"if (0) { 10 }", "10"}, // 0也是真（只有false为假）
		{"if (true) { 10 } else { 20 }", "10"},
		{"if (false) { 10 } else { 20 }", "20"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestEvalErrorHandling 测试错误处理
func TestEvalErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"5 + true", true},
		{"5 - true", true},
		{"5 * true", true},
		{"true + false", true},
		{"-true", true},
		{"foobar", true}, // undefined identifier
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Type() != object.ERROR_OBJ && tt.expectError {
			t.Errorf("eval(%q) = %q, expected error", tt.input, obj.Inspect())
		}
		if obj.Type() == object.ERROR_OBJ && !tt.expectError {
			t.Errorf("eval(%q) unexpected error: %s", tt.input, obj.Inspect())
		}
	}
}

// TestEvalClosure 测试闭包
func TestEvalClosure(t *testing.T) {
	// 测试闭包捕获定义时的环境变量
	input := `
let a = 10;
let f = fn() { return a; };
f();
`
	obj := testEval(input)
	if obj.Inspect() != "10" {
		t.Errorf("expected 10, got %s", obj.Inspect())
	}

	// 测试闭包中修改外部变量不会影响已捕获的值
	input2 := `
let a = 5;
let f = fn() { a };
let g = fn() {
	let a = 100;
	f();
};
g();
`
	obj2 := testEval(input2)
	if obj2.Inspect() != "5" {
		t.Errorf("expected 5 (closure captured), got %s", obj2.Inspect())
	}
}

// TestEvalFibonacci 测试递归（斐波那契）
func TestEvalFibonacci(t *testing.T) {
	// 简单的斐波那契实现
	input := `
let fibonacci = fn(x) {
	if (x < 2) {
		return x;
	}
	return fibonacci(x - 1) + fibonacci(x - 2);
};
fibonacci(10);
`
	obj := testEval(input)
	// fibonacci(10) = 55
	if obj.Inspect() != "55" {
		t.Errorf("expected 55, got %s", obj.Inspect())
	}
}

// TestRunEvalDebugOutput 检查调试输出包含关键信息
func TestRunEvalDebugOutput(t *testing.T) {
	// 由于 runEvalDebug 直接打印到 stdout，我们通过捕获输出来验证
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runEvalDebug()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)
	output := buf.String()

	// 验证 fibonacci(10) 的结果被正确输出
	if !strings.Contains(output, "55") {
		t.Error("expected output to contain fibonacci result: 55")
	}

	// 验证 add 函数对象输出
	if !strings.Contains(output, "add: fn(...)") {
		t.Error("expected output to contain 'add: fn(...)'")
	}

	// 验证 result 变量输出
	if !strings.Contains(output, "result: 15") {
		t.Error("expected output to contain 'result: 15'")
	}
}

// TestTruncate 测试 truncate 辅助函数
func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is a ..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

// =============================================================================
// 新功能测试：字符串、数组、哈希、内置函数
// =============================================================================

// TestEvalString 测试字符串字面量
func TestEvalString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"Hello" + " " + "World!"`, "Hello World!"},
		{`"foo" + "bar"`, "foobar"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestEvalArray 测试数组字面量和索引
func TestEvalArray(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[1, 2, 3]", "[1, 2, 3]"},
		{"[1, 2, 3][0]", "1"},
		{"[1, 2, 3][1]", "2"},
		{"[1, 2, 3][2]", "3"},
		{"[1, 2, 3][-1]", "null"}, // 越界返回 null
		{"[1, 2, 3][5]", "null"},  // 越界返回 null
		{"let arr = [1, 2, 3]; arr[0]", "1"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestEvalHash 测试哈希字面量和索引
func TestEvalHash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"name": "Tom"}`, "{name: Tom}"},
		{`{"foo": 5}["foo"]`, "5"},
		{`{"foo": 5}["bar"]`, "null"},             // key 不存在
		{`let key = "foo"; {"foo": 5}[key]`, "5"}, // 变量作为 key
		{`let h = {"a": 1, "b": 2}; h["a"]`, "1"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestBuiltinLen 测试 len 内置函数
func TestBuiltinLen(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`len("hello")`, "5"},
		{`len("")`, "0"},
		{`len([1, 2, 3])`, "3"},
		{`len([])`, "0"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestBuiltinArrayFunctions 测试数组内置函数
func TestBuiltinArrayFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"first([1, 2, 3])", "1"},
		{"first([5])", "5"},
		{"first([])", "null"}, // 空数组返回 null
		{"last([1, 2, 3])", "3"},
		{"last([5])", "5"},
		{"last([])", "null"}, // 空数组返回 null
		{"rest([1, 2, 3])", "[2, 3]"},
		{"rest([5])", "null"}, // 单元素返回 null
		{"rest([])", "null"},  // 空数组返回 null
		{"push([1, 2], 3)", "[1, 2, 3]"},
		{"push([], 1)", "[1]"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestBuiltinErrorHandling 测试内置函数错误处理
func TestBuiltinErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"len(1)", true},          // 不支持的参数类型
		{"first(1)", true},        // 参数不是数组
		{"last(1)", true},         // 参数不是数组
		{"rest(1)", true},         // 参数不是数组
		{"push(1, 2)", true},      // 参数不是数组
		{"len(1, 2)", true},       // 参数个数错误
		{"first([])", false},      // 空数组合法，返回 null
		{"len([1,2,3])", false},   // 正确用法
		{"len(\"hello\")", false}, // 正确用法
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		isError := obj.Type() == object.ERROR_OBJ
		if isError != tt.expectError {
			if tt.expectError {
				t.Errorf("eval(%q) expected error, got %q", tt.input, obj.Inspect())
			} else {
				t.Errorf("eval(%q) unexpected error: %s", tt.input, obj.Inspect())
			}
		}
	}
}

// TestFullPipelineIntegration 测试完整流水线：字符串 → 数组 → 哈希 → 内置函数
func TestFullPipelineIntegration(t *testing.T) {
	// 测试复杂场景：各种新功能组合使用
	input := `
let arr = ["a", "b", "c"];
let str = "hello";
let combinedLen = len(arr) + len(str);
combinedLen;
`
	obj := testEval(input)
	// len(arr)=3 + len(str)=5 = 8
	if obj.Inspect() != "8" {
		t.Errorf("expected 8, got %s", obj.Inspect())
	}
}

// TestBuiltinOverride 测试内置函数覆盖
func TestBuiltinOverride(t *testing.T) {
	// 用户可以覆盖内置函数
	input := `
let len = fn(x) { "自定义len" };
len(5);
`
	obj := testEval(input)
	if obj.Inspect() != "自定义len" {
		t.Errorf("expected '自定义len', got %s", obj.Inspect())
	}
}

// TestArrayWithExpressions 测试数组中的表达式
func TestArrayWithExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[1 + 1, 2 * 2, 3]", "[2, 4, 3]"},
		{"[len(\"hi\"), first([1,2]), rest([3,4])[0]]", "[2, 1, 4]"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}

// TestHashWithComputedKeys 测试哈希计算键
func TestHashWithComputedKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"one": 1 + 0}`, "{one: 1}"},
		{`let k = "key"; {k: 100}["key"]`, "100"},
		{`{"a": 1, "b": 2, "c": 3}["b"]`, "2"},
	}

	for _, tt := range tests {
		obj := testEval(tt.input)
		if obj == nil {
			t.Errorf("eval(%q) returned nil", tt.input)
			continue
		}
		if obj.Inspect() != tt.expected {
			t.Errorf("eval(%q) = %q, want %q", tt.input, obj.Inspect(), tt.expected)
		}
	}
}
