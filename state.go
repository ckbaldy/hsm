package hsm

type State string

type ActionFunc func(param interface{})

type Transition struct {
	On       Event
	NewState State
	Action   ActionFunc
}

type StateInstance struct {
	Name         State
	hsm          *HSMBase
	parent       *StateInstance
	transitions  map[Event]*Transition
	entryActions []ActionFunc
	exitActions  []ActionFunc
}

func (hsm *HSMBase) NewState(name State) *StateInstance {
	state := &StateInstance{Name: name}
	state.hsm = hsm
	state.transitions = make(map[Event]*Transition)
	// Set current state to first state added
	if len(hsm.states) == 0 {
		hsm.CurrentState = state.Name
	}
	hsm.states[state.Name] = state
	return state
}

func (state *StateInstance) AddTransitions(trans []Transition) {
	for _, tran := range trans {
		tmp := tran
		state.transitions[tran.On] = &tmp
	}
}

func (state *StateInstance) AddChildren(children ...*StateInstance) {
	for _, child := range children {
		child.parent = state
	}
}

// Entry actions are invoked only for the new state of a transition
func (state *StateInstance) AddEntryActions(actions ...ActionFunc) {
	state.entryActions = append(state.entryActions, actions...)
}

// Exit actions are accumulated as you walk up through parent states to find a tranition
func (state *StateInstance) AddExitActions(actions ...ActionFunc) {
	state.exitActions = append(state.exitActions, actions...)
}
