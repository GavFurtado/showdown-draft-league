package utils

import (
	"github.com/google/uuid"
	"time"
)

type TaskType int // not a string like my other enums cuz it's not going into the db
const (
	TurnTypeDraftTurnTimeout TaskType = 0
	TurnTypeTradingPeriodEnd TaskType = 1
	TurnTypeAccrueCredits    TaskType = 2
)

type ScheduledTask struct {
	ID        uuid.UUID
	ExecuteAt time.Time
	Type      TaskType
	Payload   any
	index     int // allows for efficient updates if we wanna go down that route
}

// payloads
type PayloadDraftTurnTimeout struct {
	DraftID  uuid.UUID
	LeagueID uuid.UUID
	PlayerID uuid.UUID // The player whose turn it is
}
type PayloadTransferPeriodEnd struct {
	LeagueID uuid.UUID
}
type PayloadTransferCreditAccrual struct {
	LeagueID uuid.UUID
}

type TaskHeap []*ScheduledTask

// container/heap package requires implementing the following methods to make the heap work

func (heap TaskHeap) Len() int {
	return len(heap)
}

func (heap TaskHeap) Less(i, j int) bool {
	return heap[i].ExecuteAt.Before(heap[j].ExecuteAt)
}

func (heap TaskHeap) Swap(i, j int) {
	heap[i], heap[j] = heap[j], heap[i]
	heap[i].index = i
	heap[j].index = j
}

func (heap *TaskHeap) Push(x any) {
	n := len(*heap)
	task := x.(*ScheduledTask)
	task.index = n
	*heap = append(*heap, task)
}

func (heap *TaskHeap) Pop() any {
	old := *heap
	n := len(old)
	task := old[n-1]
	task.index = -1   // mark as removed
	*heap = old[:n-1] // remove last element
	return task
}
