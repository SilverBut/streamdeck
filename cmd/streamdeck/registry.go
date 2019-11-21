package main

import (
	"reflect"
	"sync"

	"github.com/pkg/errors"
)

type action interface {
	Execute(attributes map[string]interface{}) error
}

type displayElement interface {
	Display(idx int, attributes map[string]interface{}) error
}

type refreshingDisplayElement interface {
	StartLoopDisplay(idx int, attributes map[string]interface{}) error
	StopLoopDisplay() error
}

var (
	registeredActions             = map[string]reflect.Type{}
	registeredActionsLock         sync.Mutex
	registeredDisplayElements     = map[string]reflect.Type{}
	registeredDisplayElementsLock sync.Mutex
)

func registerAction(name string, handler action) {
	registeredActionsLock.Lock()
	defer registeredActionsLock.Unlock()

	registeredActions[name] = reflect.TypeOf(handler)
}

func registerDisplayElement(name string, handler displayElement) {
	registeredDisplayElementsLock.Lock()
	defer registeredDisplayElementsLock.Unlock()

	registeredDisplayElements[name] = reflect.TypeOf(handler)
}

func callAction(a dynamicElement) error {
	t, ok := registeredActions[a.Type]
	if !ok {
		return errors.Errorf("Unknown action type %q", a.Type)
	}

	inst := reflect.New(t).Interface().(action)

	return inst.Execute(a.Attributes)
}

func callDisplayElement(idx int, kd keyDefinition) error {
	t, ok := registeredDisplayElements[kd.Display.Type]
	if !ok {
		return errors.Errorf("Unknown display type %q", kd.Display.Type)
	}

	var inst interface{}
	if t.Kind() == reflect.Ptr {
		inst = reflect.New(t.Elem()).Interface()
	} else {
		inst = reflect.New(t).Interface()
	}

	if t.Implements(reflect.TypeOf((*refreshingDisplayElement)(nil)).Elem()) {
		activeLoops = append(activeLoops, inst.(refreshingDisplayElement))
		return inst.(refreshingDisplayElement).StartLoopDisplay(idx, kd.Display.Attributes)
	}

	return inst.(displayElement).Display(idx, kd.Display.Attributes)
}
