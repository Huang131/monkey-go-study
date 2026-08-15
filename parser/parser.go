package parser

import (
	"fmt"
	"monkey-go/ast"
	"monkey-go/lexer"
	"monkey-go/token"
	"strconv"
)

// =============================================================================
// 业务场景
// 问题 ：Lexer 产生 Token 流，但 Parser 需要将 Token 流转换为 AST（抽象语法树） 。
// 定位 ：这是 Monkey 语言的 语法分析器 ，采用 Pratt Parser （自顶向下运算符优先级解析器）。
// =============================================================================

// 优先级数值越大，表示"绑定越紧密"
// _ = iota 跳过第一个值 0，让 lowest = 1
const (
	_           int = iota // 跳过 iota=0
	lowest                 // 1: 最低优先级（用于初始值）
	equals                 // 2: == !=（相等性比较，优先级较低）
	lessgreater            // 3: < >（大小比较）
	sum                    // 4: + -（加减法）
	product                // 5: * /（乘除法，优先级较高）
	prefix                 // 6: -x !x（前置运算符）
	call                   // 7: func()（函数调用，最高优先级之一）
)

// precedences 运算符优先级表
// 作用：告诉 Parser 遇到某运算符时，应该如何处理左右操作数的解析
//
// 设计思路：
// - "5 + 3 * 2" 应该解析为 "5 + (3 * 2)"，因为 * 优先级高于 +
// - 遇到 + 时，左边已经解析完，右边只有遇到更高优先级的运算符才继续解析
var precedences = map[token.TokenType]int{
	token.EQ:       equals,
	token.NOT_EQ:   equals,
	token.LT:       lessgreater,
	token.GT:       lessgreater,
	token.PLUS:     sum,
	token.MINUS:    sum,
	token.SLASH:    product,
	token.ASTERISK: product,
	token.LPAREN:   call,
	token.LBRACKET: call, // 数组索引 [i] 和函数调用有相同优先级
}

// =============================================================================
// 解析函数类型（Pratt Parser 的函数式解析策略）
// =============================================================================

// 前缀解析函数：解析"以某 Token 开头的表达式"
// 示例：-x（以 MINUS 开头）、!y（以 BANG 开头）、(1+2)（以 LPAREN 开头）
type prefixParseFn func() ast.Expression

// 中缀解析函数：解析"某表达式后面跟着某运算符"
// 示例：1 + 2（+ 是中缀运算符）、add(1, 2)（( 是中缀运算符）
type infixParseFn func(ast.Expression) ast.Expression

// Parser 语法分析器
// 采用 Pratt Parser（自顶向下运算符优先级解析器）
//
// 核心设计：
// - 双 Token 缓冲（curToken 和 peekToken）
// - 函数式解析策略（prefixParseFns, infixParseFns）
// - 错误收集而非立即失败
type Parser struct {
	l              *lexer.Lexer                      // 词法分析器引用
	curToken       token.Token                       // 当前 Token
	peekToken      token.Token                       // 下一个 Token（ lookahead）
	errors         []string                          // 收集解析错误
	prefixParseFns map[token.TokenType]prefixParseFn // 前缀解析函数表
	infixParseFns  map[token.TokenType]infixParseFn  // 中缀解析函数表
}

// New 创建新的 Parser
// 初始化过程：
// 1. 创建 Parser 实例
// 2. 注册所有前缀和中缀解析函数
// 3. 预读两个 Token（填充 curToken 和 peekToken）
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// 初始化解析函数表
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// 预读两个 Token
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken 让 Token 流前进一个位置
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs 检查当前 Token 是否为指定类型
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs 检查下一个 Token 是否为指定类型
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek 期望下一个 Token 是指定类型，如果是则前进并返回 true
// 这是一种"强校验"：语法结构必须严格匹配，否则记录错误
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// ParseProgram 解析整个程序
// 返回 AST 的根节点 Program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// 循环直到遇到 EOF
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// parseStatement 解析语句
// 根据当前 Token 类型分发到具体的语句解析函数
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement 解析 let 语句
// 语法: let <identifier> = <expression>;
//
// 解析步骤：
// 1. 记录 "let" Token
// 2. 读取变量名（必须是标识符）
// 3. 期望 "=" Token
// 4. 解析右侧表达式
// 5. 消费可选的分号
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// 前进到表达式开始的位置
	p.nextToken()

	// 解析赋值表达式
	stmt.Value = p.parseExpression(lowest)

	// 消费可选的分号（如 "let x = 5" 没有分号也合法）
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement 解析 return 语句
// 语法: return <expression>;
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression(lowest)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseExpressionStatement 解析表达式语句
// 语法: <expression>;
// 示例: "5 + 3;" "add(1, 2);" "foobar;"
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(lowest)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseExpression 解析表达式
// 采用 Pratt Parser 算法：利用优先级解决歧义
//
// 核心逻辑：
// 1. 根据前缀规则解析"前缀表达式"（最左边的表达式）
// 2. 循环检查中缀运算符：
//   - 如果中缀运算符优先级 > 当前优先级，解析中缀表达式
//   - 否则，停止循环，返回已解析的左表达式
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// 1. 查找前缀解析函数
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	// 2. 解析前缀表达式（左结合）
	leftExp := prefix()

	// 3. 循环处理中缀表达式（右结合）
	// 条件：下一个 Token 不是分号 AND 当前优先级 < 下一个运算符优先级
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

// parseIdentifier 解析标识符
// 示例: "x", "add", "myVar"
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 解析整数字面量
// 将字符串 "123" 转换为 int64 值
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// parsePrefixExpression 解析前缀表达式
// 示例: -5, !true, --3
//
// 前缀运算符的特点：运算符在操作数前面
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	// 递归解析右侧表达式，优先级为 prefix（较高）
	expression.Right = p.parseExpression(prefix)
	return expression
}

// parseBoolean 解析布尔字面量
// true 或 false
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curToken.Type == token.TRUE}
}

// parseGroupedExpression 解析 grouped 表达式（括号表达式）
// 示例: (1 + 2) * 3
//
// 括号会提升表达式的优先级，强制先求值
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(lowest) // 用 lowest 重置优先级

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

// parseInfixExpression 解析中缀表达式（二元运算符）
// 示例: 1 + 2, 3 * 4, 5 == 6
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	// 用当前运算符的优先级解析右侧
	// 这保证了 "5 * 3 + 2" 被解析为 "(5 * 3) + 2"
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseIfExpression 解析 if 表达式
// 语法: if (<condition>) { <consequence> } else { <alternative> }
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	// 解析条件
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(lowest)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 解析 if 分支
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	// 解析 else 分支（可选）
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

// parseBlockStatement 解析代码块
// 语法: { <statement>* }
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	// 循环解析语句直到遇到右花括号
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

// parseFunctionLiteral 解析函数字面量
// 语法: fn(<parameters>) { <body> }
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()
	return lit
}

// parseFunctionParameters 解析形参列表
//
// 设计注意：形参只能是标识符，不支持表达式
// 合法的形参: fn(a, b, c) {}
// 非法的形参: fn(a + b, 1) {}  // 语法错误
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// 空参数列表
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// 解析剩余参数
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

// parseCallExpression 解析函数调用表达式
// 语法: <expression>(<arguments>)
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

// parseCallArguments 解析实参列表
//
// 设计注意：实参可以是任意表达式，这是与形参的本质区别
// 合法的实参: add(1, 1+2, fn() {}, true, myVar)
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	// 空参数列表
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(lowest))

	// 解析剩余实参
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(lowest))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}

// peekPrecedence 查看下一个 Token 的优先级
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return lowest
}

// curPrecedence 查看当前 Token 的优先级
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return lowest
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// Errors 返回收集的所有错误
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError 记录 Token 类型不匹配的错误
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError 记录无法解析前缀表达式的错误
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// =============================================================================
// 字符串、数组、哈希解析函数
// =============================================================================

// parseStringLiteral 解析字符串字面量
// 语法: "hello world"
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseArrayLiteral 解析数组字面量
// 语法: [<element>, <element>, ...]
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parseExpressionList 解析逗号分隔的表达式列表
// 传入结束 Token，返回表达式列表
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	// 空列表
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(lowest))

	// 解析剩余元素
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(lowest))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseIndexExpression 解析索引表达式
// 语法: <left>[<index>]
// 示例: arr[0], hash["name"]
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(lowest)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// parseHashLiteral 解析哈希字面量
// 语法: { <key>: <value>, ... }
// 示例: {"name": "Tom", "age": 25}
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	// 空哈希
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return hash
	}

	// 解析第一个键值对
	p.nextToken()
	key := p.parseExpression(lowest)
	if !p.expectPeek(token.COLON) {
		return nil
	}
	p.nextToken()
	value := p.parseExpression(lowest)
	hash.Pairs[key] = value

	// 解析剩余键值对
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // 跳过逗号
		p.nextToken() // 前进到 key
		key := p.parseExpression(lowest)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken() // 前进到 value
		value := p.parseExpression(lowest)
		hash.Pairs[key] = value
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}
