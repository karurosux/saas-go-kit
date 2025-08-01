package job

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// Priority queue implementation for jobs
type jobItem struct {
	job      Job
	priority int
	index    int
}

type priorityQueue []*jobItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Higher priority first
	if pq[i].priority != pq[j].priority {
		return pq[i].priority > pq[j].priority
	}
	// Earlier scheduled jobs first for same priority
	return pq[i].job.GetCreatedAt().Before(pq[j].job.GetCreatedAt())
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*jobItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// InMemoryQueue implements JobQueue interface using a priority queue
type InMemoryQueue struct {
	pq    priorityQueue
	mu    sync.Mutex
	cond  *sync.Cond
}

func NewInMemoryQueue(ctx context.Context) JobQueue {
	q := &InMemoryQueue{
		pq:  make(priorityQueue, 0),
	}
	q.cond = sync.NewCond(&q.mu)
	heap.Init(&q.pq)
	return q
}

func (q *InMemoryQueue) Enqueue(ctx context.Context, job Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := &jobItem{
		job:      job,
		priority: int(job.GetPriority()),
	}
	heap.Push(&q.pq, item)
	q.cond.Signal()
	return nil
}

func (q *InMemoryQueue) Dequeue(ctx context.Context) (Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check for immediate availability
	if q.pq.Len() > 0 {
		item := heap.Pop(&q.pq).(*jobItem)
		return item.job, nil
	}

	// If queue is empty, wait a bit and return empty error
	// This avoids the complex cond.Wait() logic that was causing the mutex issue
	q.mu.Unlock()
	select {
	case <-ctx.Done():
		q.mu.Lock()
		return nil, ctx.Err()
	case <-time.After(100 * time.Millisecond):
		q.mu.Lock()
		if q.pq.Len() > 0 {
			item := heap.Pop(&q.pq).(*jobItem)
			return item.job, nil
		}
		return nil, ErrQueueEmpty
	}
}

func (q *InMemoryQueue) Requeue(ctx context.Context, job Job) error {
	return q.Enqueue(ctx, job)
}

func (q *InMemoryQueue) Size(ctx context.Context) (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pq.Len(), nil
}

func (q *InMemoryQueue) Clear(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pq = make(priorityQueue, 0)
	heap.Init(&q.pq)
	return nil
}