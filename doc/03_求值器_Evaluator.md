# eval-01：求值器开篇与整体架构

## 一、三层流水线（核心铁律）
```
源码字符串 → Lexer分词 → Parser建AST → Evaluator递归求值执行
```
- **Lexer**：只切字符串为Token，无语法、无计算、无优先级
- **Parser**：Token→AST，处理优先级/括号/语法规则，**只构建树，不运行**
- **Evaluator**：递归遍历AST，真正执行，产出运行时对象

> Parser决定"树长什么样"，Evaluator决定"树怎么跑"。

## 二、两大体系严格分离（新手最大误区）
| | AST静态节点 `ast.XXX` | Object运行对象 `object.XXX` |
|---|---|---|
| 阶段 | 编译期 | 运行期 |
| 作用 | 记录代码写法 | 参与运算/存储/控制流 |
| 示例 | `ast.IntegerLiteral` | `object.Integer` |

> `ast.IntegerLiteral`只是代码上写的数字；`object.Integer`是内存中参与计算的实例。

## 三、架构选型：树遍历解释器
- 优点：递归逻辑直观，适合学习
- 缺点：无字节码缓存，每次执行重复遍历AST，性能弱
- 工业级：字节码VM / JIT

## 四、两大基石
### 1. 统一递归入口
```go
func Eval(node ast.Node, env *object.Environment) object.Object
```
内部类型断言分支处理；env保存变量绑定实现作用域。

### 2. 统一Object接口
```go
type Object interface {
    Type() ObjectType   // 运行时类型判断
    Inspect() string    // REPL打印输出
}
```

## 五、遍历算法：后序遍历（表达式求值唯一解）
- 必须先子节点后父节点，运算符等操作数全部算完才能执行
- 前序/中序不能实现表达式计算
- Parser组装树形=编码优先级；Evaluator后序遍历=执行运算

## 六、evaluator.go 初始骨架
```go
func Eval(node ast.Node, env *object.Environment) object.Object {
    switch node := node.(type) {
    case *ast.Program:
        return evalProgram(node, env)
    case *ast.IntegerLiteral:
        return &object.Integer{Value: node.Value}
    case *ast.BooleanLiteral:
        return nativeBoolToBooleanObject(node.Value)
    case *ast.PrefixExpression:
        right := Eval(node.Right, env)
        return evalPrefixExpression(node.Operator, right)
    case *ast.InfixExpression:
        left := Eval(node.Left, env)
        right := Eval(node.Right, env)
        return evalInfixExpression(node.Operator, left, right)
    case *ast.BlockStatement:
        return evalBlockStatement(node, env)
    case *ast.IfExpression:
        return evalIfExpression(node, env)
    default:
        return nil
    }
}
```

## 七、全局单例优化
- `TRUE`/`FALSE`/`NULL`：全局唯一单例，全程序复用指针
- `Integer`：数值无穷多，每次运算新建对象

---

# eval-02：表达式求值体系

## 一、前缀表达式 `!`、`-`
- `!`取反：基于Monkey真值规则，真值→false，假值→true
- `-`负号：**仅支持整型**，其他类型返回错误
- 执行逻辑：先递归求右操作数，再执行一元运算

```go
func evalPrefixExpression(operator string, right object.Object) object.Object {
    switch operator {
    case "!":
        return evalBangOperatorExpression(right)
    case "-":
        return evalMinusPrefixOperatorExpression(right)
    default:
        return newError("unknown operator: %s%s", operator, right.Type())
    }
}
```

## 二、中缀表达式
### 算术 `+ - * /`
- 只允许整型之间运算，返回`*object.Integer`
- 类型不匹配返回错误

### 大小比较 `< >`
- 只支持整型，返回布尔

### 相等判断 `== !=`
- **Boolean**：直接对比单例指针，不需要取值
- **Integer**：必须取出内部Value做数值比较，不能比指针
- 类型不一致直接返回false

## 三、Monkey真值规则
**只有 `false`、`null` 视为假；数字、字符串、函数全部视为真。**
（类JavaScript，区别于Go/Rust严格布尔）

## 四、关键辅助函数
```go
func nativeBoolToBooleanObject(input bool) *object.Boolean {
    if input { return object.TRUE }
    return object.FALSE
}
```

## 五、evalInfixExpression 结构
```go
func evalInfixExpression(operator string, left, right object.Object) object.Object {
    switch {
    case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
        return evalIntegerInfixExpression(operator, left, right)
    case operator == "==":
        return nativeBoolToBooleanObject(left == right) // 单例指针比较
    case operator == "!=":
        return nativeBoolToBooleanObject(left != right)
    case left.Type() != right.Type():
        return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
    default:
        return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
    }
}
```

---

# eval-03：控制流（If表达式 + Return嵌套BUG）

## 一、If表达式设计
Monkey中`if`是**表达式**不是语句，有返回值可直接赋值：
```monkey
let a = if(1>2){10}else{20};
```
- `Consequence`：if成立执行块
- `Alternative`：else执行块，nil代表无else

### 为什么if设计成表达式
1. Monkey"一切皆表达式"设计哲学
2. 实现简化：不需要同时维护IfStatement+IfExpression两套AST/Parser/Eval逻辑
3. 单独一行使用时，Parser套一层`ExpressionStatement`把表达式当语句执行
4. 代价：没有else if语法糖，只能嵌套

### 为什么let/return不做成表达式
- let核心是变量绑定动作，强行做表达式需规定返回值语义，认知负担大，收益极低
- return核心是控制流信号（终止函数），不是计算普通值，做表达式语义混乱
- if本质是根据条件选择计算哪个分支的值，天然是计算，高频参与赋值传参

## 二、Return控制流核心设计
需求：return需立刻终止当前块，跨多层嵌套传递终止信号。
解决方案：引入包装信号对象
```go
type ReturnValue struct {
    Value Object // 真正返回的数据
}
```
核心目的：区分普通表达式结果 和 return强制终止信号。

## 三、Return嵌套致命BUG（高频面试考点）
### 复现场景
```monkey
if (10 > 1) {
    if (10 > 1) { return 10 }
    return 1
}
```
预期10，错误实际输出1。

### BUG根源
内层处理block时**提前解包ReturnValue外壳**，只返回裸数值，return终止信号标记丢失。
```go
// ❌错误
if returnVal, ok := result.(*object.ReturnValue); ok {
    return returnVal.Value // 信号壳丢失
}
```

### 错误链路
1. 内层return 10 → ReturnValue{Value:10}带壳
2. 内层错误解包 → 裸整数10，终止标记丢失
3. 外层Block拿到普通Integer，识别不到return信号，循环继续
4. 执行外层return 1 → 最终返回1

### 正确修复
**所有嵌套代码块只透传ReturnValue外壳，绝对不解包；只有最顶层evalProgram才允许解包。**
```go
// ✅正确
if returnVal, ok := result.(*object.ReturnValue); ok {
    return returnVal // 原样透传信号壳
}
```

## 四、evalBlockStatement vs evalProgram 分离
```go
// 嵌套块：只透传，不解包
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
    var result object.Object
    for _, stmt := range block.Statements {
        result = Eval(stmt, env)
        if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
            return result // 不解包
        }
    }
    return result
}

// 顶层Program：唯一允许解包的地方
func evalProgram(program *ast.Program, env *object.Environment) object.Object {
    var result object.Object
    for _, stmt := range program.Statements {
        result = Eval(stmt, env)
        if returnObj, ok := result.(*object.ReturnValue); ok {
            return returnObj.Value // 唯一解包点
        }
    }
    return result
}
```

### 职责区分
- `Program`：整个源码根节点，顶层入口，负责解包return信号
- `BlockStatement`：`{}`代码块嵌套执行环境，只透传信号，禁止解包

> 函数调用场景：函数体内也是BlockStatement只透传，在`applyFunction`内部解包。

---

# eval-04：错误处理体系

## 一、设计思想
`*object.Error` 和 `*object.ReturnValue` **设计模式完全相同**，都是包装对象：
- 正常计算 → 普通对象
- return控制流 → ReturnValue{Value}
- 运行时错误 → Error{Message}

一旦返回`*object.Error`，上层不做继续计算，直接一路向上透传。

## 二、Error对象定义（object.go）
```go
const ERROR_OBJ = "ERROR"

type Error struct {
    Message string
}
func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "error: " + e.Message }
```

## 三、newError 辅助函数
```go
func newError(format string, a ...interface{}) *object.Error {
    return &object.Error{Message: fmt.Sprintf(format, a...)}
}
```

## 四、各求值分支改造（不再返回nil）
- `evalPrefixExpression` default → `newError("unknown operator: %s%s", ...)`
- `evalMinusPrefixOperatorExpression` 非整型 → `newError("unknown operator: -%s", ...)`
- `evalInfixExpression` 类型不一致 → `newError("type mismatch: %s %s %s", ...)`
- `evalInfixExpression` default → `newError("unknown operator: %s %s %s", ...)`
- `evalIntegerInfixExpression` default → 同上

## 五、IsError 工具函数
```go
func IsError(obj object.Object) bool {
    if obj == nil { return false }
    return obj.Type() == object.ERROR_OBJ
}
```

## 六、错误透传规则
### evalProgram
```go
switch result.Type() {
case object.RETURN_VALUE_OBJ:
    return result.(*object.ReturnValue).Value // 解包
case object.ERROR_OBJ:
    return result // 不解包，原样透传
}
```

### evalBlockStatement
```go
rt := result.Type()
if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
    return result // 两种终止信号都立刻终止块，向上透传
}
```

### 所有求值点必须增加错误拦截
```go
left := Eval(node.Left, env)
if IsError(left) { return left }
right := Eval(node.Right, env)
if IsError(right) { return right }
```
> 工程规则：只要调用Eval得到返回值，第一件事IsError判断，出错直接return，绝对不要再访问对象内部字段否则panic。

## 七、错误传播链路示例（`5 + true`）
1. Eval中缀 → 左右类型不等 → newError返回`*object.Error`
2. 上层IsError命中 → 直接向上return
3. Block内evalBlockStatement识别ERROR_OBJ → 终止循环向上透传
4. 顶层evalProgram case ERROR_OBJ → 原样返回
5. REPL调用Inspect() → 打印`error: type mismatch: integer + boolean`

## 八、旧代码返回nil的致命缺陷
- nil无法携带错误文本
- 上层无法区分正常nil和求值出错
- 错误静默传播继续执行，产生诡异结果

## 九、Error不在evalProgram解包的原因
Error本身携带完整消息就是最终输出对象，没有内部`.Value`字段不需要解包；ReturnValue是包装壳真实结果存在`.Value`所以需要解包。

---

# eval‑05‑let变量绑定与Environment环境
> 包含：Environment完整定义、let语句求值、Identifier求值、标识符两种AST位置辨析、测试要点、闭包前置知识。

## 一、核心背景
Parser只负责校验语法、生成AST；**变量存储、变量查找全部由 Evaluator + Environment 完成**。
`let x = 5` 做两件事：
1. 对 `=` 右侧表达式递归求值；
2. 将变量名和运行时对象存入环境。

> 同一个结构体`ast.Identifier`，在AST树上位置不同语义完全不同：
> 1. 在`LetStatement.Name`：**变量定义（写），调用`env.Set`，不走`evalIdentifier`**
> 2. 在表达式位置：**变量读取（读），调用`evalIdentifier`，执行`env.Get`**

示例
```monkey
let a = 10;   // LetStatement.Name：定义变量，写环境
let b = a + 5;// a 处于表达式位置：读取变量，查询环境
b;            // b 处于表达式位置：读取变量，查询环境
foobar;       // 表达式位置，找不到，返回错误 identifier not found: foobar
```

| AST节点位置 | Identifier行为 | 底层逻辑 |
|---|---|---|
| `LetStatement.Name` | 变量定义 | `env.Set()`，不执行evalIdentifier |
| 表达式位置（Infix、ExpressionStatement等） | 变量读取 | `evalIdentifier(env.Get())` |

## 二、Environment 环境对象（object.go）
> Environment是求值**上下文载体，不实现Object接口，不属于Monkey语言运行时值**，不能被赋值、不能作为函数返回值。
> `outer *Environment` 父环境指针，后续实现闭包核心，形成链表实现词法作用域。

```go
type Environment struct {
	store map[string]object.Object // 当前层变量存储
	outer *Environment             // 外层父环境指针
}

// 创建顶层环境，outer为nil
func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]object.Object),
		outer: nil,
	}
}

// 创建子封闭环境，继承传入的outer父环境，函数调用时使用
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{
		store: make(map[string]object.Object),
		outer: outer,
	}
}

// Set：仅写入当前层store，不会修改外层环境
func (e *Environment) Set(name string, val object.Object) object.Object {
	e.store[name] = val
	return val
}

// Get：优先本层查找；找不到顺着outer向外递归查找；全部找不到返回 ok=false
func (e *Environment) Get(name string) (object.Object, bool) {
	obj, ok := e.store[name]
	if ok {
		return obj, true
	}
	// 递归向外找父环境
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}
```

### Environment使用时机
1. **REPL全局**：只初始化一次`NewEnvironment()`，循环复用，用户多次输入共享变量状态。
2. **单元测试**：每一个测试case调用`NewEnvironment()`，测试之间状态隔离，避免互相污染。
3. **函数调用**：调用函数时使用`NewEnclosedEnvironment(function.Env)`生成子环境，outer指向**函数定义时捕获的环境**，实现闭包。
4. ⚠️ Monkey无块级作用域：普通`{}`代码块`BlockStatement`不会新建Environment，块内let会污染外层环境；只有函数调用才新建子环境。

## 三、Eval函数签名变更（evaluator.go）
> 所有递归求值都要透传env上下文。

```go
// 修改前
// func Eval(node ast.Node) object.Object

// 修改后：增加env *object.Environment参数
func Eval(node ast.Node, env *object.Environment) object.Object
```

受影响需要同步修改签名并透传env的内部函数
```go
func evalProgram(node *ast.Program, env *object.Environment) object.Object
func evalBlockStatement(node *ast.BlockStatement, env *object.Environment) object.Object
func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object
```

### repl.go REPL改造
> REPL全局环境一次创建，循环复用，保留用户输入的变量。
```go
func Start() {
	scanner := bufio.NewScanner(os.Stdin)
	env := object.NewEnvironment() // REPL全局环境
	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			return
		}
		input := scanner.Text()
		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()
		// 求值，传入全局env
		result := evaluator.Eval(program, env)
		if result != nil {
			fmt.Println(result.Inspect())
		}
	}
}
```

### evaluator_test.go 测试工具函数改造
> **每个测试用例独立全新环境，测试之间状态隔离，避免耦合**
```go
func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment() // 每次测试全新环境
	return evaluator.Eval(program, env)
}
```

## 四、LetStatement 语句求值实现 evaluator.go
Eval入口switch增加case
```go
case *ast.LetStatement:
	val := Eval(node.Value, env)
	if IsError(val) {
		return val // 右侧表达式报错，直接透传错误，不执行Set
	}
	env.Set(node.Name.Value, val)
	// let是语句，Monkey中let执行完毕无返回值，返回nil
	return nil
```

## 五、Identifier 标识符求值实现 evaluator.go
> 只处理**表达式位置**的标识符，读取环境变量。

Eval入口switch增加case
```go
case *ast.Identifier:
	return evalIdentifier(node, env)
```

```go
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: %s", node.Value)
	}
	return val
}
```

## 六、测试用例 TestLetStatements
```monkey
// case1
let a=5; a; → 5
// case2
let a=5*5; a; →25
// case3
let a=5; let b=a; b; →5
// case4
let a=5; let b=a; let c=a+b+5; c; →15
// case5 错误用例
foobar; → error: identifier not found: foobar
```

## 七、REPL实际运行效果
```monkey
>> let a = 5;
>> let b = a > 3;
>> let c = a * 99;
>> a
5
>> b
true
>> c
495
```

## 八、闭包前置原理（尚未实现函数调用，理论铺垫）
```monkey
let a = 10;
let f = fn() { return a; };
```
1. 定义函数f时，`object.Function`保存`Env`指向**定义时刻的全局环境**；
2. 调用f()时，`NewEnclosedEnvironment(f.Env)`创建子环境，outer指向f捕获的环境；
3. 函数内部读取标识符`a`：子环境找不到，顺着outer向外递归查找到全局a=10。

> 关键点：闭包捕获的是**定义时的环境，不是调用时环境**；Environment依靠outer链表实现词法作用域，不是调用栈。

## 九、当前完成清单
✅ Environment结构体：store、outer、NewEnvironment、NewEnclosedEnvironment、Get、Set
✅ Eval签名增加env，全链路透传上下文
✅ REPL全局环境；单元测试每次新建独立环境
✅ `LetStatement`求值：右侧表达式求值，错误拦截，env.Set绑定变量
✅ `Identifier`求值（表达式位置）：env.Get读取变量，未定义返回运行时错误

## 十、下一节待实现内容 eval‑06
1. `object.Function`运行时对象，保存Parameters、Body、捕获的Env（闭包核心）
2. `*ast.FunctionLiteral`函数字面量求值，构造Function对象，**不执行函数体**
3. 后续再实现函数调用：`CallExpression`，参数绑定、新建子环境执行函数体

## 十一、面试思考题
1. REPL共用同一个env，单元测试每次新建env，策略为什么不一样？
> REPL：用户需要跨输入保留变量状态；单元测试：每个case必须完全隔离，不能残留上一个测试的变量，防止测试互相干扰。

2. Environment为什么不实现`object.Object`接口？
> Environment是求值上下文，不是Monkey语言的值，不能赋值、不能作为函数返回值，不属于运行时对象。

3. Get方法递归outer查找，Set为什么只写当前层不向外写？
> 符合词法作用域语义：let只会在当前作用域新增/覆盖变量，不会修改外层同名变量。

4. Monkey为什么没有块级作用域？
> 原版实现简化，只有函数调用才会新建子Environment；普通`{}`块复用传入的env，块内let会污染外层。

# eval‑06‑function对象
> 项目：Writing An Interpreter In Go｜面向新手教学

## 一、前置回顾
上一节 eval‑05 完成变量环境 `Environment`，依靠 `outer` 链表实现词法作用域。
现在进入函数字面量：`fn(x) { x + 2 }`。

> 重要区分：
> - `ast.FunctionLiteral`：AST静态节点，只记录源码语法
> - `object.Function`：运行时对象，函数在内存里的实例，闭包核心载体

Parser只生成AST；**定义函数阶段不执行函数体，仅仅构造运行时Function对象**。只有调用 `f()` 的时候才跑函数内部代码。

## 二、ast.FunctionLiteral AST结构
```go
type FunctionLiteral struct {
	Token      token.Token
	Parameters []*ast.Identifier // 参数列表AST
	Body       *ast.BlockStatement // 函数体AST代码块
}
```
AST只存语法信息，**没有环境Env字段**，AST是无状态的。

## 三、object.go 新增 Function 运行时对象
```go
const FUNCTION_OBJ = "FUNCTION"

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment // ⭐捕获【函数定义那一刻】的环境，闭包核心
}

// 实现 Object 接口
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

// Inspect：REPL打印函数对象文本
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := make([]string, 0)
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}
```

### 字段说明
1. `Parameters`：直接复用AST的参数标识符；
2. `Body`：保存函数体AST，调用的时候才拿来Eval；
3. `Env`：保存函数定义时的环境引用，**不是调用时的环境**。

> 新手坑：不要在这里拷贝环境，只保存指针引用。

## 四、evaluator.go：处理 *ast.FunctionLiteral
Eval 主switch增加分支
```go
case *ast.FunctionLiteral:
	return &object.Function{
		Parameters: node.Parameters,
		Body:       node.Body,
		Env:        env, // 当前正在求值的env，就是定义函数那一刻的环境
	}
```

执行 `let f = fn(x){x+2};` 时：
1. 走到`FunctionLiteral`分支；
2. 构造`*object.Function`返回；
3. 通过`env.Set`绑定到变量`f`；
4. **函数体完全没有执行**。

## 五、单元测试 TestFunctionObject
测试输入：`fn(x) { x + 2 }`

校验点：
1. Eval返回对象类型是 `object.FUNCTION_OBJ`；
2. 参数切片长度、参数名字和源码一致；
3. Body的AST内容和源码匹配。

> 本测试**不测试函数调用**，只验证函数对象构造正确。此时调用 `f(5)` 还无法工作。

## 六、闭包原理铺垫（只理论，尚未可调用）
示例 Monkey 代码
```monkey
let a = 10;
let f = fn() { return a; };
```
1. 定义`f`的时候，`&object.Function{Env: 当前全局env}`；
2. `f.Env`里面已经包含变量`a=10`；
3. 未来无论在哪调用f，都可以顺着这个Env链表找到`a`。

> 关键点：Env捕获发生在**定义时刻，不是调用时刻**。

## 七、当前完成清单
✅ 新增`object.Function`运行时对象，携带Parameters、Body、Env
✅ Eval处理`*ast.FunctionLiteral`，构造函数对象返回，不执行函数体
✅ 单元测试验证函数字面量求值

❌ 待实现 eval‑07：
- `*ast.CallExpression` 函数调用
- `evalExpressions` 批量求值实参
- `applyFunction`，`extendFunctionEnv` 创建调用子环境
- 形参实参绑定、执行函数体、unwrapReturnValue处理return信号

## 八、思考题
1. 为什么不把`Env`放在AST的`FunctionLiteral`上？
> AST是parser产出，parser没有环境概念；环境是运行时求值阶段才有的东西，只能放在运行时object。

2. 如果`Function.Env`存的是调用时环境，会发生什么？闭包直接失效。

---

# eval‑07‑函数调用 CallExpression
> 承接 eval‑06｜面向新手教学；实现 Monkey `f(arg1, arg2)`整套调用逻辑

## 一、本节目标
Monkey示例：`identity(5)`、`(fn(x){x * 2})(10)`
AST节点：`*ast.CallExpression`

完整流程：
1. 求值被调用函数本身；
2. 批量求值所有实参表达式；
3. 校验对象确实是`*object.Function`；
4. **基于函数捕获的Env创建本次调用私有子环境**；
5. 形参实参一一绑定写入子环境；
6. 在子环境内Eval执行函数体BlockStatement；
7. 解包`ReturnValue`信号，把结果返回给调用方。

> 注意：函数体内的`BlockStatement`不会解包ReturnValue，只透传信号壳；解包逻辑放在`applyFunction`。

## 二、单元测试 TestFunctionApplication
```go
tests := []struct {
	input    string
	expected int64
}{
	{"let identity = fn(x) { x }; identity(5);", 5},
	{"let identity = fn(x) { return x; }; identity(5);", 5},
	{"let double = fn(x) { x * 2 }; double(5);", 10},
	{"let add = fn(x,y) { x + y }; add(5,5);", 10},
	{"let add = fn(x,y) { x + y }; add(5+5, add(5,5));", 20},
	{"(fn(x){x})(5)", 5},
}
```
> 在未实现调用逻辑前全部失败，属于正常。

## 三、ast.CallExpression AST结构
```go
type CallExpression struct {
	Token     token.Token
	Function  ast.Expression // 可以是标识符，也可以直接是FunctionLiteral
	Arguments []ast.Expression // 实参AST切片，还不是运行时值，需要Eval
}
```

- `Function`：`add` 或者 `(fn(x){x})`；
- `Arguments`：保存的是AST表达式，例如`5+5`，必须逐个Eval。

## 四、Eval主switch增加CallExpression骨架
```go
case *ast.CallExpression:
	function := Eval(node.Function, env)
	if IsError(function) {
		return function
	}
	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && IsError(args[0]) {
		return args[0]
	}
	return applyFunction(function, args)
```

## 五、evalExpressions：批量求值实参
> 实参是AST表达式，需要全部转为运行时object；遇到错误立刻停止。

```go
func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if IsError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}
```

> 设计细节：出错返回`[]object.Object{evaluated}`，把error对象塞切片；上层判断`len(args)==1 && IsError(args[0])`取出错误向上返回。

## 六、applyFunction 入口、extendFunctionEnv、unwrapReturnValue
### 6.1 applyFunction
```go
func applyFunction(fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	extendedEnv := extendFunctionEnv(function, args)
	evaluated := Eval(function.Body, extendedEnv)
	return unwrapReturnValue(evaluated)
}
```

### 6.2 extendFunctionEnv：构造本次调用私有子环境
```go
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	// ⭐outer指向函数【定义时捕获的Env】，不是调用方env，实现闭包
	env := object.NewEnclosedEnvironment(fn.Env)
	// 形参实参绑定到本次调用子环境
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}
```
> 重点：`NewEnclosedEnvironment(fn.Env)`，outer来自`function.Env`，不是传入的调用方env。这是闭包能够工作的关键。

### 6.3 unwrapReturnValue 解包return信号
```go
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value // 函数调用这里做解包
	}
	return obj
}
```

> 区分三处解包位置
1. `evalProgram`：顶层Program解包return；
2. `applyFunction`：函数调用完成解包return；
3. `evalBlockStatement`：**绝对不解包，原样透传ReturnValue外壳**。

## 七、执行链路完整走读示例 `let f=fn(x){x+2}; f(5);`
1. 解析`fn(x){x+2}` → `*ast.FunctionLiteral`；Eval得到`*object.Function{Env:全局env}`；let存入全局env变量f；
2. 遇到`f(5)` CallExpression：
    1. Eval `f`（标识符），从全局env取出`*object.Function`；
    2. evalExpressions求值实参`5`得到`integer(5)`；
    3. extendFunctionEnv：以`function.Env`为outer新建子环境，绑定`x=5`；
    4. 在子环境Eval函数体BlockStatement；
    5. 函数体执行`x+2`得到7；没有return语句，返回普通integer；
    6. unwrapReturnValue识别不是ReturnValue，直接返回7。

## 八、闭包执行走读
```monkey
let a = 10;
let f = fn(){ return a; };
f();
```
1. 定义f，`f.Env`指向全局环境（里面有a=10）；
2. 调用f()，`NewEnclosedEnvironment(f.Env)`创建子环境；
3. 函数体内读取标识符`a`：子环境store找不到，顺着outer向外查找，找到全局a=10。

## 九、错误场景
1. 调用非函数：`5()` → 返回`error: not a function: INTEGER`；
2. 实参表达式求值报错：`f(true+5)`，实参求值报错，直接透传错误；
3. 参数数量不匹配：原版书实现**没有校验参数个数**，属于原版简化，是可扩展点。

## 十、当前完成清单
✅ `CallExpression`求值入口；
✅ `evalExpressions`批量求实参；
✅ `applyFunction`函数调用主逻辑；
✅ `extendFunctionEnv`基于函数捕获环境创建调用子环境，形参实参绑定；
✅ `unwrapReturnValue`解包函数内部return信号；
✅ 闭包可以正常工作；
✅ 单元测试验证普通函数调用、匿名函数直接调用。

❌ 待扩展：
- 内置函数（builtin）；
- 数组、hash字面量；

## 十一、思考题
1. 为什么`extendFunctionEnv`的outer是`fn.Env`而不是调用处传入的env？
> 如果使用调用方env，闭包捕获定义时变量的语义直接失效。

2. `evalBlockStatement`为什么不解包ReturnValue，而applyFunction要解包？
> BlockStatement可能多层嵌套，解包会丢失return终止信号；函数调用结束，return信号使命完成，允许解包交给上层。

3. 同一个Function对象多次调用，会不会互相干扰？
> 每次调用都会`NewEnclosedEnvironment`生成全新子环境，每次调用的局部变量互相隔离。`Function.Env`只是只读的外层引用，不会被修改。
