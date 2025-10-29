// Package emitter provides a generic, concurrent-safe event emitter.
// It allows registering callbacks for events and emitting events with arguments.
package emitter

import (
	"reflect"
	"sync"
)

// EventEmitter is a struct that manages event listeners and emits events.
// It uses reflection to call callbacks with variable arguments.
type EventEmitter struct {
	listeners     sync.Map // map[string][]reflect.Value
	onceListeners sync.Map // map[string][]reflect.Value
}

// On registers a callback function to be called whenever the specified event is emitted.
// The callback must be a function. Panics if callback is not a function.
func (e *EventEmitter) On(event string, callback any) {
	val := reflect.ValueOf(callback)
	if val.Kind() != reflect.Func {
		panic("callback must be a function")
	}

	actual, _ := e.listeners.LoadOrStore(event, []reflect.Value{})
	list := actual.([]reflect.Value)
	list = append(list, val)

	e.listeners.Store(event, list)
}

// Once registers a callback function to be called only once when the specified event is emitted.
// After emission, the callback is automatically removed.
func (e *EventEmitter) Once(event string, callback any) {
	val := reflect.ValueOf(callback)
	if val.Kind() != reflect.Func {
		panic("callback must be a function")
	}

	actual, _ := e.onceListeners.LoadOrStore(event, []reflect.Value{})
	list := actual.([]reflect.Value)
	list = append(list, val)

	e.onceListeners.Store(event, list)
}

// Off removes a previously registered callback for the specified event.
// It removes from both regular and once listeners.
func (e *EventEmitter) Off(event string, callback any) {
	val := reflect.ValueOf(callback)
	actual, ok := e.listeners.Load(event)
	if !ok {
		return
	}

	list := actual.([]reflect.Value)
	newList := make([]reflect.Value, 0, len(list))
	for _, v := range list {
		if v.Pointer() != val.Pointer() {
			newList = append(newList, v)
		}
	}

	e.listeners.Store(event, newList)

	// Also remove from onceListeners
	actual, ok = e.onceListeners.Load(event)
	if !ok {
		return
	}

	list = actual.([]reflect.Value)
	newList = make([]reflect.Value, 0, len(list))
	for _, v := range list {
		if v.Pointer() != val.Pointer() {
			newList = append(newList, v)
		}
	}

	e.onceListeners.Store(event, newList)
}

// Emit triggers all registered callbacks for the specified event, passing the provided arguments.
// It handles both regular and once listeners, with panic recovery for each callback.
func (e *EventEmitter) Emit(event string, args ...any) {
	reflectedArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		reflectedArgs[i] = reflect.ValueOf(arg)
	}

	// Handle once listeners
	if actual, ok := e.onceListeners.LoadAndDelete(event); ok {
		list := actual.([]reflect.Value)
		for _, listener := range list {
			func() {
				defer func() {
					recover()
				}()

				listener.Call(reflectedArgs)
			}()
		}
	}

	// Handle regular listeners
	if actual, ok := e.listeners.Load(event); ok {
		list := actual.([]reflect.Value)
		for _, listener := range list {
			func() {
				defer func() {
					recover()
				}()

				listener.Call(reflectedArgs)
			}()
		}
	}
}

// GetCallbackType returns the reflect.Type of the first callback registered for the event.
// Returns nil if no callbacks are registered.
func (e *EventEmitter) GetCallbackType(event string) reflect.Type {
	if listeners, ok := e.listeners.Load(event); ok {
		list := listeners.([]reflect.Value)
		if len(list) > 0 {
			return list[0].Type()
		}
	}

	return nil
}
