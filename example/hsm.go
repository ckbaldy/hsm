package example

import (
	"github.com/ckbaldy/hsm"
	"github.com/ckbaldy/hsm/example/logger"
	"github.com/sirupsen/logrus"
)

// States
const (
	S0   hsm.State = "s0"
	S1   hsm.State = "s1"
	S11  hsm.State = "s11"
	S2   hsm.State = "s2"
	S21  hsm.State = "s21"
	S211 hsm.State = "s211"
)

// Events
const (
	EventA hsm.Event = "a"
	EventB hsm.Event = "b"
	EventC hsm.Event = "c"
	EventD hsm.Event = "d"
	EventE hsm.Event = "e"
	EventF hsm.Event = "f"
	EventG hsm.Event = "g"
	EventH hsm.Event = "h"
)

var log = logger.Log

// HSM type, structure extends the base HSM
type HSM struct {
	hsm.Base
	Foo bool
}

// NewHSM creates a new annotated example, hierarchial state machine.
func NewHSM(name string) *HSM {

	sm := &HSM{}
	sm.Configure(name)
	log.SetLevel(logrus.ErrorLevel)
	sm.AddLogger(log)

	// State S0
	s0 := sm.NewState(S0)
	s0.AddTransitions([]hsm.Transition{
		{On: EventE, NewState: S211, Action: nil},
	})
	s0.AddEntryActions(sm.StateS0Entry)
	s0.AddExitActions(sm.StateS0Exit)

	// State S1
	s1 := sm.NewState(S1)
	s1.AddTransitions([]hsm.Transition{
		{On: EventA, NewState: S1, Action: nil},
	})
	s1.AddEntryActions(sm.StateS1Entry)
	s1.AddExitActions(sm.StateS1Exit)

	// State S11
	s11 := sm.NewState(S11)
	s11.AddTransitions([]hsm.Transition{
		// Internal transition, NewState is empty string
		{On: EventH, Action: sm.EventHaction, Guard: sm.EventHguard},
	})
	s11.AddEntryActions(sm.StateS11Entry)
	s11.AddExitActions(sm.StateS11Exit)

	// State S2
	s2 := sm.NewState(S2)
	s2.AddTransitions([]hsm.Transition{
		{On: EventC, NewState: S1, Action: nil},
	})
	s2.AddEntryActions(sm.StateS2Entry)
	s2.AddExitActions(sm.StateS2Exit)

	// State S21
	s21 := sm.NewState(S21)
	s21.AddTransitions([]hsm.Transition{
		// Self transition
		{On: EventH, NewState: S21, Action: sm.EventHaction, Guard: sm.EventHguard},
	})
	s21.AddEntryActions(sm.StateS21Entry)
	s21.AddExitActions(sm.StateS21Exit)

	// State S211
	s211 := sm.NewState(S211)
	s211.AddTransitions([]hsm.Transition{})
	s211.AddEntryActions(sm.StateS211Entry)
	s211.AddExitActions(sm.StateS211Exit)

	// Add children.
	s0.AddChildren(s1, s2)
	s1.AddChildren(s11)
	s2.AddChildren(s21)
	s21.AddChildren(s211)

	sm.Finalize()

	return sm
}
