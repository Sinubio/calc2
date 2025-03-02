package calculate

import (
    "sync"
)

func collectTasks(root *Task) []*Task {
    tasks := []*Task{}
    var visit func(t *Task)
    visit = func(t *Task) {
        tasks = append(tasks, t)
        if t.Operation != "num" {
            if arg1, ok := t.Arg1.(*Task); ok {
                visit(arg1)
            }
            if arg2, ok := t.Arg2.(*Task); ok {
                visit(arg2)
            }
        }
    }
    visit(root)
    return tasks
}

func isTaskReady(task *Task) bool {
    task.mu.Lock()
    defer task.mu.Unlock()

    if task.Operation == "num" {
        return true
    }

    arg1Ready := true
    if t, ok := task.Arg1.(*Task); ok {
        arg1Ready = t.Status == "completed"
    }

    arg2Ready := true
    if t, ok := task.Arg2.(*Task); ok {
        arg2Ready = t.Status == "completed"
    }

    return arg1Ready && arg2Ready
}