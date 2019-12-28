package example

import (
	"fmt"
)

// StateS0Entry entry actions
func (sm *HSM) StateS0Entry(param interface{}) error {
	fmt.Println("\t  S0 Entry")
	return nil
}

// StateS0Exit entry actions
func (sm *HSM) StateS0Exit(param interface{}) error {
	fmt.Println("\t  S0 Exit")
	return nil
}

// StateS1Entry entry actions
func (sm *HSM) StateS1Entry(param interface{}) error {
	fmt.Println("\t  S1 Entry")
	return nil
}

// StateS1Exit entry actions
func (sm *HSM) StateS1Exit(param interface{}) error {
	fmt.Println("\t  S1 Exit")
	return nil
}

// StateS11Entry entry actions
func (sm *HSM) StateS11Entry(param interface{}) error {
	fmt.Println("\t  S11 Entry")
	return nil
}

// StateS11Exit entry actions
func (sm *HSM) StateS11Exit(param interface{}) error {
	fmt.Println("\t  S11 Exit")
	return nil
}

// StateS2Entry entry actions
func (sm *HSM) StateS2Entry(param interface{}) error {
	fmt.Println("\t  S2 Entry")
	return nil
}

// StateS2Exit entry actions
func (sm *HSM) StateS2Exit(param interface{}) error {
	fmt.Println("\t  S2 Exit")
	return nil
}

// StateS21Entry entry actions
func (sm *HSM) StateS21Entry(param interface{}) error {
	fmt.Println("\t  S21 Entry")
	return nil
}

// StateS21Exit entry actions
func (sm *HSM) StateS21Exit(param interface{}) error {
	fmt.Println("\t  S21 Exit")
	return nil
}

// StateS211Entry entry actions
func (sm *HSM) StateS211Entry(param interface{}) error {
	fmt.Println("\t  S211 Entry")
	return nil
}

// StateS211Exit entry actions
func (sm *HSM) StateS211Exit(param interface{}) error {
	fmt.Println("\t  S211 Exit")
	return nil
}
