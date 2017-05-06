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
	// Get loggger from instance
	Log() *logrus.Logger
	// Call instance to log transition before it is applied
	LogTransition(tran *Transition, param interface{})
}

// Base object for HSM
type HSMBase struct {
	Name         string
	CurrentState State
	instance     HSMInstance
	states       map[State]*StateInstance
}

// Initialize HSM base object
func (hsm *HSMBase) Init(name string, instance HSMInstance) {
	hsm.instance = instance
	hsm.Name = name
	hsm.states = make(map[State]*StateInstance)
	hsm.instance.Log().WithFields(logrus.Fields{
		"hsm": hsm.Name,
	}).Debug("Init")
}

// Inject event into HSM, return error is transition is not found
func (hsm *HSMBase) Inject(event Event, param interface{}) error {
	// Find current state
	state, ok := hsm.states[hsm.CurrentState]
	if !ok {
		hsm.instance.Log().WithFields(logrus.Fields{
			"hsm":   hsm.Name,
			"state": hsm.CurrentState,
			"event": event,
			"param": param,
		}).Debug("Current state not found")
		return errors.New(http.StatusInternalServerError,
			fmt.Sprintf("Current state not found for %s from %s on %s with %+v", hsm.Name, hsm.CurrentState, event, param))
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
		"hsm":   hsm.Name,
		"state": hsm.CurrentState,
		"event": event,
		"param": param,
	}).Debug("Illegal transition")
	return errors.New(http.StatusConflict,
		fmt.Sprintf("Illegal transition for %s from %s on %s with %+v", hsm.Name, hsm.CurrentState, event, param))
}

// Apply transition to state machine
func (hsm *HSMBase) applyTransition(tran *Transition, actions []ActionFunc, param interface{}) error {
	// Add action for transition
	if tran.Action != nil {
		actions = append(actions, tran.Action)
	}
	// Add entry actions
	newState, ok := hsm.states[tran.NewState]
	if !ok {
		hsm.instance.Log().WithFields(logrus.Fields{
			"hsm":   hsm.Name,
			"state": hsm.CurrentState,
			"on":    tran.On,
			"new":   tran.NewState,
			"param": param,
		}).Debug("New state not found")
		return errors.New(http.StatusInternalServerError,
			fmt.Sprintf("New state not found for %s from %s on %s with %+v", hsm.Name, hsm.CurrentState, tran.On, param))
	}
	actions = append(actions, newState.entryActions...)
	// Run Actions
	for _, action := range actions {
		hsm.instance.Log().WithFields(logrus.Fields{
			"hsm":   hsm.Name,
			"state": hsm.CurrentState,
			"on":    tran.On,
			"new":   tran.NewState,
			"func":  runtime.FuncForPC(reflect.ValueOf(action).Pointer()).Name(),
			"param": param,
		}).Debug("Run action")
		action(param)
	}
	// Log transition
	hsm.instance.LogTransition(tran, param)
	// Set New state
	hsm.instance.Log().WithFields(logrus.Fields{
		"hsm":   hsm.Name,
		"state": hsm.CurrentState,
		"on":    tran.On,
		"new":   tran.NewState,
		"param": param,
	}).Debug("Set state")
	hsm.CurrentState = tran.NewState
	return nil
}
