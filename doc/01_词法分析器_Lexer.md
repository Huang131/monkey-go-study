# doc/01_词法分析器_Lexer.md
> 章节名称：**第01章 词法分析器：把源代码字节流切割成Token**
> 目标：手写Monkey解释器的词法分析模块，理解Tokenizer完整工作流程、设计思路、踩坑点。

## 1. 本章目标与定位
词法分析器（Lexer / Tokenizer），是解释器/编译器的第一道工序。
输入：源代码原始字符串（字节流）
输出：一组`Token`（源代码最小语义单元），交给下一层语法解析器Parser。

> ⚠重要边界：
> **词法分析器只负责切分token，不做语法合法性校验。**
比如 `5 < 10 > 5` 在语法层面是非法表达式，Lexer依旧正常输出：`INT LT INT GT INT`。
语法是否合法，由后续Parser（AST抽象语法树）负责。

## 2. 整体架构与目录
```
monkey‑go
├── token/
│   └── token.go      // Token结构体、常量定义、关键字映射
├── lexer/
│   ├── lexer.go      // 词法分析器核心实现
│   └── lexer_test.go // TDD单元测试
└── doc/
    └── 01_词法分析器_Lexer.md  # 本文档
```

## 3. 核心数据结构说明

### 3.1 Token（token/token.go）
```go
type Token struct {
	Type    TokenType  // token类型，字符串常量，如 "INT" "EQ" "RETURN"
	Literal string     // 源代码原始文本，原样保存
}
```
- `Type`：告诉上层这是什么类型符号；
- `Literal`：保存原始字符串，数字、标识符、关键字都需要。

关键字处理思路：
维护`keywords map[string]TokenType`，`LookupIdent()`函数做分发。
> 不需要为关键字单独写解析逻辑，**关键字本质是特殊的标识符**。
> 读到一串字母，先当标识符读出来，再查表判断是不是关键字。命中则返回关键字类型，否则返回普通`IDENT`。

### 3.2 Lexer 状态结构体（lexer/lexer.go）
```go
type Lexer struct {
	input        string // 原始源代码字符串
	position     int    // 当前字符下标：指向已经处理完的字符
	readPosition int    // 下一个将要读取的字符下标
	ch           byte   // 当前正在处理的单个字节
}
```
两个指针是理解Lexer的关键：
1. `position`：上一次读取到的位置；
2. `readPosition`：下一次要读的位置，永远 = position + 1；
3. `ch`：缓存当前字节，避免频繁切片访问input。

> 边界：当`readPosition >= len(input)`，说明读到文件末尾，`ch = 0`代表EOF。

### 3.3 两个核心读取函数（区分消费 / 偷看）
1. `readChar()`
    - 读取`readPosition`处字节，赋值给`ch`；
    - 更新`position`、`readPosition`；
    - ✅**消费字符，指针向前移动**。
2. `peekChar()`
    - 只返回`readPosition`位置的字节；
    - ❌**不修改任何指针，不消费字符，只偷看**。
> 使用场景：处理双字符运算符 `==`、`!=`。switch只能匹配单个字节，偷看后面一个字符，判断是否组成双字符符号。

### 3.4 辅助函数清单
|函数|作用|
|---|---|
|`skipWhitespace()`|吃掉空格、`\t`、`\n`、`\r`；空白字符没有语义，只做排版，直接丢弃|
|`readIdentifier()`|读取标识符/关键字，连续字母下划线，遇到非字母停止|
|`readNumber()`|读取连续数字，本项目只支持整数，不支持浮点数|
|`isLetter()`|判断是否是字母/下划线（标识符允许下划线开头）|
|`isDigit()`|判断是否是数字0‑9|

## 4. NextToken：词法分析主循环
`NextToken()`每次调用，返回下一个Token，是上层调用方唯一入口。

### 两条执行路径（重点，极易踩坑）
1. **路径A：单字符符号 `= + - / * < > ! ( ) { } , ;`**
    - switch case内部：只构造token，**不调用readChar消费字符**。
    - switch结束后，执行函数底部统一的 `l.readChar()`，统一消费当前字符，指针前进。
2. **路径B：多字符：标识符 / 数字**
    - `readIdentifier()` / `readNumber()`内部已经在循环里多次调用`readChar()`，指针已经推进到下一个符号。
    - 必须**提前return tok**，**不能走到底部的l.readChar()**。
    > 如果忘记return，会额外多读一个字节，token序列整体错位，出现诡异bug。

> 两种设计思路对比：
> 方案A（本项目采用）：单字符统一在函数底部做readChar，case分支干净；缺点是存在两套路径，要记住多字符必须return。
> 方案B：每个case内部写`readChar()`，去掉底部统一readChar；优点流程统一；缺点大量重复代码，新增case容易忘记写readChar。

### 双字符运算符处理逻辑（`==`、`!=`）
以`case '='`举例：
1. 当前读到`=`；
2. `peekChar()`偷看后面字节；
3. 如果下一个也是`=`：
    - 手动调用`readChar()`吃掉第二个`=`；
    - 构造`EQ("==")`token；
4. 否则就是普通赋值号`ASSIGN("=")`。

> `!=`逻辑完全同理。本项目只有这两组双字符符号，直接复制代码，没有抽象工具函数；如果后续扩展`>= <=`，建议抽公共辅助函数。

### EOF处理
当`l.ch == 0`，返回`token.EOF`，**绝对不能调用readChar()**，否则数组越界panic。

## 5. TDD 测试驱动开发实践
本模块全程TDD流程：
1. 编写测试输入字符串，写好预期token列表；
2. 运行测试，观察测试失败；
3. 修改生产代码，直到单元测试全部通过。

> 好处：后续修改、重构lexer，只要跑单元测试，就可以快速判断是否破坏原有逻辑。
运行命令：
```bash
go test ./lexer -v
```

## 6. 当前Lexer支持能力
✅ 单字符运算符：`! - / * < > + = ( ) { } , ;`
✅ 双字符运算符：`==`、`!=`
✅ 关键字：`let fn if else return true false`
✅ 用户标识符
✅ 整数数字
✅ 自动跳过空白字符 ` ` `\t` `\n` `\r`

❌ 暂未实现：字符串字面量、浮点数、注释、负数字面量。

## 7. 高频踩坑清单（复习重点）
1. 多字符读取函数`readIdentifier/readNumber`内部已经推进指针，NextToken中必须提前return，不能走到底部`l.readChar()`；否则token错位，丢失符号。
2. EOF状态禁止调用`readChar()`，会数组越界panic。
3. switch case只能匹配单个字节，双字符运算符必须依靠`peekChar()`偷看预判；匹配双字符后，必须手动调用一次`readChar()`消费第二个字节。
4. 关键字不需要新增case分支，复用标识符读取逻辑，依靠`keywords map + LookupIdent`做区分。
5. `ILLEGAL`非法字符：生成ILLEGAL token之后，会走到底部执行readChar，消费掉这个非法字节，继续向下解析。

## 8. 本章结束之后，下一个阶段是什么？
词法分析完成，输出token流。
下一步进入**语法解析器Parser**：
1. 定义AST抽象语法树节点；
2. 实现递归下降解析器；
3. 实现前缀解析函数、中缀解析函数，处理表达式优先级；
4. 把token序列翻译成AST。

---

> 文档使用说明：
> 1. 一边看本文的思路、概念，一边对照`token/token.go`、`lexer/lexer.go`、`lexer_test.go`阅读；
> 2. 接手项目的人，可以先阅读本文，理解整体设计意图，再阅读代码，避免只看代码看不懂设计思路。