package object

import (
	"bytes"
	"fmt"
	"monkey-go/ast"
)

// =============================================================================
// 一、类型系统：定义所有运行时对象的"类型标签"
// =============================================================================

// ObjectType 运行时对象的类型标识
// 作用：用字符串常量标记对象的"种类"，用于类型判断和类型分支
type ObjectType string

// 所有运行时对象类型常量
// 注意：这些是字符串常量，不是 Go 枚举，可在运行时动态比较
const (
	INTEGER_OBJ      = "INTEGER"      // 整数值对象
	BOOLEAN_OBJ      = "BOOLEAN"      // 布尔值对象
	NULL_OBJ         = "NULL"         // 空值对象
	RETURN_VALUE_OBJ = "RETURN_VALUE" // return 语句的返回值包装
	ERROR_OBJ        = "ERROR"        // 运行时错误
	FUNCTION_OBJ     = "FUNCTION"     // 函数对象
)

// =============================================================================
// 二、核心接口：运行时值的抽象
// =============================================================================

// Object 所有运行时值必须实现的接口
// 设计思路：统一所有值的"类型查询"和"字符串表示"行为
type Object interface {
	Type() ObjectType // 返回对象的运行时类型
	Inspect() string  // 返回对象的字符串表示（用于调试输出）
}

// =============================================================================
// 三、基础类型实现
// =============================================================================

// Integer 整数值对象
// 对应 AST 的 IntegerLiteral，存储解析后的整数值
type Integer struct {
	Value int64 // 使用 int64 而非 int，保证跨平台一致性
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

// Boolean 布尔值对象
//
// 设计决策：复用全局单例 TRUE/FALSE，避免重复堆分配
// 使用方式：object.TRUE, object.FALSE，不要 new(Boolean) 实例化
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

// Null 空值对象
// 对应源码中的 null，表示"无值"或"未定义"
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// ReturnValue return 语句的返回值包装对象
//
// 为什么要包装？
// 求值器遇到 return 语句需要"跳出"多层嵌套的函数调用，
// 但 Go 没有结构化跳出机制。使用 ReturnValue 对象作为"信号"，
// 外层求值函数捕获后提取其 .Value
//
// 注意：这是求值器内部机制，不暴露给用户代码
type ReturnValue struct {
	Value Object // return 后面表达式的求值结果
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// Error 运行时错误对象
// 与 Go 的 error 不同，Monkey 的错误是"一等公民"（可以作为值传递）
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// =============================================================================
// 四、全局单例：避免重复分配
// =============================================================================

// 全局单例对象
// 作用：Boolean 和 Null 的值域是固定的（true/false/null），
//
//	使用单例可以避免重复创建相同的对象
//
// 为什么不用 const？
// const 是编译期常量，无法实现 Object 接口（方法需要 receiver）
//
// 正确用法：
//
//	if result == object.TRUE { ... }
//	if result == object.NULL { ... }
var (
	TRUE  = &Boolean{Value: true}  // 布尔 true 单例
	FALSE = &Boolean{Value: false} // 布尔 false 单例
	NULL  = &Null{}                // 空值单例
)

// =============================================================================
// 五、环境：变量作用域管理
// =============================================================================

// Environment 变量环境，实现词法作用域（Lexical Scoping）
//
// 关键设计：链式作用域
// - store: 当前作用域的变量映射
// - outer: 外层作用域指针（形成作用域链）
//
// 查找变量时：从当前作用域开始，沿 outer 链向上查找
type Environment struct {
	store map[string]Object // 当前作用域的变量表
	outer *Environment      // 外层作用域（支持闭包）
}

// NewEnvironment 创建新的顶层环境（无外层作用域）
func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]Object),
		outer: nil,
	}
}

// NewEnclosedEnvironment 创建内层环境（闭包场景）
// 用于函数调用时，创建函数的执行环境
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{
		store: make(map[string]Object),
		outer: outer, // 捕获外层环境，实现闭包
	}
}

// Get 获取变量的值
//
// 查找顺序：
// 1. 先在当前作用域 store 查找
// 2. 找不到则沿 outer 链向上查找
// 3. 链到头都没找到，返回 (nil, false)
//
// 这就是词法作用域的工作原理！
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if ok {
		return obj, true
	}
	// 当前作用域没有，向上查找
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}

// Set 设置变量的值（只在当前作用域设置）
//
// 注意：只在当前作用域设置，不影响外层
// 这与 "var x = 5" 在当前层创建变量一致
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// =============================================================================
// 六、函数对象：闭包的载体
// =============================================================================

// Function 函数字面量的运行时表示
// 当解析 fn(x, y) { x + y } 时，创建此对象
type Function struct {
	Parameters []*ast.Identifier   // 形参列表 ["x", "y"]
	Body       *ast.BlockStatement // 函数体 AST
	Env        *Environment        // 定义时的环境（实现闭包的关键！）
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

// Inspect 返回函数的可读字符串表示
// 格式：fn(参数1, 参数2) { 函数体 }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	// 拼接参数列表
	out.WriteString("fn(")
	for i, p := range f.Parameters {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(p.String())
	}
	out.WriteString(") {\n")

	// 拼接函数体
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}
