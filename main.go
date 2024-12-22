package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type RequestData struct {
	Expression string `json:"expression"`
}

type ResponseSuccess struct {
	Result string `json:"result"`
}

type ResponseError struct {
	Error string `json:"error"`
}

func main() {
	http.HandleFunc("/api/v1/calculate", calculateHandler)
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RequestData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if strings.TrimSpace(req.Expression) == "" {
		sendError(w, "Expression is not valid", http.StatusUnprocessableEntity)
		return
	}
	validRegex := regexp.MustCompile(`^[0-9+\-*/()\s]+$`)
	if !validRegex.MatchString(req.Expression) {
		sendError(w, "Expression is not valid", http.StatusUnprocessableEntity)
		return
	}
	res, err := evaluate(req.Expression)
	if err != nil {
		sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, fmt.Sprintf("%v", res))
}

func evaluate(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	if len(expr) > 1000 {
		return 0, fmt.Errorf("expression too long")
	}
	val, err := simpleEvaluate(expr)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func simpleEvaluate(expr string) (float64, error) {
	return evaluateManually(expr)
}

func evaluateManually(expr string) (float64, error) {
	stack := []string{}
	postfix, err := infixToPostfix(expr)
	if err != nil {
		return 0, err
	}
	for _, token := range postfix {
		switch token {
		case "+", "-", "*", "/":
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			bStr := stack[len(stack)-1]
			aStr := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			a, errA := strconv.ParseFloat(aStr, 64)
			b, errB := strconv.ParseFloat(bStr, 64)
			if errA != nil || errB != nil {
				return 0, fmt.Errorf("invalid number")
			}
			var res float64
			switch token {
			case "+":
				res = a + b
			case "-":
				res = a - b
			case "*":
				res = a * b
			case "/":
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				res = a / b
			}
			stack = append(stack, fmt.Sprintf("%v", res))
		default:
			stack = append(stack, token)
		}
	}
	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression final")
	}
	return strconv.ParseFloat(stack[0], 64)
}

func infixToPostfix(expr string) ([]string, error) {
	var output []string
	var stack []rune
	precedence := func(op rune) int {
		switch op {
		case '+', '-':
			return 1
		case '*', '/':
			return 2
		}
		return 0
	}
	for i := 0; i < len(expr); i++ {
		ch := rune(expr[i])
		if isDigit(ch) {
			num := strings.Builder{}
			num.WriteRune(ch)
			for i+1 < len(expr) && (isDigit(rune(expr[i+1])) || rune(expr[i+1]) == '.') {
				i++
				num.WriteRune(rune(expr[i]))
			}
			output = append(output, num.String())
		} else if ch == '(' {
			stack = append(stack, ch)
		} else if ch == ')' {
			var found bool
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top == '(' {
					found = true
					break
				}
				output = append(output, string(top))
			}
			if !found {
				return nil, fmt.Errorf("mismatched parentheses")
			}
		} else if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top == '(' || precedence(ch) > precedence(top) {
					break
				}
				stack = stack[:len(stack)-1]
				output = append(output, string(top))
			}
			stack = append(stack, ch)
		} else {
			return nil, fmt.Errorf("invalid character: %v", string(ch))
		}
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top == '(' || top == ')' {
			return nil, fmt.Errorf("mismatched parentheses in the end")
		}
		output = append(output, string(top))
	}
	return output, nil
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || ch == '.'
}

func sendSuccess(w http.ResponseWriter, result string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseSuccess{Result: result})
}

func sendError(w http.ResponseWriter, errMsg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ResponseError{Error: errMsg})
}
