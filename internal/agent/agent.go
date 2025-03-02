package agent

import (
    "encoding/json"
    "net/http"
    "os"
    "strconv"
    "sync"
    "time"
)

type Agent struct {
    serverURL string
    wg        sync.WaitGroup
}

func NewAgent(serverURL string) *Agent {
    return &Agent{serverURL: serverURL}
}

func (a *Agent) Start() {
    computePower := getComputePower()
    for i := 0; i < computePower; i++ {
        a.wg.Add(1)
        go a.worker()
    }
    a.wg.Wait()
}

func (a *Agent) worker() {
    defer a.wg.Done()
    for {
        task := a.getTask()
        if task == nil {
            time.Sleep(time.Second)
            continue
        }

        opTime := getOperationTime(task.Operation)
        time.Sleep(time.Duration(opTime) * time.Millisecond)

        result := calculateResult(task)
        a.postResult(task.ID, result)
    }
}

func (a *Agent) getTask() *Task {
    resp, err := http.Get(a.serverURL + "/internal/task")
    if err != nil || resp.StatusCode != http.StatusOK {
        return nil
    }
    defer resp.Body.Close()

    var data struct {
        Task *Task `json:"task"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil
    }
    return data.Task
}

func (a *Agent) postResult(taskID string, result float64) {
    data := map[string]interface{}{
        "id":     taskID,
        "result": result,
    }
    resp, err := http.Post(a.serverURL+"/internal/task", "application/json", jsonBody(data))
    if err != nil || resp.StatusCode != http.StatusOK {
        return
    }
    resp.Body.Close()
}

func calculateResult(task *Task) float64 {
    arg1 := getArgValue(task.Arg1)
    arg2 := getArgValue(task.Arg2)
    switch task.Operation {
    case "+":
        return arg1 + arg2
    case "-":
        return arg1 - arg2
    case "*":
        return arg1 * arg2
    case "/":
        return arg1 / arg2
    default:
        return 0
    }
}

func getArgValue(arg interface{}) float64 {
    switch v := arg.(type) {
    case float64:
        return v
    case *Task:
        return v.Result
    default:
        return 0
    }
}

func getOperationTime(op string) int {
    switch op {
    case "+":
        return getEnvInt("TIME_ADDITION_MS", 1000)
    case "-":
        return getEnvInt("TIME_SUBTRACTION_MS", 1000)
    case "*":
        return getEnvInt("TIME_MULTIPLICATIONS_MS", 1000)
    case "/":
        return getEnvInt("TIME_DIVISIONS_MS", 1000)
    default:
        return 1000
    }
}

func getComputePower() int {
    return getEnvInt("COMPUTING_POWER", 2)
}

func getEnvInt(key string, defaultValue int) int {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    i, _ := strconv.Atoi(value)
    return i
}

func jsonBody(data interface{}) *bytes.Buffer {
    buf := new(bytes.Buffer)
    json.NewEncoder(buf).Encode(data)
    return buf
}