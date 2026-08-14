# doc/02_语法解析器_Parser.md
> 章节名称：**第02章 语法解析器：Token流构建AST抽象语法树**
> 目标：手写Monkey解释器Parser模块，理解Pratt优先级解析+递归下降工作流程、设计思路、踩坑点。
> 完整实现代码请查阅仓库对应源码文件：`ast/ast.go`、`parser/parser.go`、`parser/parser_test.go`、`repl/repl.go`。

## 1. 本章目标与定位
语法解析器 Parser，位于解释器第二道工序。
输入：Lexer输出的Token序列
输出：`*ast.Program` AST抽象语法树根节点；同时收集语法错误列表。

> ⚠重要边界：
> 1. Parser**只做语法解析，生成AST，不执行、不求值**；代码执行逻辑交给后续Evaluator模块。
> 2. Lexer只切Token，不校验语法；Parser负责校验语法结构是否合法。
> 3. 本项目采用 **递归下降 + Pratt优先级解析**，专门用来处理表达式运算符优先级问题。

整体流水线：
`源代码字符串 → Lexer → Token流 → Parser → AST抽象语法树`

## 2. 整体架构与目录
```
monkey‑go
├── token/
│   └── token.go         // Token类型常量
├── lexer/
│   └── lexer.go         // 词法分析器
├── ast/
│   └── ast.go           // AST全部节点接口与结构体定义
├── parser/
│   ├── parser.go        // Parser核心实现
│   └── parser_test.go   // Parser单元测试，TDD表驱动测试
├── repl/
│   └── repl.go          // 交互式终端，接入Parser
└── doc/
    ├── 01_词法分析器_Lexer.md
    └── 02_语法解析器_Parser.md  # 本文档
```

## 3. AST抽象语法树前置知识（ast/ast.go）
> AST把线性Token流，转化成树形结构。
两个核心概念：
- `Statement` 语句：**不产生返回值**，例如 `let`、`return`。
- `Expression` 表达式：**一定会产生一个值**，`1+2`、`add(1,2)`、变量`x`都属于表达式。

### 核心接口片段
```go
// Node 所有AST节点统一接口
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement 语句
type Statement interface {
	Node
	statementNode()
}

// Expression 表达式
type Expression interface {
	Node
	expressionNode()
}

// Program 整个程序根节点，保存全部语句
type Program struct {
	Statements []Statement
}
```

关键结构体（节选）
- `LetStatement`：`let x = expr;`，`Name`标识符，`Value`为右侧表达式
- `ReturnStatement`：`return expr;`，`ReturnValue`为返回的表达式
- `ExpressionStatement`：把表达式包装成语句，例如 `1+2;`
- `Identifier`：标识符变量
- `IntegerLiteral / Boolean`：字面量
- `PrefixExpression`：前缀表达式 `!true`、`-10`
- `InfixExpression`：中缀二元表达式 `a + b`、`a == b`
- `IfExpression`：if条件表达式，Monkey中if是表达式，可以返回值
- `FunctionLiteral`：函数字面量 `fn(a,b){...}`
- `CallExpression`：函数调用 `add(1,2)`

> 完整结构体定义请查看 `ast/ast.go`。

## 4. Parser核心数据结构（parser/parser.go）
### Parser结构体
```go
type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token  // 当前正在处理的token
	peekToken token.Token  // 预读下一个token，LL(2)向前看一个token做决策
	errors    []string     // 收集语法错误，不panic

	prefixParseFns map[token.TokenType]prefixParseFn // 前缀解析函数表
	infixParseFns  map[token.TokenType]infixParseFn  // 中缀解析函数表
}
```

### 解析函数类型
```go
// 前缀解析：没有左侧表达式，直接产出表达式
type prefixParseFn func() ast.Expression

// 中缀解析：已经拿到左侧表达式，接收left，返回组装完成的表达式
type infixParseFn func(left ast.Expression) ast.Expression
```

### 运算符优先级常量与映射表
```go
const (
	lowest int = iota
	equals      // == !=
	lessgreater // < >
	sum         // + -
	product     // * /
	call        // 函数调用，LPAREN，最高优先级
)

var precedences = map[token.TokenType]int{
	token.EQ:       equals,
	token.NOT_EQ:   equals,
	token.LT:       lessgreater,
	token.GT:       lessgreater,
	token.PLUS:     sum,
	token.MINUS:    sum,
	token.ASTERISK: product,
	token.SLASH:    product,
	token.LPAREN:   call,
}
```

> ⭐重点：`token.LPAREN` 左括号拥有双重语义
1. **prefix场景**：`(1+2)`，表达式开头遇到左括号，作为分组表达式。
2. **infix场景**：`add(1,2)`，已经解析完左侧表达式，peek看到左括号，作为函数调用。
> 实现手段：同时在`prefixParseFns`、`infixParseFns`两张表注册LPAREN，call优先级最高。

### 工具函数说明
|函数|作用|
|---|---|
|`nextToken()`|推动curToken、peekToken向前移动|
|`curTokenIs(t)`|判断当前token类型|
|`peekTokenIs(t)`|判断预读token类型，不消费token|
|`expectPeek(t)`|校验peek是否为目标token；匹配则消费，不匹配记录错误|
|`peekPrecedence()`|获取peekToken的运算符优先级|
|`curPrecedence()`|获取curToken的运算符优先级|
|`Errors()`|返回收集到的语法错误列表|

## 5. Pratt优先级解析核心原理
Pratt解析专门解决表达式优先级、结合性问题，核心入口函数 `parseExpression(precedence int)`。

### 核心代码片段
```go
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// 第一步：执行前缀解析，拿到左侧表达式
	left := p.prefixParseFns[p.curToken.Type]()

	// 第二步循环：peek运算符优先级高于传入precedence，则执行中缀解析
	for !p.curTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			break
		}
		p.nextToken()
		left = infix(left)
	}
	return left
}
```

关键逻辑拆解：
1. 先跑prefix：处理标识符、数字、`!`、`-`、`(`、`if`、`fn`；
2. 循环条件 `precedence < p.peekPrecedence()` 是Pratt灵魂；
    - 如果后面运算符优先级更高：调用infix函数，把left作为左操作数，生成新left；
    - 如果后面运算符优先级更低/相等：直接退出循环，返回left；
3. 不需要手写大量if‑else判断运算符优先级，全部由precedences表驱动。

## 6. 主要模块设计思路
> 完整实现查看 `parser/parser.go`

### 6.1 程序入口 ParseProgram
循环读取token，调用`parseStatement()`解析每一条语句，直到EOF，组装`*ast.Program`。

### 6.2 语句解析 parseStatement
```go
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
```

#### 历史Bug
早期版本 `parseLetStatement`、`parseReturnStatement`，识别关键字之后，**没有调用parseExpression解析右侧表达式**，直接循环读到分号，导致`Value = nil`。
修复方案：
```go
stmt.Value = p.parseExpression(lowest)        // let x = expr
stmt.ReturnValue = p.parseExpression(lowest) // return expr
```
> `lowest`代表最低优先级，解析完整整条表达式。

### 6.3 prefix前缀解析函数表
| TokenType | 解析函数 | 用途 |
|---|---|---|
| IDENT | parseIdentifier | 变量标识符 |
| INT | parseIntegerLiteral | 整数字面量 |
| TRUE/FALSE | parseBoolean | 布尔字面量 |
| BANG / MINUS | parsePrefixExpression | 前缀运算符 `!` `-` |
| LPAREN | parseGroupedExpression | 分组表达式 `(expr)` |
| IF | parseIfExpression | if条件表达式 |
| FN | parseFunctionLiteral | 函数字面量 |

### 6.4 infix中缀解析函数表
| TokenType | 解析函数 | 用途 |
|---|---|---|
| +‑ * / == != < > | parseInfixExpression | 二元算术、比较运算 |
| LPAREN | parseCallExpression | 函数调用表达式 |

> 注意：函数调用解析中，**左侧已经解析完成的left表达式作为Function**；再解析实参列表。

### 6.5 形参 vs 实参区别
1. `parseFunctionParameters()`：解析函数**形参列表**，只允许标识符；
2. `parseCallArguments()`：解析函数**实参列表**，可以传入任意复杂表达式。

### 6.6 分号处理
Monkey语言分号`;`可选。解析语句末尾，只做peek判断，如果遇到SEMICOLON，才消费token，没有分号也合法。

## 7. TDD 测试驱动开发（parser/parser_test.go）
Parser全程TDD开发模式：
1. 编写输入源码字符串；
2. 预期AST节点结构；
3. 执行解析，断言AST；测试失败 → 修改生产代码 → 全部通过。

关键测试用例：
- `TestLetStatements`：let语句表驱动测试；复现早期Value=nilbug；
- `TestReturnStatements`：return返回表达式解析；
- `TestIdentifierExpression`：标识符表达式；
- `TestOperatorPrecedenceParsing`：运算符优先级，Pratt解析核心测试；
- `TestIfExpression`：if‑else表达式；
- `TestFunctionLiteralParsing`：函数字面量形参解析；
- `TestCallExpressionParameterParsing`：函数调用实参解析。

运行测试命令：
```bash
go test ./parser -v
```

## 8. REPL接入Parser（repl/repl.go）
> 早期REPL只调用Lexer，打印token；改造之后接入Parser。

执行流程：
1. 读取终端一行输入；
2. 构造lexer，再构造parser；
3. 调用`p.ParseProgram()`得到AST；
4. 如果`len(p.Errors())>0`：打印猴子ASCII图案 + parser语法错误；
5. 无错误：调用`program.String()`输出AST文本。

关键片段：
```go
line := scanner.Text()
l := lexer.New(line)
p := parser.New(l)

program := p.ParseProgram()
if len(p.Errors()) != 0 {
	printParserErrors(out, p.Errors())
	continue
}
io.WriteString(out, program.String())
```

> ⚠当前REPL仅打印AST，**不会计算执行结果**。完整代码查看`repl/repl.go`。

## 9. 踩坑记录 & 面试高频考点
1. **LPAREN双重语义**：必须同时注册prefix、infix解析函数，并且设置call最高优先级；只注册一边会直接解析失败。
2. 历史bug：`parseLetStatement`、`parseReturnStatement`忘记调用`parseExpression(lowest)`，Value字段为nil。
3. Pratt解析核心循环条件 `precedence < p.peekPrecedence()`，控制运算符优先级与结合性。
4. `curToken / peekToken`双token预读，LL(2)；`expectPeek`做语法校验，不匹配则记录错误，不panic。
5. 区分Statement与Expression：Statement不产生值，Expression产生值；let语句的Value字段是Expression，支持任意复杂表达式。
6. 函数**形参只能是标识符**；**实参可以是任意表达式**。
7. 分号`;`可选，仅peek到分号时才消费token。
8. Parser不panic，错误收集到errors切片，交给上层REPL展示，方便交互式使用。

## 10. 当前阶段边界 & 下一章节预告
✅已完成：
1. Lexer词法分析切割Token；
2. AST全部节点定义；
3. Pratt+递归下降Parser完整实现；
4. 全套单元测试，TDD开发；
5. REPL接入Parser，输出AST、展示语法错误。

❌当前缺失：**没有解释执行逻辑，不会计算表达式结果**。

> 下一章节：Evaluator求值器
1. `Environment`环境对象，维护变量名与值绑定；
2. 递归遍历AST各个节点，做解释求值；
3. 实现整数、布尔、标识符、中缀运算、if表达式、函数调用执行；
4. REPL升级：不再打印AST，输出Evaluator执行运算结果。

---

> 文档使用说明：
> 1. 阅读本文理解设计思想，完整实现细节查阅仓库对应`.go`源码文件；
> 2. 修改代码之后，优先跑单元测试，快速回归校验Parser逻辑正确性。