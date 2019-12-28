package example

// EventHguard guards transition on event h
func (sm *HSM) EventHguard(param interface{}) (bool, error) {
	foo := param.(bool)
	return foo, nil
}

// EventHaction is an action that occures on the event-H transition
func (sm *HSM) EventHaction(param interface{}) error {
	sm.Foo = !sm.Foo
	return nil
}
