package calculator

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// 改为计算器

var specialOps = []string{
	"sin", "cos", "tan", "exp", "sqrt",
}

func HandleAndCalculate(exp string) string {
	if len(exp) > 100 {
		return "表达式过长，不能超过100字符"
	}
	defer func() {
		if err := recover(); err != nil {

		}
	}()

	exp = strings.ReplaceAll(exp, " ", "")
	exp = exp[:len(exp)-1]
	for _, c := range exp {
		if c >= 'A' && c <= 'E' {
			return "表达式输入有误或不符合规范"
		}
	}

	for i, op := range specialOps {
		exp = strings.ReplaceAll(exp, op, string('A'+i))
	}

	compile, _ := regexp.Compile("^[A-E0-9.+\\-*/^()]+$")
	if !compile.MatchString(exp) {
		return "表达式输入有误或不符合规范"
	}

	return calculate("(" + exp + ")")
}

func calculate(ex string) string {

	opStack := NewStack(50)
	numStack := NewStack(50)
	isNegative := false
	for i := 0; i < len(ex); {
		if ex[i] == '-' && (numStack.IsEmpty() || ex[i-1] == '(') {
			isNegative = true
			i++
		}
		if ex[i] >= '0' && ex[i] <= '9' {
			numBuff := ""
			for i < len(ex) && ex[i] >= '0' && ex[i] <= '9' || ex[i] == '.' {
				numBuff += string(ex[i])
				i++
			}
			num, err := strconv.ParseFloat(numBuff, 64)
			if err != nil {
				return "表达式输入有误或不符合规范"
			}
			if isNegative {
				num = num * (-1)
				isNegative = false
			}
			numStack.Push(num)
		}
		if i >= len(ex) {
			break
		}
		if ex[i] == ')' || ex[i] == '+' || ex[i] == '-' || ex[i] == '*' || ex[i] == '/' || ex[i] == '^' {
			switch ex[i] {
			case '+':
				fallthrough
			case '-':
				fallthrough
			case ')':
				for flag := true; flag; {
					flag = calcAddAndSub(&opStack, &numStack, ex[i])
				}

			case '*':
				fallthrough
			case '/':
				for flag := true; flag; {
					flag = calcMulAndDiv(&opStack, &numStack)

				}
			case '^':
			}
			if ex[i] != ')' {
				opStack.Push(ex[i])
			}
			i++
		}
		if i >= len(ex) {
			break
		}

		switch ex[i] {
		case 'A':
			fallthrough
		case 'B':
			fallthrough
		case 'C':
			fallthrough
		case 'D':
			fallthrough
		case 'E':
			fallthrough
		case '(':
			opStack.Push(ex[i])
			i++
		}

	}

	return fmt.Sprintf("%v", numStack.Pop().(float64))
}

func calcAddAndSub(opStack *Stack, numStack *Stack, curOp byte) bool {
	if opStack.IsNotEmpty() {
		op := opStack.Pop().(byte)
		switch op {
		case '+':
			r := numStack.Pop().(float64)
			l := numStack.Pop().(float64)
			numStack.Push(l + r)
		case '-':
			r := numStack.Pop().(float64)
			l := numStack.Pop().(float64)
			numStack.Push(l - r)
		case '*':
			r := numStack.Pop().(float64)
			l := numStack.Pop().(float64)
			numStack.Push(l * r)
		case '/':
			r := numStack.Pop().(float64)
			l := numStack.Pop().(float64)
			numStack.Push(l / r)
		case '^':
			r := numStack.Pop().(float64)
			l := numStack.Pop().(float64)
			numStack.Push(math.Pow(l, r))
		case 'A':
			l := numStack.Pop().(float64)
			numStack.Push(math.Sin(l))
		case 'B':
			l := numStack.Pop().(float64)
			numStack.Push(math.Cos(l))
		case 'C':
			l := numStack.Pop().(float64)
			numStack.Push(math.Tan(l))
		case 'D':
			l := numStack.Pop().(float64)
			numStack.Push(math.Exp(l))
		case 'E':
			l := numStack.Pop().(float64)
			numStack.Push(math.Sqrt(l))
		case '(':
			if curOp == ')' {
				return false
			} else {
				opStack.Push(op)
				return false
			}
		}
		return true
	}
	return false
}
func calcMulAndDiv(opStack *Stack, numStack *Stack) bool {
	if opStack.IsNotEmpty() {
		op := opStack.Pop().(byte)
		switch op {
		case '+':
			fallthrough
		case '-':
			fallthrough
		case '(':
			opStack.Push(op)
			return false
		case '*':
			r := numStack.Pop().(float64)
			l := numStack.Pop().(float64)
			numStack.Push(l * r)
		case '/':
			r := numStack.Pop().(float64)
			l := numStack.Pop().(float64)
			numStack.Push(l / r)
		case '^':
			r := numStack.Pop().(float64)
			l := numStack.Pop().(float64)
			numStack.Push(math.Pow(l, r))
		case 'A':
			l := numStack.Pop().(float64)
			numStack.Push(math.Sin(l))
		case 'B':
			l := numStack.Pop().(float64)
			numStack.Push(math.Cos(l))
		case 'C':
			l := numStack.Pop().(float64)
			numStack.Push(math.Tan(l))
		case 'D':
			l := numStack.Pop().(float64)
			numStack.Push(math.Exp(l))
		case 'E':
			l := numStack.Pop().(float64)
			numStack.Push(math.Sqrt(l))
		}
		return true
	}
	return false
}
