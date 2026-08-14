# Monkey Interpreter

一个用 Go 编写的儿童友好型解释器，基于《Writing An Interpreter In Go》实现。

## 项目结构

```
monkey-go-study/
├── ast/          # AST 节点定义（Statement、Expression 接口）
├── evaluator/    # 递归求值器，遍历 AST 执行计算
├── lexer/        # 词法分析器，将源码切分为 Token 流
├── object/       # 运行时对象（Integer、Boolean、Function 等）
├── parser/       # 语法分析器，Pratt Parser 实现
├── repl/         # 交互式终端入口
├── token/        # Token 类型定义
├── main.go       # 主入口，含调试命令
├── main_test.go  # 主入口测试
└── Makefile      # 构建命令
```

## 快速开始

### 交互式 REPL

```bash
make run
# 或
go run .
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
make lex      # 词法分析：显示 Token 流
make parse    # 语法分析：显示 AST 结构
make eval     # 求值调试：完整流程 Lexer → Parser → Evaluator
```

### 运行测试

```bash
make test     # 详细输出
make bench    # 性能测试
```

## 语言特性

| 特性 | 示例 |
|-----|------|
| 变量绑定 | `let x = 5;` |
| 整数运算 | `+`, `-`, `*`, `/` |
| 布尔运算 | `true`, `false`, `!`, `==`, `!=`, `<`, `>` |
| 条件分支 | `if (x > 0) { x } else { -x }` |
| 函数定义 | `let add = fn(a, b) { a + b };` |
| 函数调用 | `add(1, 2)` → `3` |
| 闭包 | 捕获定义时的环境变量 |
| 递归 | 支持递归函数 |
| Return 语句 | `return x + 1;` |

## 三层架构

```
┌─────────────────────────────────────────────┐
│                  源码                        │
│            let x = 5 + 3;                   │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Lexer（词法分析器）                          │
│  输入: 源码字符串                             │
│  输出: Token 流                              │
│  源码 → "let" "x" "=" "5" "+" "3" ";"       │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Parser（语法分析器，Pratt Parser）            │
│  输入: Token 流                              │
│  输出: AST（抽象语法树）                       │
│  LetStatement{Name=x, Value=InfixExpression} │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Evaluator（递归求值器）                      │
│  输入: AST                                   │
│  输出: 运行时对象                             │
│  object.Integer{Value: 8}                    │
└─────────────────────────────────────────────┘
```

## 学习路线

1. **01_词法分析器_Lexer.md** - 理解 Token 切分
2. **02_语法解析器_Parser.md** - 理解 Pratt Parser 优先级解析
3. **03_求值器_Evaluator.md** - 理解递归求值与闭包

文档位于 `doc/` 目录。

## 参考

- [Writing An Interpreter In Go](https://interpreterbook.com/) - Thorsten Ball
- [Writing A Compiler In Go](https://compilerbook.com/) - Thorsten Ball