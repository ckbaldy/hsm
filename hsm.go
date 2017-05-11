// Copyright 2017 F. Alan Jones.  All rights reserved.

package hsm

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/watercraft/errors"
)

// Type for events
type Event string

// Interface for calling into HSM instance
type HSMInstance interface {
	// Get name of instance
	Name() string
	// Get loggger from instance
	Log() *logrus.Logger
	// Call instance to log transition after it is applied
	// The initial state of a superstate can be found in the CurrentState of HSMBase
	LogTransition(from State, tran *Transition, param interface{})
}

// Base object for HSM
type HSMBase struct {
	CurrentState State
	instance     HSMInstance
	states       map[State]*StateInstance
}

// Initialize HSM base object
func (hsm *HSMBase) Init(instance HSMInstance) {
	hsm.instance = instance
	hsm.states = make(map[State]*StateInstance)
	hsm.instance.Log().Debug("Init")
}

func (hsm *HSMBase) lookupState(name State) (*StateInstance, error) {
	state, ok := hsm.states[name]
	if !ok {
		hsm.instance.Log().WithFields(logrus.Fields{
			"hsm":   hsm.instance.Name(),
			"state": name,
		}).Debug("State not found")
		return nil, errors.New(http.StatusInternalServerError,
			fmt.Sprintf("State %s not found for %s", name, hsm.instance.Name()))
	}
	return state, nil
}

// Inject event into HSM, return error is transition is not found
func (hsm *HSMBase) Inject(event Event, param interface{}) error {
	// Find current state
	state, err := hsm.lookupState(hsm.CurrentState)
	if err != nil {
		return err
	}
	// Build list of actions to run if a valid transition is found
	actions := state.exitActions
	// If direct transition found, run it
	tran, ok := state.transitions[event]
	if ok {
		return hsm.applyTransition(tran, actions, param)
	}
	// Walk parents looking for matching transition
	parent := state.parent
	for parent != nil {
		actions := append(actions, parent.exitActions...)
		tran, ok := parent.transitions[event]
		if !ok {
			parent = parent.parent
			continue
		}
		return hsm.applyTransition(tran, actions, param)
	}
	// Match not found
	hsm.instance.Log().WithFields(logrus.Fields{
		"hsm":   hsm.instance.Name(),
		"state": hsm.CurrentState,
		"event": event,
		"param": param,
	}).Debug("Illegal transition")
	return errors.New(http.StatusConflict,
		fmt.Sprintf("Illegal transition for %s from %s on %s with %+v", hsm.instance.Name(), hsm.CurrentState, event, param))
}

// Apply transition to state machine
func (hsm *HSMBase) applyTransition(tran *Transition, actions []ActionFunc, param interface{}) error {
	// Add entry actions for the new state
	newState, err := hsm.lookupState(tran.NewState)
	if err != nil {
		return err
	}
	actions = append(actions, newState.entryActions...)
	// Add entry actions for intial state of superstate
	if newState.initialState != "" {
		newState, err = hsm.lookupState(newState.initialState)
		if err != nil {
			return err
		}
		actions = append(actions, newState.entryActions...)
	}
	// Add action for transition
	if tran.Action != nil {
		actions = append(actions, tran.Action)
	}
	// Run Actions
	for _, action := range actions {
		err := action(param)
		hsm.instance.Log().WithFields(logrus.Fields{
			"hsm":   hsm.instance.Name(),
			"state": hsm.CurrentState,
			"on":    tran.On,
			"new":   tran.NewState,
			"func":  runtime.FuncForPC(reflect.ValueOf(action).Pointer()).Name(),
			"param": param,
			"err":   err,
		}).Debug("Action done")
		if err != nil {
			return err
		}
	}
	// Set New state
	from := hsm.CurrentState
	hsm.instance.Log().WithFields(logrus.Fields{
		"hsm":   hsm.instance.Name(),
		"state": hsm.CurrentState,
		"on":    tran.On,
		"new":   newState.Name,
		"param": param,
	}).Debug("Set state")
	hsm.CurrentState = newState.Name
	// Log transition
	hsm.instance.LogTransition(from, tran, param)
	return nil
}
