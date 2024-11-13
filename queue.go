package main

import "sync"

type Command struct {
	Type  string
	Args  []string
	Event Event
}

type CommandQueue struct {
	queue []Command
	lock  sync.Mutex
}

func (q *CommandQueue) Enqueue(cmd Command) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.queue = append(q.queue, cmd)
}

func (q *CommandQueue) Dequeue() *Command {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.queue) == 0 {
		return nil
	}
	cmd := q.queue[0]
	q.queue = q.queue[1:]
	return &cmd
}
