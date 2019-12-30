package hsm

// State name
type State string

// ActionFunc is a callback for transition, entry and exit actions.
type ActionFunc func(param interface{}) error

// GuardFunc is a callback that returns true if transition is allowed
type GuardFunc func(param interface{}) (bool, error)

// Transition defines an event for the state, the next state following
// the event transition and any action that might occur during the
// event transition.
type Transition struct {
	On       Event
	NewState State
	Action   ActionFunc
	Guard    GuardFunc
}

// StateInstance defines state entry/exit actions and relationships with
// other states in the machine hierarchy.
type StateInstance struct {
	Name         State
	initialState State
	parent       *StateInstance
	transitions  map[Event]*Transition
	entryActions []ActionFunc
	exitActions  []ActionFunc
}

// NewState creates a new state with the hierarchial state machine.
func (hsm *Base) NewState(name State) *StateInstance {
	var state *StateInstance
	if hsm.runState == INITIALIZING {
		state = &StateInstance{}
		state.Name = name
		state.transitions = make(map[Event]*Transition)
		// Set current state to first state added
		if len(hsm.states) == 0 {
			hsm.CurrentState = state.Name
		}
		hsm.states[state.Name] = state
	}
	return state
}

// AddTransitions adds/defines the allowed transitions for a given state.
func (state *StateInstance) AddTransitions(trans []Transition) {
	if state != nil {
		for _, tran := range trans {
			state.transitions[tran.On] = &tran
		}
	}
}

// AddChildren adds child state instances to the current state definition.
func (state *StateInstance) AddChildren(children ...*StateInstance) {
	if state != nil {
		for _, child := range children {
			if state.initialState == "" {
				state.initialState = child.Name
			}
			child.parent = state
		}
	}
}

// AddEntryActions defines one or more entry actions for a given state.
func (state *StateInstance) AddEntryActions(actions ...ActionFunc) {
	if state != nil {
		state.entryActions = append(state.entryActions, actions...)
	}
}

// AddExitActions defines one or more exit actions for a given state.
func (state *StateInstance) AddExitActions(actions ...ActionFunc) {
	if state != nil {
		state.exitActions = append(state.exitActions, actions...)
	}
}
