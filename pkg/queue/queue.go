package queue

import (
	"github.com/unikiosk/unikiosk/pkg/models"
)

type Queue interface {
	Emit(event models.KioskState)
	Listen() <-chan models.KioskState
}

type queue struct {
	workQueue chan models.KioskState
}

func New(size int) (*queue, error) {
	return &queue{
		workQueue: make(chan models.KioskState, size),
	}, nil
}

func (q *queue) Emit(event models.KioskState) {
	q.workQueue <- event
}

func (q *queue) Listen() <-chan models.KioskState {
	return q.workQueue
}

func (q *queue) Close() {
	close(q.workQueue)
}
