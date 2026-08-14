package ast

import (
	"bytes"
	"monkey-go/token"
)

// =============================================================================
// 一、三个核心接口：定义 AST 节点的规范
// =============================================================================

// Node 所有 AST 节点的公共接口
// 每个节点必须能返回：1) TokenLiteral() 自身关键字 2) String() 打印自身
type Node interface {
	TokenLiteral() string // 返回节点对应的 Token 原始字面量（如 "let", "if", "5"）
	String() string       // 返回该节点的字符串表示，用于调试输出
}

// Statement 语句接口
// 语句：不返回值（或者说返回值是"副作用"），如 let、return、if 等
type Statement interface {
	Node
	statementNode() // 空方法，仅用于类型区分，无实际意义
}

// Expression 表达式接口
// 表达式：有值，可以参与运算、赋值、作为函数参数等
type Expression interface {
	Node
	expressionNode() // 空方法，仅用于类型区分，无实际意义
}

// =============================================================================
// 二、顶层结构：Program
// =============================================================================

// Program 程序的根节点
// 整个源码文件解析后，会生成一个 Program，包含所有顶层语句
type Program struct {
	Statements []Statement // 顶层语句列表（如 let x = 5; return 10;）
}

// TokenLiteral 返回第一个语句的关键字，用于调试
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String 将整个程序转成字符串，用于调试输出 AST
// 内部使用 bytes.Buffer 避免字符串拼接性能问题
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// =============================================================================
// 三、语句节点
// =============================================================================

// LetStatement 变量绑定语句
// 语法: let <identifier> = <expression>;
// 示例: let x = 5; let add = fn(a, b) { a + b };
type LetStatement struct {
	Token token.Token // "let" Token
	Name  *Identifier // 变量名（只能是 Identifier）
	Value Expression  // 变量值表达式（可以是任意表达式）
}

func (ls *LetStatement) statementNode()       {}                          // 语句标记
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal } // 返回 "let"
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ") // "let "
	out.WriteString(ls.Name.String())        // "x"
	out.WriteString(" = ")                   // " = "
	if ls.Value != nil {
		out.WriteString(ls.Value.String()) // "5"
	}
	out.WriteString(";") // ";"
	return out.String()  // "let x = 5;"
}

// ReturnStatement return 语句
// 语法: return <expression>;
// 示例: return 10; return x + y;
type ReturnStatement struct {
	Token       token.Token // "return" Token
	ReturnValue Expression  // 返回值表达式（可为空，如单独的 "return;"）
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal } // 返回 "return"
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ") // "return "
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String()) // "10"
	}
	out.WriteString(";")
	return out.String() // "return 10;"
}

// ExpressionStatement 表达式包装成语句
// 作用：将表达式作为独立语句执行（如 "5 + 3;" "add(1, 2);")
// 注意：单独的 "5;" 是表达式语句，不是 let 语句
type ExpressionStatement struct {
	Token      token.Token // 表达式开头的 Token
	Expression Expression  // 内部封装的表达式
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// =============================================================================
// 四、表达式节点
// =============================================================================

// Identifier 标识符（变量名）
// 语法: <identifier>
// 示例: x, y, add, result
type Identifier struct {
	Token token.Token // IDENT Token
	Value string      // 标识符名字（"x", "add" 等）
}

func (i *Identifier) expressionNode()      {}                         // 表达式标记
func (i *Identifier) TokenLiteral() string { return i.Token.Literal } // 返回变量名
func (i *Identifier) String() string       { return i.Value }         // 同上，简化输出

// IntegerLiteral 整数字面量
// 语法: <integer>
// 示例: 5, 100, -42
type IntegerLiteral struct {
	Token token.Token // INT Token
	Value int64       // 整数值
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal } // 返回 "5"
func (il *IntegerLiteral) String() string       { return il.Token.Literal } // 同上

// BooleanLiteral 布尔字面量
// 语法: true / false
type BooleanLiteral struct {
	Token token.Token // TRUE 或 FALSE Token
	Value bool        // true 或 false
}

// Boolean 是 BooleanLiteral 的别名，用于简化代码
type Boolean = BooleanLiteral

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// PrefixExpression 前缀表达式（一元运算符）
// 语法: <operator><expression>
// 示例: !true, -5, !!false
type PrefixExpression struct {
	Token    token.Token // "!" 或 "-" Token
	Operator string      // 运算符（"!" 或 "-"）
	Right    Expression  // 右操作数表达式
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)       // "!" 或 "-"
	out.WriteString(pe.Right.String()) // 操作数
	out.WriteString(")")
	return out.String() // "(!true)" 或 "(-5)"
}

// InfixExpression 中缀表达式（二元运算符）
// 语法: <left> <operator> <right>
// 示例: 5 + 3, x == y, a + b * c
type InfixExpression struct {
	Token    token.Token // "+", "-", "*", "/", "==", "!=", "<", ">" 等 Token
	Left     Expression  // 左操作数
	Operator string      // 运算符
	Right    Expression  // 右操作数
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())        // "5"
	out.WriteString(" " + ie.Operator + " ") // " + "
	out.WriteString(ie.Right.String())       // "3"
	out.WriteString(")")
	return out.String() // "(5 + 3)"
}

// IfExpression if-else 表达式（注意：Monkey 的 if 是表达式不是语句）
// 语法: if (<condition>) { <consequence> } else { <alternative> }
// 示例: if (x > 0) { return x } else { return -x }
type IfExpression struct {
	Token       token.Token     // "if" Token
	Condition   Expression      // 条件表达式
	Consequence *BlockStatement // if 成立时执行的代码块
	Alternative *BlockStatement // else 代码块（可为空，即无 else 分支）
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String()) // 条件
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String()) // if 块
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String()) // else 块
	}
	return out.String() // "if(x > 0) {...} else {...}"
}

// BlockStatement 代码块（语句的集合，用 {} 包裹）
// 语法: { <statements> }
// 示例: { let x = 5; return x; }
type BlockStatement struct {
	Token      token.Token // "{" Token
	Statements []Statement // 块内的语句列表
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// FunctionLiteral 函数字面量（函数定义）
// 语法: fn (<parameters>) { <body> }
// 示例: fn(x, y) { x + y }
type FunctionLiteral struct {
	Token      token.Token     // "fn" Token
	Parameters []*Identifier   // 形参列表 ["x", "y"]
	Body       *BlockStatement // 函数体代码块
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	// 收集参数名
	params := make([]string, 0, len(fl.Parameters))
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	// 输出: fn(参数) { 函数体 }
	out.WriteString(fl.TokenLiteral()) // "fn"
	out.WriteString("(")
	for i, p := range params {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p)
	}
	out.WriteString(")")
	out.WriteString(fl.Body.String())
	return out.String() // "fn(x, y) { x + y; }"
}

// CallExpression 函数调用表达式
// 语法: <function>(<arguments>)
// 示例: add(1, 2), fn(x){x}(5)
type CallExpression struct {
	Token     token.Token  // "(" Token
	Function  Expression   // 被调用的函数（可以是标识符或直接是函数字面量）
	Arguments []Expression // 实参列表
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	// 收集实参字符串
	args := make([]string, 0, len(ce.Arguments))
	for _, arg := range ce.Arguments {
		args = append(args, arg.String())
	}

	// 输出: 函数(实参1, 实参2, ...)
	out.WriteString(ce.Function.String()) // "add"
	out.WriteString("(")
	for i, a := range args {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(a)
	}
	out.WriteString(")")
	return out.String() // "add(1, 2)"
}
