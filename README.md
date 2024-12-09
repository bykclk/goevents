# Go Event Dispatcher

A simple and flexible event dispatcher library for Go, inspired by Symfony's EventDispatcher. It supports synchronous and asynchronous event dispatching, event subscribers, listener prioritization, and more.

## Features

- **Synchronous and Asynchronous Dispatching:** Choose between blocking and non-blocking event dispatch.
- **Event Subscribers:** Group multiple listeners into subscribers for better organization.
- **Listener Prioritization:** Control the order in which listeners are invoked.
- **"Once" Listeners:** Register listeners that are automatically removed after being invoked once.
- **Generics Support:** Type-safe event and listener handling using Go's generics (Go 1.18+).

## Installation

```bash
go get github.com/bykclk/go-event-dispatcher
```

## Usage

### Defining Events

Implement the EventInterface for your custom event.

```go
package main

// MyEvent is a concrete event type implementing EventInterface.
type MyEvent struct {
	name               string
	payload            string
	propagationStopped bool
}

func (e *MyEvent) Name() string {
	return e.name
}

func (e *MyEvent) StopPropagation() {
	e.propagationStopped = true
}

func (e *MyEvent) IsPropagationStopped() bool {
	return e.propagationStopped
}

func NewMyEvent(name string, payload string) *MyEvent {
	return &MyEvent{name: name, payload: payload}
}
```

### Create a Subscriber

Create a Subscriber

```go   
// MySubscriber is an example subscriber.
type MySubscriber struct{}

func (ms MySubscriber) GetSubscribedEvents() map[string][]goevents.SubscriberInfo[*MyEvent] {
	return map[string][]goevents.SubscriberInfo[*MyEvent]{
		"user.created": {
			{
				Listener: func(e *MyEvent) {
					fmt.Println("MySubscriber: User created event:", e.Payload())
				},
				Priority: 10,
			},
			{
				Listener: func(e *MyEvent) {
					fmt.Println("MySubscriber: Another listener for user.created")
				},
				Priority: 5,
			},
		},
	}
}
```

### Initialize the Dispatcher and Dispatch Events

```go
func main() {
	dispatcher := goevents.NewEventDispatcher[*MyEvent]()

	// Add subscriber
	dispatcher.AddSubscriber(MySubscriber{})

	// Add manual listener
	dispatcher.AddListener("user.created", func(e *MyEvent) {
		fmt.Println("Manual listener: user.created, payload:", e.Payload())
	}, 0, false)

	// Dispatch synchronously
	event := NewMyEvent("user.created", "John Doe")
	dispatcher.Dispatch(event)

	// Dispatch asynchronously
	event2 := NewMyEvent("user.created", "Jane Doe")
	dispatcher.DispatchAsync(event2)
}
```