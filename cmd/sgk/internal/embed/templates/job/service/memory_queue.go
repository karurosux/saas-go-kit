package jobservice

import (
	"container/heap"
	"context"
	"errors"
	"sync"
	"time"
	
	"{{.Project.GoModule}}/internal/job/interface"
)

// priorityQueue implements heap.Interface for job priority queue
type priorityQueue []jobinterface.Job

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Higher priority first
	if pq[i].GetPriority() != pq[j].GetPriority() {
		return pq[i].GetPriority() > pq[j].GetPriority()
	}
	// Earlier scheduled time first
	return pq[i].GetScheduledAt().Before(pq[j].GetScheduledAt())
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	job := x.(jobinterface.Job)
	*pq = append(*pq, job)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	job := old[n-1]
	*pq = old[0 : n-1]
	return job
}

// MemoryQueue implements an in-memory job queue
type MemoryQueue struct {
	queue     priorityQueue
	mutex     sync.Mutex
	maxSize   int
	waitGroup sync.WaitGroup
	closed    bool
}

// NewMemoryQueue creates a new in-memory queue
func NewMemoryQueue(maxSize int) jobinterface.JobQueue {
	if maxSize <= 0 {
		maxSize = 10000
	}
	
	q := &MemoryQueue{
		queue:   make(priorityQueue, 0),
		maxSize: maxSize,
	}
	
	heap.Init(&q.queue)
	return q
}

// Enqueue adds a job to the queue
func (q *MemoryQueue) Enqueue(ctx context.Context, job jobinterface.Job) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if q.closed {
		return errors.New("queue is closed")
	}
	
	if len(q.queue) >= q.maxSize {
		return errors.New("queue is full")
	}
	
	heap.Push(&q.queue, job)
	return nil
}

// Dequeue retrieves the next job from the queue
func (q *MemoryQueue) Dequeue(ctx context.Context) (jobinterface.Job, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if q.closed {
		return nil, errors.New("queue is closed")
	}
	
	for len(q.queue) > 0 {
		job := heap.Pop(&q.queue).(jobinterface.Job)
		
		// Check if job is scheduled for now or past
		if job.GetScheduledAt().Before(time.Now()) {
			return job, nil
		}
		
		// Put it back if scheduled for future
		heap.Push(&q.queue, job)
		break
	}
	
	return nil, errors.New("no jobs available")
}

// DequeueByType retrieves the next job of a specific type
func (q *MemoryQueue) DequeueByType(ctx context.Context, jobType string) (jobinterface.Job, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if q.closed {
		return nil, errors.New("queue is closed")
	}
	
	// Find job of specific type
	for i, job := range q.queue {
		if job.GetType() == jobType && job.GetScheduledAt().Before(time.Now()) {
			// Remove from queue
			q.queue[i] = q.queue[len(q.queue)-1]
			q.queue = q.queue[:len(q.queue)-1]
			heap.Fix(&q.queue, i)
			return job, nil
		}
	}
	
	return nil, errors.New("no jobs of specified type available")
}

// Peek looks at the next job without removing it
func (q *MemoryQueue) Peek(ctx context.Context) (jobinterface.Job, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if q.closed {
		return nil, errors.New("queue is closed")
	}
	
	if len(q.queue) == 0 {
		return nil, errors.New("queue is empty")
	}
	
	return q.queue[0], nil
}

// Size returns the number of jobs in the queue
func (q *MemoryQueue) Size(ctx context.Context) (int64, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	return int64(len(q.queue)), nil
}

// Clear removes all jobs from the queue
func (q *MemoryQueue) Clear(ctx context.Context) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if q.closed {
		return errors.New("queue is closed")
	}
	
	q.queue = make(priorityQueue, 0)
	heap.Init(&q.queue)
	return nil
}

// Close closes the queue
func (q *MemoryQueue) Close() error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if q.closed {
		return errors.New("queue already closed")
	}
	
	q.closed = true
	return nil
}