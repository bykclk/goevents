package goevents

import (
	"sort"
	"sync"
)

// listenerItem holds a listener and its priority.
type listenerItem[E Event] struct {
	priority int
	listener func(E)
	once     bool
}

// EventDispatcher manages event listeners and subscribers.
type EventDispatcher[E Event] struct {
	listeners map[string][]listenerItem[E]
	mu        sync.RWMutex
}

// NewEventDispatcher creates and returns a new EventDispatcher instance.
func NewEventDispatcher[E Event]() *EventDispatcher[E] {
	return &EventDispatcher[E]{
		listeners: make(map[string][]listenerItem[E]),
	}
}

// AddListener registers a listener for the given event name with a specified priority.
func (d *EventDispatcher[E]) AddListener(eventName string, listener func(E), priority int, once bool) {
	d.listeners[eventName] = append(d.listeners[eventName], listenerItem[E]{
		priority: priority,
		listener: listener,
		once:     once,
	})

	// Sort by priority (high to low)
	sort.Slice(d.listeners[eventName], func(i, j int) bool {
		return d.listeners[eventName][i].priority > d.listeners[eventName][j].priority
	})
}

// AddOnceListener registers a listener that is removed after being invoked once.
func (d *EventDispatcher[E]) AddOnceListener(eventName string, listener func(E), priority int) {
	d.AddListener(eventName, listener, priority, true)
}

// AddSubscriber registers all listeners returned by the subscriber's GetSubscribedEvents method.
func (d *EventDispatcher[E]) AddSubscriber(sub EventSubscriber[E]) {
	events := sub.GetSubscribedEvents()
	for eventName, infos := range events {
		for _, info := range infos {
			d.AddListener(eventName, info.Listener, info.Priority, false)
		}
	}
}

// Dispatch synchronously calls all registered listeners for the event in order of priority.
func (d *EventDispatcher[E]) Dispatch(event E) {
	d.mu.RLock()
	ls, ok := d.listeners[event.Name()]
	d.mu.RUnlock()
	if !ok {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for i := 0; i < len(ls); i++ {
		item := ls[i]
		if event.IsPropagationStopped() {
			break
		}
		item.listener(event)
		if item.once {
			// Remove the listener
			ls = append(ls[:i], ls[i+1:]...)
		}
	}
	d.listeners[event.Name()] = ls
}

// DispatchAsync calls all registered listeners in separate goroutines.
// It uses a WaitGroup to wait until all listeners have finished.
func (d *EventDispatcher[E]) DispatchAsync(event E) {
	if ls, ok := d.listeners[event.Name()]; ok {
		var wg sync.WaitGroup
		for _, item := range ls {
			if event.IsPropagationStopped() {
				break
			}
			wg.Add(1)
			go func(l func(E), ev E) {
				defer wg.Done()
				l(ev)
			}(item.listener, event)
		}
		wg.Wait()
	}
}
