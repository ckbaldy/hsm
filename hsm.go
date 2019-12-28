// Copyright 2017 F. Alan Jones.  All rights reserved.

package hsm

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"

	"github.com/sirupsen/logrus"
)

// Event is the type for triggers
type Event string

// hsmConfigState enumeration type that reflects the internal,
// hsm run-level, configuration state.
type hsmConfigState int

const (
	initializing hsmConfigState = iota
	finalized
	on
	exiting
	off
)

const (
	// InitEvent causes the default transition from initial pseudostate
	hsmInitEvent Event = "InitialTransition"
	// ExitEvent causes the default transition to the final pseudostate
	hsmExitEvent Event = "ExitTransition"
	// TopState manages entry/exit into/from the define top state machine
	hsmTopState State = "TopState"
)

// Base object for HSM
type Base struct {
	Name         string
	CurrentState State
	states       map[State]*StateInstance
	topState     *StateInstance
	runState     hsmConfigState
	logger       *logrus.Logger
	log          *logrus.Entry
}

var topState *StateInstance

// TODO:  add Off method which runs all exit actions and puts the
// state machine back in the state so it cannot process events.
// On needs be run to turn it back on.

// TODO:  add internal machine states.
// 1. can only configure in confure state.
// 2. after finalize is run, cannot change state machine config
// 3. On.  Can process events.
// 4. Off.  cannot process events

// Configure initializes the state machine, creating a state machine map
// This must be called once prior to intializing or defining the states,
// their transitions and child/parent relations.
func (hsm *Base) Configure(name string) {
	if hsm.states == nil {
		hsm.Name = name
		hsm.states = make(map[State]*StateInstance)
		hsm.DisableLogger()
	}
	hsm.runState = initializing
}

// AddLogger adds an externally defined logrus logger,  enabling logging of
// state machine transitions.
func (hsm *Base) AddLogger(externalLogger *logrus.Logger) {
	hsm.logger = externalLogger
	hsm.addLoggerEntry()
}

// DisableLogger disables state machine transition logging, removing
// the previously added logger.
func (hsm *Base) DisableLogger() {
	hsm.logger = logrus.New()
	hsm.logger.Out = ioutil.Discard
	hsm.addLoggerEntry()
}

// Finalize verifies the top state and disables further configuration.
// Configuration can later be changed by calling 'Configure' again.
// An error is returned if children relationships have not be created.
func (hsm *Base) Finalize() error {

	var err error
	var numStatesWithParent int
	topStates := []*StateInstance{}

	for _, state := range hsm.states {
		if state.parent == nil {
			topStates = append(topStates, state)
		} else {
			// Parent defined in state
			numStatesWithParent++
		}
	}
	if len(topStates) == 1 && numStatesWithParent == (len(hsm.states)-1) {

		topState = hsm.NewState(hsmTopState)
		topState.AddTransitions([]Transition{
			{On: hsmInitEvent, NewState: topState.Name}})
		topStates[0].AddTransitions([]Transition{
			{On: hsmExitEvent, NewState: topState.Name}})
		// The initial child state is the user configured, top state.
		topState.AddChildren(topStates[0])

		hsm.CurrentState = topState.Name
		hsm.runState = finalized

	} else if numStatesWithParent == 0 {
		err = fmt.Errorf("no child states were added in %s", hsm.Name)
		hsm.log.Error(err)

	} else if len(topStates) > 1 {
		// There is more than one state with an undefined parent.  This
		// is not allowed.
		var topStateNames State
		for _, state := range topStates {
			topStateNames += state.Name + ", "
		}
		err = fmt.Errorf("found more than one top state in %s %s",
			hsm.Name, topStateNames)
		hsm.log.Error(err)
		hsm.log.Errorf("there needs to be a parent defined for every state" +
			" except the top state")
	}
	return err
}

// On starts a finalized state machine, using the initial, default transtition
// for the top state.
func (hsm *Base) On() error {
	var err error
	if hsm.runState == finalized || hsm.runState == off {
		currentState, err := hsm.lookupState(hsm.CurrentState)
		if err != nil {
			return err
		}
		if currentState == topState {
			hsm.runState = on
			hsm.Inject(hsmInitEvent, nil)
		}
	} else {
		err = fmt.Errorf("cannot start hsm %s; it is not finalized", hsm.Name)
		hsm.log.Error(err)
	}
	return err
}

// Off exits (turns off) the state machine, running all the required exit
// actions and setting the hsm run state to off.
func (hsm *Base) Off() error {
	var err error
	if hsm.runState == on {
		currentState, err := hsm.lookupState(hsm.CurrentState)
		if err != nil {
			return err
		}
		if currentState != topState {
			hsm.runState = exiting
			hsm.Inject(hsmExitEvent, nil)
			hsm.runState = off
		}
	} else {
		err = fmt.Errorf("cannot start hsm %s; it is not finalized", hsm.Name)
		hsm.log.Error(err)
	}
	return err
}

func (hsm *Base) addLoggerEntry() {
	hsm.log = hsm.logger.WithFields(logrus.Fields{"prefix": hsm.Name})
}

func (hsm *Base) lookupState(name State) (*StateInstance, error) {
	state, ok := hsm.states[name]
	if !ok {
		hsm.log.WithFields(logrus.Fields{
			"state": name,
		}).Debug("state not found")
		return nil, fmt.Errorf("state %s not found for %s", name, hsm.Name)
	}
	return state, nil
}

// Inject event into HSM. Return an error if the event/transition is not found.
func (hsm *Base) Inject(event Event, param interface{}) error {

	if hsm.runState != on && hsm.runState != exiting {
		err := fmt.Errorf("cannot inject events into hsm %s; it is not on",
			hsm.Name)
		hsm.log.Error(err)
		return err
	}

	// Find the first composite state in the state tree that has a defined
	// transition for the event.
	sourceState, tran, exitActions, err := hsm.eventSource(event)
	if err != nil {
		return err
	}

	// A transiton was found.  If transition has a guard and the guard
	// returns true, apply the transition.
	if tran.Guard != nil {
		tranAllowed, err := tran.Guard(param)
		if err != nil {
			hsm.logAction("guard function failed", tran, tran.Guard, param)
			return err
		}
		if !tranAllowed {
			hsm.logAction("transition guarded", tran, tran.Guard, param)
			return nil
		}
	}
	return hsm.applyTransition(tran, exitActions, param, sourceState)
}

// eventSource searches the composite state tree for a state that can handle
// (has a transition defined for) the event. The search begins with the current
// state  and proceeds up the state tree until either a transition is found or
// the top state is reached. If an event transition is found, it returns the
// source state handling the event, the event transition and the exit actions
// for states in the state tree, starting at the current state up to and
// including  the source state.  An 'unhandled event' error is returned if a
// transition is not found.
func (hsm *Base) eventSource(event Event) (*StateInstance,
	*Transition, []ActionFunc, error) {

	currentState, err := hsm.lookupState(hsm.CurrentState)
	if err != nil {
		return nil, nil, nil, err
	}

	// If there is a direct transition for the current state ...
	tran, ok := currentState.transitions[event]
	if ok {
		return currentState, tran, currentState.exitActions, nil
	}

	// Walk parents looking for a matching transition event for the state,
	// appending any exit actions required while progressing up the tree.
	parent := currentState.parent
	exitActions := currentState.exitActions
	for parent != nil {
		tran, ok := parent.transitions[event]
		if !ok {
			// If event is not defined for the state, append exit actions
			// for the parent and proceed up the tree.
			exitActions = append(exitActions, parent.exitActions...)
			parent = parent.parent
			continue
		} else {
			// Event found.
			sourceState := parent
			return sourceState, tran, exitActions, nil
		}
	}
	// Top state reached. A matching event was not found in the state tree.
	hsm.log.WithFields(logrus.Fields{
		"state": hsm.CurrentState,
		"on":    event,
	}).Error("unhandled event/transition")
	err = fmt.Errorf("unhandled event/transition in %s,"+
		"current state:%s, event: %s", hsm.Name, hsm.CurrentState, event)
	return nil, nil, nil, err
}

// leastCommonAncestor finds the exit and entry actions for the
// transition based on the common parent or the least common ancestor (LCA)
// for the source and target states.
func (hsm *Base) leastCommonAncestor(sourceState *StateInstance,
	targetState *StateInstance) ([]ActionFunc, []ActionFunc, error) {

	var err error
	var lca *StateInstance

	exitActions := []ActionFunc{}
	entryActions := []ActionFunc{}

	// If this is a self transition
	if sourceState == targetState {
		exitActions = sourceState.exitActions
		entryActions = sourceState.entryActions
		return exitActions, entryActions, nil
	}

	// Look for least common ancestor (LCA) state
	for sourceParent := sourceState; lca == nil; {
		for targetParent := targetState; targetParent != nil; {
			if targetParent == sourceParent {
				lca = targetParent
				break
			}
			targetParent = targetParent.parent
		}
		if lca != nil {
			break
		}
		// Add exit actions while looking for the LCA
		exitActions = append(exitActions, sourceParent.exitActions...)
		sourceParent = sourceParent.parent
	}

	// Add entry actions for states up to by not including the LCA state.
	for targetParent := targetState; targetParent != lca; {
		entryActions = append(targetParent.entryActions, entryActions...)
		targetParent = targetParent.parent
	}

	return exitActions, entryActions, err
}

// Apply transition to state machine
func (hsm *Base) applyTransition(tran *Transition, exitActions []ActionFunc,
	param interface{}, sourceState *StateInstance) error {

	// If internal transition, only execute the transition action and return.
	if tran.NewState == "" {
		var err error
		if tran.Action != nil {
			err = tran.Action(param)
			hsm.logAction("internal/", tran, tran.Action, param)
		}
		return err
	}

	targetState, _ := hsm.lookupState(tran.NewState)
	defaultStateName := targetState.initialState
	finalStateName := targetState.Name
	entryActions := []ActionFunc{}

	// Get entry and/or exit actions required for the transition.
	if sourceState.parent != nil && hsm.runState != off {
		// Not in top state so use LCA to collect exit/entry actions
		exitActionsFromSourceState, lcaEntryActions, err :=
			hsm.leastCommonAncestor(sourceState, targetState)
		if err != nil {
			hsm.logAction("transition failed", tran, tran.Action, param)
			hsm.log.Error(err)
			return err
		}
		// Append exit actions from target state through the source state up to,
		// but not including the least common ancestor (LCA) state.
		exitActions = append(exitActions, exitActionsFromSourceState...)
		entryActions = lcaEntryActions
	} else {
		// In Top state
		// Handle the default entry actions, starting from top state.
		defaultStateName = sourceState.initialState
	}

	// Add entry actions associated with default transitons
	if hsm.runState != exiting {
		for defaultStateName != "" {
			defaultState, err := hsm.lookupState(defaultStateName)
			if err != nil {
				return err
			}
			entryActions = append(entryActions, defaultState.entryActions...)
			finalStateName = defaultStateName
			defaultStateName = defaultState.initialState
		}
	}

	// Run exit actions
	for _, action := range exitActions {
		err := action(param)
		hsm.logAction("exit/    ", tran, action, param)
		if err != nil {
			return err
		}
	}

	// Run the transition action
	if tran.Action != nil {
		err := tran.Action(param)
		hsm.logAction("tran/    ", tran, tran.Action, param)
		if err != nil {
			return err
		}
	}

	// Run entry actions
	for _, action := range entryActions {
		err := action(param)
		hsm.logAction("entry/   ", tran, action, param)
		if err != nil {
			return err
		}
	}

	// Set New state
	from := hsm.CurrentState
	hsm.CurrentState = finalStateName
	// Log transition
	hsm.log.WithFields(logrus.Fields{
		"<state": from,
		">state": finalStateName,
		"on":     tran.On,
	}).Debug("set state")
	return nil
}

func (hsm *Base) logAction(actionType string, tran *Transition, fn interface{}, param interface{}) {
	// TODO:  strip full path off name, as it is noisy and not needed.
	fnName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	hsm.log.WithFields(logrus.Fields{
		"<state": hsm.CurrentState,
		">state": tran.NewState,
		"on":     tran.On,
		"action": fnName,
		"param":  param,
	}).Debug(actionType)
}
