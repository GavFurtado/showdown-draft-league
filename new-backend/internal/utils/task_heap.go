package utils

import (
	"github.com/google/uuid"
	"time"
)

type TaskType int // not a string like my other enums cuz it's not going into the db
const (
	TaskTypeDraftTurnTimeout TaskType = iota
	TaskTypeTradingPeriodEnd
	TaskTypeTradingPeriodStart
)

func (t TaskType) String() string {
	switch t {
	case TaskTypeDraftTurnTimeout:
		return "DRAFT_TURN_TIMEOUT"
	case TaskTypeTradingPeriodEnd:
		return "TRADING_PERIOD_END"
	case TaskTypeTradingPeriodStart:
		return "TRADING_PERIOD_START"
	}
	return ""
}

type ScheduledTask struct {
	ID        string
	ExecuteAt time.Time
	Type      TaskType
	Payload   any
	Index     int
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
type PayloadTransferPeriodStart struct {
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
	heap[i].Index = i
	heap[j].Index = j
}

func (heap *TaskHeap) Push(x any) {
	n := len(*heap)
	task := x.(*ScheduledTask)
	task.Index = n
	*heap = append(*heap, task)
}

func (heap *TaskHeap) Peek() (*ScheduledTask, bool) {
	if len(*heap) == 0 {
		return nil, false
	}
	return (*heap)[0], true
}

func (heap *TaskHeap) Pop() any {
	old := *heap
	n := len(old)
	task := old[n-1]
	task.Index = -1   // mark as removed
	*heap = old[:n-1] // remove last element
	return task
}
