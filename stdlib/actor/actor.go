package actor

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// Message encapsulates payload sent between isolated actors
type Message struct {
	SenderID string
	Type     string
	Payload  interface{}
}

// Mailbox is a high-throughput lock-free ring/channel mailbox
type Mailbox struct {
	queue chan Message
}

func NewMailbox(capacity int) *Mailbox {
	return &Mailbox{
		queue: make(chan Message, capacity),
	}
}

// HandlerFunc processes incoming messages inside an actor context
type HandlerFunc func(ctx context.Context, msg Message) error

// Actor represents an isolated zero-ambient computation unit
type Actor struct {
	ID        string
	mailbox   *Mailbox
	handler   HandlerFunc
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning int32
	mu        sync.RWMutex
}

// System manages actor lifecycles and supervision trees
type System struct {
	actors     map[string]*Actor
	deadLetter chan Message
	mu         sync.RWMutex
}

func NewSystem() *System {
	return &System{
		actors:     make(map[string]*Actor),
		deadLetter: make(chan Message, 1024),
	}
}

// DeadLetterQueue returns channel with unhandled / dropped messages
func (s *System) DeadLetterQueue() <-chan Message {
	return s.deadLetter
}

func (s *System) Spawn(id string, mailboxSize int, handler HandlerFunc) (*Actor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.actors[id]; exists {
		return nil, errors.New("actor already exists: " + id)
	}

	ctx, cancel := context.WithCancel(context.Background())
	a := &Actor{
		ID:        id,
		mailbox:   NewMailbox(mailboxSize),
		handler:   handler,
		ctx:       ctx,
		cancel:    cancel,
		isRunning: 1,
	}

	s.actors[id] = a
	go a.eventLoop()
	return a, nil
}

func (a *Actor) eventLoop() {
	for {
		select {
		case <-a.ctx.Done():
			atomic.StoreInt32(&a.isRunning, 0)
			return
		case msg := <-a.mailbox.queue:
			if a.handler != nil {
				_ = a.handler(a.ctx, msg)
			}
		}
	}
}

func (s *System) Send(targetID string, msg Message) error {
	s.mu.RLock()
	target, exists := s.actors[targetID]
	s.mu.RUnlock()

	if !exists {
		return errors.New("actor not found: " + targetID)
	}

	if atomic.LoadInt32(&target.isRunning) == 0 {
		s.routeToDeadLetter(msg)
		return errors.New("actor is stopped")
	}

	select {
	case target.mailbox.queue <- msg:
		return nil
	default:
		s.routeToDeadLetter(msg)
		return errors.New("actor mailbox is full (backpressure triggered, redirected to dead letter)")
	}
}

func (s *System) routeToDeadLetter(msg Message) {
	select {
	case s.deadLetter <- msg:
	default:
		// DeadLetter channel itself saturated, drain one to make room for newest
		select {
		case <-s.deadLetter:
		default:
		}
		select {
		case s.deadLetter <- msg:
		default:
		}
	}
}

func (s *System) StopActor(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, exists := s.actors[id]
	if !exists {
		return errors.New("actor not found: " + id)
	}

	a.cancel()
	delete(s.actors, id)
	return nil
}

// Stop terminates all actors registered within the system
func (s *System) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, a := range s.actors {
		a.cancel()
		delete(s.actors, id)
	}
}
