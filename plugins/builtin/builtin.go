package builtin

import (
	"fmt"
	"math"
	"strings"
	"time"

	"gogent/plugins"
)

type BuiltinCalculator struct{}

func (b *BuiltinCalculator) Name() string { return "calculator" }
func (b *BuiltinCalculator) Description() string {
	return "Evaluate mathematical expressions. Use for calculations, math problems, numbers, percentages, etc."
}
func (b *BuiltinCalculator) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "The mathematical expression to evaluate (e.g. 2+3*4, (1+2)*5)",
			},
		},
		"required": []string{"expression"},
	}
}
func (b *BuiltinCalculator) Execute(args map[string]interface{}) (string, error) {
	expr, ok := args["expression"].(string)
	if !ok {
		return "", fmt.Errorf("expression must be a string")
	}
	expr = strings.ReplaceAll(expr, "^", "**")
	expr = strings.ReplaceAll(expr, "×", "*")
	expr = strings.ReplaceAll(expr, "÷", "/")
	result, err := evaluateMath(expr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Result: %s = %g", expr, result), nil
}

func evaluateMath(expr string) (float64, error) {
	parser := NewMathParser(expr)
	result, err := parser.ParseExpression()
	if err != nil {
		return 0, err
	}
	return result, nil
}

type MathParser struct {
	expr string
	pos  int
}

func NewMathParser(expr string) *MathParser {
	return &MathParser{expr: expr, pos: 0}
}

func (p *MathParser) ParseExpression() (float64, error) {
	result, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.expr) {
		op := p.expr[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			result += right
		} else {
			result -= right
		}
	}
	return result, nil
}

func (p *MathParser) parseTerm() (float64, error) {
	result, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.expr) {
		op := p.expr[p.pos]
		if op != '*' && op != '/' && op != '%' {
			break
		}
		p.pos++
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		switch op {
		case '*':
			result *= right
		case '/':
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			result /= right
		case '%':
			result = math.Mod(result, right)
		}
	}
	return result, nil
}

func (p *MathParser) parseFactor() (float64, error) {
	if p.pos >= len(p.expr) {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	if p.expr[p.pos] == '(' {
		p.pos++
		result, err := p.ParseExpression()
		if err != nil {
			return 0, err
		}
		if p.pos >= len(p.expr) || p.expr[p.pos] != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return result, nil
	}
	return p.parseNumber()
}

func (p *MathParser) parseNumber() (float64, error) {
	start := p.pos
	for p.pos < len(p.expr) && ((p.expr[p.pos] >= '0' && p.expr[p.pos] <= '9') || p.expr[p.pos] == '.') {
		p.pos++
	}
	if start == p.pos {
		return 0, fmt.Errorf("expected number at position %d", p.pos)
	}
	numStr := p.expr[start:p.pos]
	var num float64
	fmt.Sscanf(numStr, "%f", &num)
	return num, nil
}

type BuiltinTimeTool struct{}

func (b *BuiltinTimeTool) Name() string { return "get_current_time" }
func (b *BuiltinTimeTool) Description() string {
	return "Get the current date and time. Use when user asks about time, date, or current moment."
}
func (b *BuiltinTimeTool) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}
func (b *BuiltinTimeTool) Execute(args map[string]interface{}) (string, error) {
	return fmt.Sprintf("Current time: %s", time.Now().Format(time.RFC1123)), nil
}

type BuiltinPlugin struct{}

func NewBuiltinPlugin() *BuiltinPlugin {
	return &BuiltinPlugin{}
}

func (p *BuiltinPlugin) Name() string        { return "builtin" }
func (p *BuiltinPlugin) Version() string     { return "1.0.0" }
func (p *BuiltinPlugin) Description() string { return "Built-in tools for calculator and time" }

func (p *BuiltinPlugin) Initialize(config map[string]interface{}) error {
	return nil
}

func (p *BuiltinPlugin) GetTools() []plugins.Tool {
	return []plugins.Tool{
		&BuiltinCalculator{},
		&BuiltinTimeTool{},
	}
}

func init() {
	plugins.RegisterPlugin(NewBuiltinPlugin())
}

var _ plugins.Plugin = (*BuiltinPlugin)(nil)
