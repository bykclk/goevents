package main

import (
	"fmt"
	"github.com/bykclk/goevents"
)

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
