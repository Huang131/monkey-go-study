package lexer

import "monkey-go/token"

// =============================================================================
// 一、核心数据结构：Lexer
// =============================================================================

// Lexer 词法分析器
// 设计思路：流式读取字符，每次 NextToken() 返回一个 Token
//
// 三个指针的作用：
// - position: 当前字符的起始位置（已经读取完毕的字符之后）
// - readPosition: 下一个要读取字符的位置（待读取）
// - ch: 当前字符（position 位置处的字符）
type Lexer struct {
	input        string // 源代码输入
	position     int    // 当前字符位置（已读取）
	readPosition int    // 下一个字符位置（待读取）
	ch           byte   // 当前正在查看的字符
}

// New 创建新的词法分析器
// 初始化后立即调用 readChar() 读取第一个字符
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // 预读第一个字符，准备好 ch
	return l
}

// =============================================================================
// 二、字符读取：流式扫描的核心
// =============================================================================

// readChar 读取下一个字符
//
// 核心逻辑：
// 1. 检查是否到达输入末尾
// 2. 设置 ch 为当前位置的字符
// 3. 移动 position 到当前位置
// 4. 移动 readPosition 指向下一个位置
//
// 为什么用 readPosition 而不是直接 position+1？
// 因为 peekChar() 需要"偷看"下一个字符而不移动位置
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL，表示输入结束
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// peekChar 查看下一个字符（不移动位置）
//
// 作用：用于判断是否是双字符运算符（如 ==, !=, <=, >=）
// 示例：遇到 '=' 时，先 peek 看看下一个是否是 '='
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// =============================================================================
// 三、Token 生成：主循环
// =============================================================================

// NextToken 返回下一个 Token
// 这是 Parser 调用的主要方法，驱动整个词法分析过程
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	// 1. 跳过空白字符（空格、制表符、换行等）
	l.skipWhitespace()

	// 2. 根据当前字符类型生成 Token
	switch l.ch {
	// ----- 双字符运算符（需要 peek 辅助） -----
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar() // 消费第二个 '='
			// 组合成 "==" 并生成 EQ Token
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}

	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.BANG, l.ch)
		}

	// ----- 单字符运算符 -----
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)

	// ----- 分隔符 -----
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)

	// ----- 字符串字面量 -----
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok // readString 已消费完字符，必须提前返回

	// ----- 结束标记 -----
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF

	// ----- 其他字符（标识符或数字） -----
	default:
		// 字母或下划线 → 标识符（可能是关键字）
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			// LookupIdent 会将 "let", "fn" 等关键字转为对应的 TokenType
			tok.Type = token.LookupIdent(tok.Literal)
			return tok // 注意：这里提前返回，因为 readIdentifier 已消费完标识符
		} else if isDigit(l.ch) { // 数字 → 整数字面量
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok // 同上，提前返回
		} else { // 无法识别的字符
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	// 3. 单字符 Token 生成后，需要消费该字符（移动到下一个）
	// 注意：标识符和数字已经在 readIdentifier/readNumber 中消费完了
	l.readChar()

	return tok
}

// newToken 创建单字符 Token 的辅助函数
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// =============================================================================
// 四、标识符和数字读取
// =============================================================================

// readIdentifier 读取连续的字 letter 字符（标识符或关键字）
//
// 示例：输入 "let x = 5"
// 遇到 'l' 时调用 readIdentifier，返回 "let" 并停在空格处
func (l *Lexer) readIdentifier() string {
	startPosition := l.position // 记录起始位置
	for isLetter(l.ch) {        // 只要是字母就继续读
		l.readChar()
	}
	return l.input[startPosition:l.position]
}

// readNumber 读取连续的数字字符
//
// 注意：当前只支持整数，暂不支持浮点数
// 如果需要支持 "3.14"，需要在遇到 '.' 时继续读取小数部分
func (l *Lexer) readNumber() string {
	startPosition := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPosition:l.position]
}

// isLetter 判断是否是有效标识符字符
// 包含：a-z, A-Z, 下划线（支持 snake_case）
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit 判断是否是数字字符
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// skipWhitespace 跳过空白字符
//
// 词法分析阶段忽略空白符，因为它们不是语法的一部分
// 但有些语言（如 Python）依赖缩进表示语法，此时不能跳过换行
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readString 读取字符串字面量（双引号包裹的内容）
//
// 设计：读取双引号之间的所有字符，遇到结束双引号停止
// 注意：当前不支持转义字符，如 "hello\nworld"
func (l *Lexer) readString() string {
	startPosition := l.position + 1 // 跳过开始的双引号
	l.readChar()                    // 移动到第一个字符

	for l.ch != '"' {
		l.readChar()
	}

	// 此时 l.ch == '"'，l.position 指向结束双引号
	// 提取 startPosition 到 position（不含）的内容，即去掉引号
	result := l.input[startPosition:l.position]

	l.readChar() // 跳过结束的双引号，准备读取下一个 token
	return result
}
