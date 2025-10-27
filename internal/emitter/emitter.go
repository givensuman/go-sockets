// Package emitter provides an event emitter implementation.
package emitter

import (
	"reflect"
	"sync"
)

type EventEmitter struct {
	listeners     sync.Map // map[string][]reflect.Value
	onceListeners sync.Map // map[string][]reflect.Value
}

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
