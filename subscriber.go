package goevents

// SubscriberInfo holds the listener function and its priority.
type SubscriberInfo[E Event] struct {
	Listener func(E)
	Priority int
}

// EventSubscriber interface should be implemented by any subscriber.
// It returns a map of event names to a slice of SubscriberInfo.
type EventSubscriber[E Event] interface {
	GetSubscribedEvents() map[string][]SubscriberInfo[E]
}
