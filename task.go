package task

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type task struct {
	id     int64
	name   string
	cancel func()
	f      TaskFn
}

type Manager struct {
	mu    sync.RWMutex
	ID    int64
	F     TaskFn
	tasks map[int64]*task
}

func NewManager(opts ...func(*Manager) error) (*Manager, error) {
	rc := Manager{
		tasks: map[int64]*task{},
	}
	for _, opt := range opts {
		if err := opt(&rc); err != nil {
			return nil, err
		}
	}
	return &rc, nil
}

type TaskFn func(context.Context) error

func (m *Manager) newTask(ctx context.Context, name string, f TaskFn) (*task, context.Context, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	id := atomic.AddInt64(&m.ID, 1)

	ctx, cancel := context.WithCancel(ctx)

	t := task{
		id:     id,
		name:   name,
		cancel: cancel,
		f:      f,
	}

	return &t, ctx, nil
}

type Task struct {
	ID   int64
	Name string
}

func (m *Manager) Kill(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tasks[id]
	if !ok {
		return fmt.Errorf("no such task %d", id)
	}

	t.cancel()

	return nil
}

func (m *Manager) PS() []Task {
	m.mu.Lock()
	defer m.mu.Unlock()

	rc := []Task{}
	for _, t := range m.tasks {
		rc = append(rc, Task{ID: t.id, Name: t.name})
	}

	return rc
}

func (m *Manager) Run(ctx context.Context, name string, fn TaskFn) error {
	task, ctx, err := m.newTask(ctx, name, fn)

	if err != nil {
		return err
	}

	m.mu.Lock()
	m.tasks[task.id] = task
	m.mu.Unlock()

	return func() error {
		defer func() {
			m.mu.Lock()
			delete(m.tasks, task.id)
			m.mu.Unlock()
		}()

		return task.f(ctx)
	}()
}
