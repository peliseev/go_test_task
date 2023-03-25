package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type queue struct {
	mu    *sync.Mutex
	items chan []string
	empty chan struct{}
}

func newQueue() *queue {
	mu := &sync.Mutex{}
	items := make(chan []string, 1)
	empty := make(chan struct{}, 1)
	empty <- struct{}{}
	return &queue{
		mu:    mu,
		items: items,
		empty: empty,
	}
}

func (q *queue) get(ctx context.Context) (string, error) {
	var items []string
	select {
	case items = <-q.items:
	case <-ctx.Done():
		return "", ctx.Err()
	}

	v := items[0]
	items = items[1:]

	if len(items) == 0 {
		q.empty <- struct{}{}
	} else {
		q.items <- items
	}

	return v, nil
}

func (q *queue) put(value string) {
	select {
	case items := <-q.items:
		items = append(items, value)
		q.items <- items
	case <-q.empty:
		q.items <- []string{value}
	}
}

type queueManager struct {
	mu     *sync.Mutex
	queues map[string]*queue
}

func newQueueManager() *queueManager {
	return &queueManager{
		mu:     &sync.Mutex{},
		queues: make(map[string]*queue),
	}
}

func (m *queueManager) getQueue(name string) *queue {
	m.mu.Lock()
	q, ok := m.queues[name]
	if !ok {
		q = newQueue()
		m.queues[name] = q
	}
	m.mu.Unlock()
	return q
}

var (
	port    = flag.Int("p", 8080, "server port")
	manager = newQueueManager()
)

func main() {
	flag.Parse()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), handler()))
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			put(w, r)
		case http.MethodGet:
			get(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Path
	timeoutStr := r.URL.Query().Get("timeout")
	timeout, _ := strconv.Atoi(timeoutStr)

	q := manager.getQueue(queueName)

	ctx := r.Context()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	msg, err := q.get(ctx)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func put(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Path
	msg := r.URL.Query().Get("v")
	if msg == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q := manager.getQueue(queueName)
	q.put(msg)
}
