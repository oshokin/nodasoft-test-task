package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// worker представляет собой описание обработчика задач
type worker struct {
	wg              *sync.WaitGroup
	queue           chan *task
	successfulTasks []*task
	failedTasks     []*task
	startedAt       time.Time
	elapsedTime     time.Duration
}

const (
	defaultQueueSize   = 10
	maxExecutionTime   = -20 * time.Second
	processingTimeout  = 150 * time.Millisecond
	maxTasksShownCount = 100
)

func newWorker(queueSize int) *worker {
	if queueSize == 0 {
		queueSize = defaultQueueSize
	}

	return &worker{
		wg:              &sync.WaitGroup{},
		queue:           make(chan *task, queueSize),
		successfulTasks: make([]*task, 0, queueSize),
		failedTasks:     make([]*task, 0, queueSize),
	}
}

func (w *worker) startCreatingTasks(ctx context.Context) {
	w.startedAt = time.Now()
	w.wg.Add(1)

	go func(ctx context.Context, w *worker) {
		defer w.wg.Done()

		for {
			var (
				now = time.Now()
				err error
			)

			// formattedTime := now.Format(time.RFC3339)
			// вот такое условие появления ошибочных задач
			if now.Nanosecond()%2 > 0 {
				err = errors.New("some error occured")
			}

			select {
			case <-ctx.Done():
				close(w.queue)
				return
			default:
				// передаем задачу на выполнение
				w.queue <- &task{
					id:        now.Unix(),
					createdAt: now,
					err:       err,
				}
			}
		}
	}(ctx, w)
}

func (w *worker) startProcessingTasks(ctx context.Context) {
	w.wg.Add(1)

	go func(ctx context.Context, w *worker) {
		defer w.wg.Done()

		for t := range w.queue {
			curTask := t
			w.wg.Add(1)

			go func(t *task) {
				defer w.wg.Done()

				w.processTask(t)
				w.checkForErrors(t)
				time.Sleep(processingTimeout)
			}(curTask)
		}
	}(ctx, w)
}

func (w *worker) processTask(t *task) {
	now := time.Now()

	if t.err == nil {
		if t.createdAt.After(now.Add(maxExecutionTime)) {
			t.err = nil
		} else {
			t.err = errors.New("something went wrong")
		}
	}

	t.finishedAt = now
}

func (w *worker) checkForErrors(t *task) {
	if t.err != nil {
		w.failedTasks = append(w.failedTasks, t)
		return
	}

	w.successfulTasks = append(w.successfulTasks, t)
}

func (w *worker) waitForTaskToFinish() {
	w.wg.Wait()

	w.elapsedTime = time.Since(w.startedAt)
}

func (w *worker) printResults() {
	fmt.Printf("elapsed time: %s\n", w.elapsedTime)

	w.prettyPrintTasks(w.failedTasks, "tasks finished with error")
	w.prettyPrintTasks(w.successfulTasks, "successfully finished tasks")
}

func (w *worker) prettyPrintTasks(tasks []*task, title string) {
	switch {
	case len(tasks) > maxTasksShownCount:
		fmt.Printf("count of %s: %d\n", title, len(tasks))
	case len(tasks) > 0 && len(tasks) <= maxTasksShownCount:
		fmt.Printf("%s:\n", title)
		for _, t := range tasks {
			fmt.Println(t)
		}
	}
}
