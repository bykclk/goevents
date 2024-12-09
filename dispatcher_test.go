package goevents

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// MockEvent is a simple implementation of EventInterface for testing purposes.
type MockEvent struct {
	name               string
	payload            string
	propagationStopped bool
}

func (e *MockEvent) Name() string {
	return e.name
}

func (e *MockEvent) StopPropagation() {
	e.propagationStopped = true
}

func (e *MockEvent) IsPropagationStopped() bool {
	return e.propagationStopped
}

func NewMockEvent(name, payload string) *MockEvent {
	return &MockEvent{
		name:    name,
		payload: payload,
	}
}

// MockSubscriber implements EventSubscriber for testing.
type MockSubscriber struct {
	listeners map[string][]SubscriberInfo[*MockEvent]
}

func (ms *MockSubscriber) GetSubscribedEvents() map[string][]SubscriberInfo[*MockEvent] {
	return ms.listeners
}

func TestAddListenerAndDispatch(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	var wg sync.WaitGroup
	wg.Add(2)

	// Listener with priority 10
	dispatcher.AddListener("test.event", func(e *MockEvent) {
		defer wg.Done()
		if e.payload != "Hello, World!" {
			t.Errorf("Expected payload 'Hello, World!', got '%s'", e.payload)
		}
	}, 10, false)

	// Listener with priority 5
	dispatcher.AddListener("test.event", func(e *MockEvent) {
		defer wg.Done()
		if e.payload != "Hello, World!" {
			t.Errorf("Expected payload 'Hello, World!', got '%s'", e.payload)
		}
	}, 5, false)

	event := NewMockEvent("test.event", "Hello, World!")
	dispatcher.Dispatch(event)

	// Wait for all listeners to be called
	wg.Wait()
}

func TestListenerPriority(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	var callOrder []int
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3)

	// Listener with priority 20
	dispatcher.AddListener("priority.event", func(e *MockEvent) {
		defer wg.Done()
		mutex.Lock()
		callOrder = append(callOrder, 1)
		mutex.Unlock()
	}, 20, false)

	// Listener with priority 10
	dispatcher.AddListener("priority.event", func(e *MockEvent) {
		defer wg.Done()
		mutex.Lock()
		callOrder = append(callOrder, 2)
		mutex.Unlock()
	}, 10, false)

	// Listener with priority 30
	dispatcher.AddListener("priority.event", func(e *MockEvent) {
		defer wg.Done()
		mutex.Lock()
		callOrder = append(callOrder, 0)
		mutex.Unlock()
	}, 30, false)

	event := NewMockEvent("priority.event", "Check priority")
	dispatcher.Dispatch(event)

	wg.Wait()

	expectedOrder := []int{0, 1, 2}
	for i, v := range expectedOrder {
		if callOrder[i] != v {
			t.Errorf("Expected listener %d to be called at position %d, got %d", v, i, callOrder[i])
		}
	}
}

func TestOnceListener(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	callCount := 0

	// Once listener
	dispatcher.AddOnceListener("once.event", func(e *MockEvent) {
		fmt.Println("Once listener called")
		callCount++
	}, 10)

	event1 := NewMockEvent("once.event", "First Dispatch")
	dispatcher.Dispatch(event1)

	event2 := NewMockEvent("once.event", "Second Dispatch")
	dispatcher.DispatchAsync(event2)

	if callCount != 1 {
		t.Errorf("Expected once listener to be called once, but was called %d times", callCount)
	}
}

func TestAddSubscriber(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	var wg sync.WaitGroup
	wg.Add(2)

	subscriber := &MockSubscriber{
		listeners: map[string][]SubscriberInfo[*MockEvent]{
			"subscriber.event": {
				{
					Listener: func(e *MockEvent) {
						defer wg.Done()
						if e.payload != "Subscriber Test" {
							t.Errorf("Expected payload 'Subscriber Test', got '%s'", e.payload)
						}
					},
					Priority: 15,
				},
				{
					Listener: func(e *MockEvent) {
						defer wg.Done()
						if e.payload != "Subscriber Test" {
							t.Errorf("Expected payload 'Subscriber Test', got '%s'", e.payload)
						}
					},
					Priority: 5,
				},
			},
		},
	}

	dispatcher.AddSubscriber(subscriber)

	event := NewMockEvent("subscriber.event", "Subscriber Test")
	dispatcher.Dispatch(event)

	wg.Wait()
}

func TestDispatchAsync(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	var wg sync.WaitGroup
	wg.Add(2)

	// Listener 1
	dispatcher.AddListener("async.event", func(e *MockEvent) {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Simulate work
		if e.payload != "Async Test" {
			t.Errorf("Expected payload 'Async Test', got '%s'", e.payload)
		}
	}, 10, false)

	// Listener 2
	dispatcher.AddListener("async.event", func(e *MockEvent) {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond) // Simulate work
		if e.payload != "Async Test" {
			t.Errorf("Expected payload 'Async Test', got '%s'", e.payload)
		}
	}, 5, false)

	event := NewMockEvent("async.event", "Async Test")
	start := time.Now()
	dispatcher.DispatchAsync(event)
	duration := time.Since(start)

	// Ensure that DispatchAsync returns before listeners complete
	if duration > 200*time.Millisecond {
		t.Errorf("DispatchAsync took too long: %v", duration)
	}

	// Wait for listeners to complete
	wg.Wait()
}

func TestStopPropagation(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	var wg sync.WaitGroup

	// Listener 1
	dispatcher.AddListener("stop.event", func(e *MockEvent) {
		e.StopPropagation()
	}, 10, false)

	// Listener 2
	dispatcher.AddListener("stop.event", func(e *MockEvent) {
	}, 5, false)

	// Listener 3 - Should not be called
	dispatcher.AddListener("stop.event", func(e *MockEvent) {
		t.Errorf("Listener 3 should not be called due to propagation stop")
	}, 1, false)

	event := NewMockEvent("stop.event", "Stop Test")
	dispatcher.Dispatch(event)

	wg.Wait()
}

func TestNoListeners(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	// Dispatch an event with no listeners
	event := NewMockEvent("no.listener.event", "No Listener Test")
	dispatcher.Dispatch(event)

	// If no panic occurs, the test passes
}

func TestMultipleSubscribers(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	subscriber1 := &MockSubscriber{
		listeners: map[string][]SubscriberInfo[*MockEvent]{
			"multi.event": {
				{
					Listener: func(e *MockEvent) {
						if e.payload != "Multi Subscriber Test" {
							t.Errorf("Subscriber1: Expected payload 'Multi Subscriber Test', got '%s'", e.payload)
						}
					},
					Priority: 20,
				},
			},
		},
	}

	subscriber2 := &MockSubscriber{
		listeners: map[string][]SubscriberInfo[*MockEvent]{
			"multi.event": {
				{
					Listener: func(e *MockEvent) {
						if e.payload != "Multi Subscriber Test" {
							t.Errorf("Subscriber2: Expected payload 'Multi Subscriber Test', got '%s'", e.payload)
						}
					},
					Priority: 10,
				},
			},
		},
	}

	dispatcher.AddSubscriber(subscriber1)
	dispatcher.AddSubscriber(subscriber2)

	event := NewMockEvent("multi.event", "Multi Subscriber Test")
	dispatcher.Dispatch(event)
}

func TestListenerExecutionOrder(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	var executionOrder []int
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3)

	// Listener 1
	dispatcher.AddListener("order.event", func(e *MockEvent) {
		defer wg.Done()
		mutex.Lock()
		executionOrder = append(executionOrder, 1)
		mutex.Unlock()
	}, 10, false)

	// Listener 2
	dispatcher.AddListener("order.event", func(e *MockEvent) {
		defer wg.Done()
		mutex.Lock()
		executionOrder = append(executionOrder, 2)
		mutex.Unlock()
	}, 20, false)

	// Listener 3
	dispatcher.AddListener("order.event", func(e *MockEvent) {
		defer wg.Done()
		mutex.Lock()
		executionOrder = append(executionOrder, 3)
		mutex.Unlock()
	}, 15, false)

	event := NewMockEvent("order.event", "Execution Order Test")
	dispatcher.Dispatch(event)

	wg.Wait()

	expectedOrder := []int{2, 3, 1} // Priorities: 20, 15, 10
	for i, v := range expectedOrder {
		if executionOrder[i] != v {
			t.Errorf("Expected listener %d to be called at position %d, got %d", v, i, executionOrder[i])
		}
	}
}

func TestListenerWithSamePriority(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	var callOrder []int
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	// Both listeners have the same priority
	dispatcher.AddListener("same.priority.event", func(e *MockEvent) {
		defer wg.Done()
		mutex.Lock()
		callOrder = append(callOrder, 1)
		mutex.Unlock()
	}, 10, false)

	dispatcher.AddListener("same.priority.event", func(e *MockEvent) {
		defer wg.Done()
		mutex.Lock()
		callOrder = append(callOrder, 2)
		mutex.Unlock()
	}, 10, false)

	event := NewMockEvent("same.priority.event", "Same Priority Test")
	dispatcher.Dispatch(event)

	wg.Wait()

	// Since priorities are the same, the order should be the same as addition
	expectedOrder := []int{1, 2}
	for i, v := range expectedOrder {
		if callOrder[i] != v {
			t.Errorf("Expected listener %d to be called at position %d, got %d", v, i, callOrder[i])
		}
	}
}

func TestConcurrentDispatch(t *testing.T) {
	dispatcher := NewEventDispatcher[*MockEvent]()

	var wg sync.WaitGroup
	wg.Add(100)

	// Add a single listener
	dispatcher.AddListener("concurrent.event", func(e *MockEvent) {
		defer wg.Done()
	}, 10, false)

	// Dispatch the same event concurrently 100 times
	for i := 0; i < 100; i++ {
		go dispatcher.Dispatch(NewMockEvent("concurrent.event", "Concurrent Test"))
	}

	// Wait for all dispatches to complete
	wg.Wait()
}
