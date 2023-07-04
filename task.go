package main

import (
	"fmt"
	"time"
)

// task представляет собой описание задачи
type task struct {
	id         int64     // ID задачи
	createdAt  time.Time // время создания
	finishedAt time.Time // время выполнения
	err        error     // ошибка выполнения
}

func (t *task) String() string {
	if t.err != nil {
		return fmt.Sprintf("id: %d, created at: %s, finished at: %s, error: %v",
			t.id,
			t.createdAt.Format(time.RFC3339),
			t.finishedAt.Format(time.RFC3339),
			t.err)
	}

	return fmt.Sprintf("id: %d, created at: %s, finished at: %s",
		t.id,
		t.createdAt.Format(time.RFC3339),
		t.finishedAt.Format(time.RFC3339))
}
