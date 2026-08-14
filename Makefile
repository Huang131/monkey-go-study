# Monkey Interpreter Makefile

.PHONY: help run lex parse eval test clean repl

# 默认目标：显示帮助
help:
	@echo "Monkey Interpreter - Makefile"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  run      运行交互式 REPL（默认）"
	@echo "  lex      运行词法分析调试"
	@echo "  parse    运行语法分析调试"
	@echo "  eval     运行求值调试（完整流程）"
	@echo "  test     运行所有测试"
	@echo "  bench    运行性能测试"
	@echo "  clean    清理缓存文件"
	@echo ""

# 交互式 REPL
repl run:
	go run .

# 词法分析调试：显示 Token 流
lex:
	go run . --lex

# 语法分析调试：显示 AST 结构
parse:
	go run . --parse

# 求值调试：完整流程 Lexer → Parser → Evaluator
# 注意：GoLand 中运行 make eval 可能无法点击，改用 make run-eval
run-eval:
	go run . --eval

# 运行所有测试
test:
	go test -v ./...

# 运行测试（简洁输出）
test-quiet:
	go test ./...

# 性能测试
bench:
	go test -bench=. -benchmem ./...

# 清理缓存文件
clean:
	go clean
	find . -name "*.test" -delete 2>/dev/null || true