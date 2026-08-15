# Monkey Interpreter

一个用 Go 编写的解释器，基于《Writing An Interpreter In Go》实现。

> 本项目源自 Thorsten Ball 的经典著作《Writing An Interpreter In Go》。从零开始构建完整的 Monkey 语言解释器，涵盖词法分析、语法解析、递归求值三大核心模块。

## 项目结构

```
monkey-go-study/
├── ast/           # AST 节点定义（Statement、Expression 接口）
├── evaluator/     # 递归求值器，遍历 AST 执行计算
│   ├── evaluator.go   # 核心求值逻辑
│   └── builtins.go    # 内置函数（len/first/last/rest/push/puts）
├── lexer/         # 词法分析器，将源码切分为 Token 流
├── object/        # 运行时对象（Integer、Boolean、String、Array、Hash、Function 等）
├── parser/        # 语法分析器，Pratt Parser 实现
├── repl/          # 交互式终端入口
├── token/         # Token 类型定义
├── main.go        # 主入口，含调试命令
├── main_test.go   # 主入口测试
└── Makefile       # 构建命令
```

## 快速开始

### 交互式 REPL

```bash
go run main.go
# 或
make run
```

```monkey
>> let add = fn(x, y) { x + y };
>> add(2, 3)
5
>> let fib = fn(x) {
>>   if (x < 2) { return x }
>>   return fib(x-1) + fib(x-2)
>> };
>> fib(10)
55
```

### 调试命令

```bash
go run main.go --lex    # 词法分析：显示 Token 流
go run main.go --parse  # 语法分析：显示 AST 结构
go run main.go --eval   # 求值调试：完整流程 Lexer → Parser → Evaluator
```

### 运行测试

```bash
make test     # 运行所有测试
make bench    # 性能测试
```

## 语言特性

| 特性 | 示例 |
|-----|------|
| **变量绑定** | `let x = 5;` |
| **整数运算** | `+`, `-`, `*`, `/` |
| **布尔运算** | `true`, `false`, `!`, `==`, `!=`, `<`, `>` |
| **条件分支** | `if (x > 0) { x } else { -x }` |
| **函数定义** | `let add = fn(a, b) { a + b };` |
| **函数调用** | `add(1, 2)` → `3` |
| **闭包** | 捕获定义时的环境变量 |
| **递归** | 支持递归函数 |
| **Return 语句** | `return x + 1;` |
| **字符串** | `"Hello" + " " + "World!"` → `"Hello World!"` |
| **数组** | `[1, 2, 3]`, `arr[0]` |
| **哈希表** | `{"name": "Tom", "age": 25}` |
| **内置函数** | `len()`, `first()`, `last()`, `rest()`, `push()`, `puts()` |

## 完整能力清单

本项目实现了一个功能完整的解释器，涵盖：

- ✅ 数学表达式（算术、比较、逻辑运算）
- ✅ 变量绑定 (`let`)
- ✅ 函数定义与调用
- ✅ 高阶函数与函数式编程
- ✅ 闭包（捕获定义时的环境变量）
- ✅ 数据类型：`Integer`、`Boolean`、`String`、`Array`、`Hash`
- ✅ 内置函数：`len`、`first`、`last`、`rest`、`push`、`puts`
- ✅ REPL 交互式终端

## 四层分层架构

```
┌──────────────────────────────────────────────────────┐
│                    源码文本                           │
│              let x = "hello"; arr[0];                │
└──────────────────────┬───────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────┐
│  Layer 1: Lexer（词法分析器）                          │
│  把源码字符串切成最小语义单元 Token 流                  │
│  LET → IDENT "x" → ASSIGN → STRING "hello"           │
└──────────────────────┬───────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────┐
│  Layer 2: AST（抽象语法树）                            │
│  把 Token 流组织成树形结构，表达语法关系                │
│  LetStatement{ Name: Identifier, Value: StringLiteral }│
└──────────────────────┬───────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────┐
│  Layer 3: Object（运行时对象）                         │
│  实际参与计算的对象，包裹宿主语言（Go）类型              │
│  object.String{Value: "hello"}, object.Array{...}    │
└──────────────────────┬───────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────┐
│  Layer 4: Evaluator（递归求值器）                      │
│  遍历 AST，产出 Object，执行运算                        │
│  Eval(StringLiteral) → object.String{Value: "hello"} │
└──────────────────────────────────────────────────────┘
```

## 内置函数详解

| 函数 | 说明 | 示例 |
|-----|------|------|
| `len(x)` | 返回字符串/数组长度 | `len("hello")` → `5` |
| `first(arr)` | 返回数组第一个元素 | `first([1,2,3])` → `1` |
| `last(arr)` | 返回数组最后一个元素 | `last([1,2,3])` → `3` |
| `rest(arr)` | 返回去掉首元素的新数组 | `rest([1,2,3])` → `[2, 3]` |
| `push(arr, x)` | 追加元素返回新数组（不可变） | `push([1,2], 3)` → `[1, 2, 3]` |
| `puts(x)` | 打印输出，返回 null | `puts("hello")` → `null` |

> 所有数组操作遵循不可变原则：`rest` 和 `push` 返回新数组，不修改原数组。

## 学习路线

文档位于 `doc/` 目录，按章节顺序学习：

| 章节 | 内容 | 核心概念 |
|-----|------|---------|
| [01_词法分析器_Lexer.md](doc/01_词法分析器_Lexer.md) | Token 切分、流式读取 | Token、readChar、peekChar |
| [02_语法解析器_Parser.md](doc/02_语法解析器_Parser.md) | Pratt Parser 优先级解析 | precedence、中缀/前缀解析 |
| [03_求值器_Evaluator.md](doc/03_求值器_Evaluator.md) | 递归求值、闭包实现 | Env、extendFunctionEnv、unwrapReturnValue |
| [04_字符串与内置函数.md](doc/04_字符串与内置函数.md) | 字符串、数组、哈希、内置函数 | HashKey、Builtin、Hashable 接口 |

## 设计要点

### 为什么需要四层分离？

| 层次 | 关注点 | 职责 |
|-----|--------|------|
| Token | 文本切分 | 把源码字符串切成最小语义单元 |
| AST | 语法结构 | 把 Token 流组织成树形结构 |
| Object | 运行时存储 | 实际参与计算的对象 |
| Evaluator | 执行逻辑 | 遍历 AST，产出 Object |

好处：每层只做一件事，职责清晰；新增类型只需在各层添加对应处理。

### 内置函数的本质

内置函数是 **Go 函数包装成 Object 对象**，在 Evaluator 的 CallExpression 分支统一处理：

```go
// 内置函数对象
type Builtin struct {
    Fn BuiltinFunction  // Go 原生函数
}

// 标识符查找顺序：变量 → 内置函数
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
    val, ok := env.Get(node.Value)
    if ok { return val }
    if builtin, ok := getBuiltin(node.Value); ok { return builtin }
    return newError("identifier not found: %s", node.Value)
}
```

### HashKey 的设计

Go 的 map key 必须是可比较类型，但 Monkey 对象是指针。需要 `HashKey` 结构体解决指针比较问题：

```go
type HashKey struct {
    Type  ObjectType  // 防止不同类型哈希碰撞
    Value uint64      // 哈希值（fnv-1a 算法）
}

type Hashable interface {
    HashKey() HashKey
}
```

## 参考

- [Writing An Interpreter In Go](https://interpreterbook.com/) - Thorsten Ball
- [Writing A Compiler In Go](https://compilerbook.com/) - Thorsten Ball

---