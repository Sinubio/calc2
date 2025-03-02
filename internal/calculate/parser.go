package calculate

import (
    "errors"
    "strconv"
    "strings"
)

type Task struct {
    ID         string
    Operation  string
    Arg1       interface{}
    Arg2       interface{}
    Result     float64
    Status     string
    DependsOn  []*Task
    DependedBy []*Task
}

func ParseExpression(expr string) (*Task, error) {
    expr = strings.ReplaceAll(expr, " ", "")
    var stack []*Task
    var opStack []rune
    num := ""

    for _, ch := range expr {
        switch ch {
        case '+', '-', '*', '/':
            if num != "" {
                val, _ := strconv.ParseFloat(num, 64)
                stack = append(stack, &Task{Operation: "num", Result: val, Status: "completed"})
                num = ""
            }
            for len(opStack) > 0 && precedence(ch) <= precedence(opStack[len(opStack)-1]) {
                processOperation(&stack, &opStack)
            }
            opStack = append(opStack, ch)
        default:
            if (ch >= '0' && ch <= '9') || ch == '.' {
                num += string(ch)
            } else {
                return nil, errors.New("unsupported character")
            }
        }
    }

    if num != "" {
        val, _ := strconv.ParseFloat(num, 64)
        stack = append(stack, &Task{Operation: "num", Result: val, Status: "completed"})
    }

    for len(opStack) > 0 {
        processOperation(&stack, &opStack)
    }

    if len(stack) != 1 {
        return nil, errors.New("invalid expression")
    }

    return stack[0], nil
}

func processOperation(stack *[]*Task, opStack *[]rune) {
    op := (*opStack)[len(*opStack)-1]
    *opStack = (*opStack)[:len(*opStack)-1]
    arg2 := (*stack)[len(*stack)-1]
    arg1 := (*stack)[len(*stack)-2]
    *stack = (*stack)[:len(*stack)-2]
    newTask := &Task{
        Operation: string(op),
        Arg1:      arg1,
        Arg2:      arg2,
        Status:    "pending",
    }
    newTask.DependsOn = getTaskDependencies(newTask)
    for _, dep := range newTask.DependsOn {
        dep.DependedBy = append(dep.DependedBy, newTask)
    }
    *stack = append(*stack, newTask)
}

func getTaskDependencies(task *Task) []*Task {
    var deps []*Task
    if t, ok := task.Arg1.(*Task); ok {
        deps = append(deps, t)
    }
    if t, ok := task.Arg2.(*Task); ok {
        deps = append(deps, t)
    }
    return deps
}

func precedence(op rune) int {
    switch op {
    case '+', '-':
        return 1
    case '*', '/':
        return 2
    default:
        return 0
    }
}