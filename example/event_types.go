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

func (e *MyEvent) Payload() string {
	return e.payload
}

func NewMyEvent(name string, payload string) *MyEvent {
	return &MyEvent{name: name, payload: payload}
}
