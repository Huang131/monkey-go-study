package token

// =============================================================================
// 业务场景
// 问题 ：Token 是词法分析的产物，但需要一个统一的结构来承载所有类型的词法单元。
// 定位 ：这是 Monkey 语言的 Token 类型定义 ，是 Lexer 和 Parser 之间的"契约"。
// =============================================================================

// TokenType 用字符串定义所有 Token 类型
// 设计思路：用字符串而非枚举，方便打印和调试
type TokenType string

// Token 词法单元
// 每个 Token 包含：类型（语法含义）和字面量（原始文本）
type Token struct {
	Type    TokenType // 语法类型：LET, IDENT, INT, ...
	Literal string    // 原始文本："let", "x", "5", "+", ...
}

// =============================================================================
// 所有 Token 类型常量
// =============================================================================
const (
	// 非法字符和结束标记
	ILLEGAL = "ILLEGAL" // 无法识别的字符（如 @, # 等）
	EOF     = "EOF"     // 输入结束标记

	// 标识符和字面量
	IDENT  = "IDENT"  // 标识符：变量名、函数名等（如 "x", "add"）
	INT    = "INT"    // 整数字面量（如 "42", "100"）
	STRING = "STRING" // 字符串字面量（如 "hello"）

	// 运算符
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="

	// 分隔符
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
	COLON     = ":"

	// 关键字（保留字）
	// 注意：关键字本质上是特殊的 IDENT，通过 LookupIdent 转换
	FUNCTION = "FUNCTION" // 函数声明
	LET      = "LET"      // 变量声明
	TRUE     = "TRUE"     // 布尔真
	FALSE    = "FALSE"    // 布尔假
	IF       = "IF"       // 条件语句
	ELSE     = "ELSE"     // else 分支
	RETURN   = "RETURN"   // 返回语句
)

// =============================================================================
// 四、关键字表：从字面量到类型的映射
// =============================================================================

// keywords 关键字查找表
// 作用：将标识符的字符串形式转换为关键字 TokenType
var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

// LookupIdent 查找标识符对应的 TokenType
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok // 关键字：let, fn, if, ...
	}
	return IDENT // 普通标识符：变量名、函数名
}
