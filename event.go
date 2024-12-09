package goevents

// Event is a constraint that any event must implement.
type Event interface {
	Name() string
	StopPropagation()
	IsPropagationStopped() bool
}
