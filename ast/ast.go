package ast

import (
	"monkey-go/token"
)

// Node AST节点统一接口
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement 语句接口
type Statement interface {
	Node
	statementNode()
}

// Expression 表达式接口
type Expression interface {
	Node
	expressionNode()
}

// Program 根节点，保存所有语句
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// LetStatement let 语句 let x = 10;
type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out string
	out += ls.TokenLiteral() + " "
	out += ls.Name.String()
	out += " = "
	if ls.Value != nil {
		out += ls.Value.String()
	}
	out += ";"
	return out
}

// Identifier 标识符
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// ReturnStatement return 语句
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out string
	out += rs.TokenLiteral() + " "
	if rs.ReturnValue != nil {
		out += rs.ReturnValue.String()
	}
	out += ";"
	return out
}

// ExpressionStatement 表达式语句，把表达式包装成语句
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// IntegerLiteral 整数字面量
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// PrefixExpression 前缀表达式 !true, -5
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	if pe.Right == nil {
		return "(" + pe.Operator + "<nil>)"
	}
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression 中缀表达式 a + b
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var left, right string
	if ie.Left != nil {
		left = ie.Left.String()
	} else {
		left = "<nil>"
	}
	if ie.Right != nil {
		right = ie.Right.String()
	} else {
		right = "<nil>"
	}
	return "(" + left + " " + ie.Operator + " " + right + ")"
}

// Boolean 布尔
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// IfExpression if‑else 表达式
type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out string
	out += "if"
	out += ie.Condition.String()
	out += " "
	out += ie.Consequence.String()
	if ie.Alternative != nil {
		out += "else "
		out += ie.Alternative.String()
	}
	return out
}

// BlockStatement 代码块 { ... }
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out string
	for _, s := range bs.Statements {
		out += s.String()
	}
	return out
}

// FunctionLiteral 函数字面量 fn(x,y){...}
type FunctionLiteral struct {
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out string
	params := make([]string, 0, len(fl.Parameters))
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out += fl.TokenLiteral()
	out += "("
	for i, p := range params {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	out += ") "
	out += fl.Body.String()
	return out
}

// CallExpression 函数调用 add(1,2)
type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out string
	args := make([]string, 0, len(ce.Arguments))
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out += ce.Function.String()
	out += "("
	for i, arg := range args {
		if i > 0 {
			out += ", "
		}
		out += arg
	}
	out += ")"
	return out
}
