package server

import (
    "encoding/json"
    "errors"
    "net/http"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/Sinubio/mycalcservice/pkg/calculate"
)

type Expression struct {
    ID       string
    Status   string
    Result   float64
    Tasks    []*calculate.Task
    RootTask *calculate.Task
    mu       sync.Mutex
}

type Server struct {
    expressions sync.Map
    taskQueue   chan *calculate.Task
    addr        string
}

func NewServer(addr string) *Server {
    return &Server{
        addr:      addr,
        taskQueue: make(chan *calculate.Task, 100),
    }
}

func (s *Server) Run() error {
    http.HandleFunc("/api/v1/calculate", s.handleCalculate)
    http.HandleFunc("/api/v1/expressions", s.handleGetExpressions)
    http.HandleFunc("/api/v1/expressions/", s.handleGetExpression)
    http.HandleFunc("/internal/task", s.handleTask)

    go s.processTaskQueue()

    return http.ListenAndServe(s.addr, nil)
}

func (s *Server) handleCalculate(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Expression string `json:"expression"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    rootTask, err := calculate.ParseExpression(req.Expression)
    if err != nil {
        http.Error(w, "Invalid expression", http.StatusUnprocessableEntity)
        return
    }

    exprID := uuid.New().String()
    tasks := collectTasks(rootTask)
    expr := &Expression{
        ID:       exprID,
        Status:   "pending",
        Tasks:    tasks,
        RootTask: rootTask,
    }

    s.expressions.Store(exprID, expr)

    for _, task := range tasks {
        if isTaskReady(task) {
            s.taskQueue <- task
        }
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"id": exprID})
}

func collectTasks(root *calculate.Task) []*calculate.Task {
    tasks := []*calculate.Task{}
    var visit func(t *calculate.Task)
    visit = func(t *calculate.Task) {
        tasks = append(tasks, t)
        if t.Operation != "num" {
            if arg1, ok := t.Arg1.(*calculate.Task); ok {
                visit(arg1)
            }
            if arg2, ok := t.Arg2.(*calculate.Task); ok {
                visit(arg2)
            }
        }
    }
    visit(root)
    return tasks
}

func isTaskReady(task *calculate.Task) bool {
    if task.Operation == "num" {
        return true
    }
    arg1Ready := true
    if t, ok := task.Arg1.(*calculate.Task); ok {
        arg1Ready = t.Status == "completed"
    }
    arg2Ready := true
    if t, ok := task.Arg2.(*calculate.Task); ok {
        arg2Ready = t.Status == "completed"
    }
    return arg1Ready && arg2Ready
}

func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        select {
        case task := <-s.taskQueue:
            task.Status = "processing"
            resp := map[string]interface{}{
                "task": map[string]interface{}{
                    "id":         task.ID,
                    "arg1":       getArgValue(task.Arg1),
                    "arg2":       getArgValue(task.Arg2),
                    "operation":  task.Operation,
                    "operation_time": getOperationTime(task.Operation),
                },
            }
            json.NewEncoder(w).Encode(resp)
        default:
            http.Error(w, "No tasks available", http.StatusNotFound)
        }
    } else if r.Method == http.MethodPost {
        var req struct {
            ID     string  `json:"id"`
            Result float64 `json:"result"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        var foundTask *calculate.Task
        s.expressions.Range(func(key, value interface{}) bool {
            expr := value.(*Expression)
            for _, task := range expr.Tasks {
                if task.ID == req.ID {
                    foundTask = task
                    return false
                }
            }
            return true
        })

        if foundTask == nil {
            http.Error(w, "Task not found", http.StatusNotFound)
            return
        }

        foundTask.mu.Lock()
        foundTask.Result = req.Result
        foundTask.Status = "completed"
        foundTask.mu.Unlock()

        expr := foundTask.DependedBy[0].RootTask.Expression
        expr.mu.Lock()
        defer expr.mu.Unlock()
        for _, depTask := range foundTask.DependedBy {
            if isTaskReady(depTask) {
                s.taskQueue <- depTask
            }
        }

        if expr.RootTask.Status == "completed" {
            expr.Status = "completed"
            expr.Result = expr.RootTask.Result
        }
    }
}

func getArgValue(arg interface{}) float64 {
    switch v := arg.(type) {
    case *calculate.Task:
        return v.Result
    case float64:
        return v
    default:
        return 0
    }
}

func getOperationTime(op string) int {
    switch op {
    case "+":
        return getTimeEnv("TIME_ADDITION_MS")
    case "-":
        return getTimeEnv("TIME_SUBTRACTION_MS")
    case "*":
        return getTimeEnv("TIME_MULTIPLICATIONS_MS")
    case "/":
        return getTimeEnv("TIME_DIVISIONS_MS")
    default:
        return 1000
    }
}

func getTimeEnv(key string) int {
    value := os.Getenv(key)
    if value == "" {
        return 1000
    }
    i, _ := strconv.Atoi(value)
    return i
}