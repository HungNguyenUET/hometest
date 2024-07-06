package util

import "fmt"

type Queue struct {
	items []string
}

func (q *Queue) Enqueue(item string) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() string {
	if len(q.items) == 0 {
		fmt.Println("Queue is empty")
		return ""
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *Queue) Front() string {
	if len(q.items) == 0 {
		fmt.Println("Queue is empty")
		return ""
	}
	return q.items[0]
}

func (q *Queue) Size() int {
	return len(q.items)
}
