package spider

import (
	"errors"
	"sync"
)

type QueueItem struct {
	URL   string
	Depth int
}

type URLQueue struct {
	items []QueueItem
	mutex sync.Mutex
}

func NewURLQueue() *URLQueue {
	return &URLQueue{
		items: make([]QueueItem, 0),
	}
}

func (q *URLQueue) Push(url string, depth int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, QueueItem{URL: url, Depth: depth})
}

func (q *URLQueue) Pop() (QueueItem, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) == 0 {
		return QueueItem{}, errors.New("queue is empty")
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

func (q *URLQueue) IsEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items) == 0
}

func (q *URLQueue) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items)
}

func (q *URLQueue) List() []QueueItem {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.items
}

func (q *URLQueue) LastItem() *QueueItem {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.items) == 0 {
		return nil
	}

	return &q.items[len(q.items)-1]
}
