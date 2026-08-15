package parser

import (
	"monkey-go/ast"
	"monkey-go/lexer"
	"testing"
)

// =============================================================================
// 工具函数：测试辅助
// =============================================================================

func checkParserErrors(t *testing.T, p *Parser) {
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			t.Errorf("Parser error: %s", err)
		}
		t.FailNow()
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp is not *ast.Identifier. got=%T", exp)
		return
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
	}
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		testIdentifier(t, exp, v)
		return true
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	return false
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, expected int64) bool {
	lit, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp is not *ast.IntegerLiteral. got=%T", exp)
		return false
	}
	if lit.Value != expected {
		t.Errorf("lit.Value not %d. got=%d", expected, lit.Value)
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, expected bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp is not *ast.Boolean. got=%T", exp)
		return false
	}
	if bo.Value != expected {
		t.Errorf("bo.Value not %t. got=%t", expected, bo.Value)
		return false
	}
	return true
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%+v)", exp, exp)
		return false
	}
	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}
	if opExp.Operator != operator {
		t.Errorf("opExp.Operator is not '%s'. got=%s", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}
	return true
}

// =============================================================================
// 语句解析测试
// =============================================================================

// TestLetStatement 测试 let 语句解析
func TestLetStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedName  string
		expectedValue interface{}
	}{
		{"简单整数赋值", "let x = 5;", "x", 5},
		{"布尔值赋值", "let flag = true;", "flag", true},
		{"标识符赋值", "let name = myVar;", "name", "myVar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("预期 1 条语句，实际 %d 条", len(program.Statements))
			}

			stmt := program.Statements[0]
			letStmt, ok := stmt.(*ast.LetStatement)
			if !ok {
				t.Fatalf("不是 LetStatement，实际是 %T", stmt)
			}

			// 验证变量名
			if letStmt.Name.Value != tt.expectedName {
				t.Errorf("变量名错误，预期 %s，实际 %s", tt.expectedName, letStmt.Name.Value)
			}

			// 验证值
			testLiteralExpression(t, letStmt.Value, tt.expectedValue)
		})
	}
}

// TestReturnStatement 测试 return 语句解析
func TestReturnStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue interface{}
	}{
		{"返回整数", "return 5;", 5},
		{"返回布尔", "return true;", true},
		{"返回标识符", "return result;", "result"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("预期 1 条语句，实际 %d 条", len(program.Statements))
			}

			stmt := program.Statements[0]
			returnStmt, ok := stmt.(*ast.ReturnStatement)
			if !ok {
				t.Fatalf("不是 ReturnStatement，实际是 %T", stmt)
			}

			testLiteralExpression(t, returnStmt.ReturnValue, tt.expectedValue)
		})
	}
}

// TestExpressionStatement 测试表达式语句解析
func TestExpressionStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"标识符", "foobar;", "foobar"},
		{"整数", "42;", 42},
		{"布尔 true", "true;", true},
		{"布尔 false", "false;", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("预期 1 条语句，实际 %d 条", len(program.Statements))
			}

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			testLiteralExpression(t, stmt.Expression, tt.expected)
		})
	}
}

// =============================================================================
// 表达式解析测试
// =============================================================================

// TestPrefixExpression 测试前缀表达式（!x, -x）
func TestPrefixExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
		value    interface{}
	}{
		{"非布尔", "!true", "!", true},
		{"非整数", "!5", "!", 5},
		{"负整数", "-10", "-", 10},
		{"负整数2", "-5", "-", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			exp := stmt.Expression.(*ast.PrefixExpression)

			if exp.Operator != tt.operator {
				t.Errorf("运算符错误，预期 %s，实际 %s", tt.operator, exp.Operator)
			}
			testLiteralExpression(t, exp.Right, tt.value)
		})
	}
}

// TestNestedPrefixExpression 测试嵌套前缀表达式（!!x, ---x）
func TestNestedPrefixExpression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedStr string // AST.String() 的预期输出
	}{
		{"双重非", "!!false", "(!(!false))"},
		{"三重非", "!!!true", "(!(!(!true)))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			expected := tt.expectedStr
			actual := stmt.Expression.String()
			if actual != expected {
				t.Errorf("嵌套前缀表达式错误\n输入: %s\n预期: %s\n实际: %s", tt.input, expected, actual)
			}
		})
	}
}

// TestInfixExpression 测试中缀表达式（x + y, x == y）
func TestInfixExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		// 算术运算
		{"加法", "5 + 5", 5, "+", 5},
		{"减法", "5 - 3", 5, "-", 3},
		{"乘法", "3 * 4", 3, "*", 4},
		{"除法", "10 / 2", 10, "/", 2},
		// 比较运算
		{"小于", "3 < 5", 3, "<", 5},
		{"大于", "5 > 3", 5, ">", 3},
		// 布尔运算
		{"相等", "true == true", true, "==", true},
		{"不等", "true != false", true, "!=", false},
		// 布尔值参与
		{"布尔比较", "false == false", false, "==", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			testInfixExpression(t, stmt.Expression, tt.left, tt.operator, tt.right)
		})
	}
}

// TestIfExpression 测试 if 表达式
func TestIfExpression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		conditionL  interface{}
		conditionOp string
		conditionR  interface{}
		hasElse     bool
	}{
		{
			name:       "无 else 分支",
			input:      "if (x < y) { x }",
			conditionL: "x", conditionOp: "<", conditionR: "y",
			hasElse: false,
		},
		{
			name:       "有 else 分支",
			input:      "if (x < y) { x } else { y }",
			conditionL: "x", conditionOp: "<", conditionR: "y",
			hasElse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			exp := stmt.Expression.(*ast.IfExpression)

			// 验证条件
			testInfixExpression(t, exp.Condition, tt.conditionL, tt.conditionOp, tt.conditionR)

			// 验证 if 分支
			if len(exp.Consequence.Statements) != 1 {
				t.Errorf("if 分支语句数错误，预期 1，实际 %d", len(exp.Consequence.Statements))
			}

			// 验证 else 分支
			if tt.hasElse && exp.Alternative == nil {
				t.Error("应该有 else 分支，但实际为 nil")
			}
			if !tt.hasElse && exp.Alternative != nil {
				t.Errorf("不应有 else 分支，但实际有 %d 条语句", len(exp.Alternative.Statements))
			}
		})
	}
}

// TestFunctionLiteral 测试函数字面量
func TestFunctionLiteral(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedParams  []string
		expectedBodyOps []string // body 语句的操作符列表
	}{
		{
			name:            "无参函数",
			input:           "fn() { 0 }",
			expectedParams:  []string{},
			expectedBodyOps: []string{},
		},
		{
			name:            "单参函数",
			input:           "fn(x) { x }",
			expectedParams:  []string{"x"},
			expectedBodyOps: []string{},
		},
		{
			name:            "多参函数",
			input:           "fn(x, y, z) { x + y + z }",
			expectedParams:  []string{"x", "y", "z"},
			expectedBodyOps: []string{"+", "+"},
		},
		{
			name:            "函数体有多条语句",
			input:           "fn(a) { let b = a; return b; }",
			expectedParams:  []string{"a"},
			expectedBodyOps: []string{}, // 有 let 和 return
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			fn := stmt.Expression.(*ast.FunctionLiteral)

			// 验证参数数量
			if len(fn.Parameters) != len(tt.expectedParams) {
				t.Errorf("参数数量错误，预期 %d，实际 %d", len(tt.expectedParams), len(fn.Parameters))
			}

			// 验证参数名
			for i, param := range tt.expectedParams {
				if fn.Parameters[i].Value != param {
					t.Errorf("参数 %d 错误，预期 %s，实际 %s", i, param, fn.Parameters[i].Value)
				}
			}
		})
	}
}

// TestCallExpression 测试函数调用
func TestCallExpression(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedFuncName string
		expectedArgCount int
	}{
		{
			name:             "空参数调用",
			input:            "add()",
			expectedFuncName: "add",
			expectedArgCount: 0,
		},
		{
			name:             "单参数调用",
			input:            "double(5)",
			expectedFuncName: "double",
			expectedArgCount: 1,
		},
		{
			name:             "多参数调用",
			input:            "add(1, 2, 3)",
			expectedFuncName: "add",
			expectedArgCount: 3,
		},
		{
			name:             "表达式参数",
			input:            "add(1 + 2, 3 * 4)",
			expectedFuncName: "add",
			expectedArgCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			call := stmt.Expression.(*ast.CallExpression)

			// 验证函数名
			testIdentifier(t, call.Function, tt.expectedFuncName)

			// 验证参数数量
			if len(call.Arguments) != tt.expectedArgCount {
				t.Errorf("参数数量错误，预期 %d，实际 %d", tt.expectedArgCount, len(call.Arguments))
			}
		})
	}
}

// =============================================================================
// 整体集成测试
// =============================================================================

// TestOperatorPrecedence 用表格驱动测试所有优先级场景
func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // AST.String() 的预期输出
	}{
		// 优先级：prefix (!, -) > product (*) > sum (+) > comparison (<, >) > equality (==, !=)
		{"前缀运算", "-a * b", "((-a) * b)"},
		{"双重前缀", "!-a", "(!(-a))"},
		{"左结合加法", "a + b + c", "((a + b) + c)"},
		{"左结合乘法", "a * b * c", "((a * b) * c)"},
		{"乘除优先于加减", "a + b * c", "(a + (b * c))"},
		{"加减乘除混合", "a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"括号改变优先级", "a + (b + c)", "(a + (b + c))"},
		{"括号包裹表达式", "(a + b) * c", "((a + b) * c)"},
		{"复杂表达式", "3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"逻辑非", "!(true == true)", "(!(true == true))"},
		{"比较运算", "5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		// 函数调用
		{"函数调用", "add(a, b)", "add(a, b)"},
		{"函数调用含表达式参数", "add(a + b, c * d)", "add((a + b), (c * d))"},
		{"嵌套函数调用", "add(add(1, 2), 3)", "add(add(1, 2), 3)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			actual := program.String()
			if actual != tt.expected {
				t.Errorf("\n输入: %s\n预期: %s\n实际: %s", tt.input, tt.expected, actual)
			}
		})
	}
}

// TestComplexProgram 测试复杂程序
func TestComplexProgram(t *testing.T) {
	input := `
	let add = fn(x, y) { x + y };
	let result = add(5, 10);
	if (result > 10) {
		return true;
	} else {
		return false;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// 验证程序有 3 条语句：let add, let result, if...else
	if len(program.Statements) != 3 {
		t.Fatalf("预期 3 条语句，实际 %d 条", len(program.Statements))
	}

	// 验证第一句：let add = fn(x, y) { x + y };
	letAdd := program.Statements[0].(*ast.LetStatement)
	if letAdd.Name.Value != "add" {
		t.Errorf("第一句变量名错误，预期 add，实际 %s", letAdd.Name.Value)
	}
	if _, ok := letAdd.Value.(*ast.FunctionLiteral); !ok {
		t.Error("第一句赋值不是函数字面量")
	}

	// 验证第二句：let result = add(5, 10);
	letResult := program.Statements[1].(*ast.LetStatement)
	if letResult.Name.Value != "result" {
		t.Errorf("第二句变量名错误，预期 result，实际 %s", letResult.Name.Value)
	}
	if _, ok := letResult.Value.(*ast.CallExpression); !ok {
		t.Error("第二句赋值不是函数调用")
	}

	// 验证第三句：if...else 语句
	ifStmt, ok := program.Statements[2].(*ast.ExpressionStatement).Expression.(*ast.IfExpression)
	if !ok {
		t.Error("第三句不是 if 表达式")
	}
	if ifStmt.Alternative == nil {
		t.Error("if 表达式没有 else 分支")
	}

	// 验证 if 条件：result > 10
	condition, ok := ifStmt.Condition.(*ast.InfixExpression)
	if !ok {
		t.Error("if 条件不是中缀表达式")
	}
	if condition.Operator != ">" {
		t.Errorf("if 条件运算符错误，预期 >，实际 %s", condition.Operator)
	}
}
