// Copyright 2017 F. Alan Jones.  All rights reserved.

// Implement a primative interactive calculator using
// the Hierachical State Machine described here:
// https://en.wikipedia.org/wiki/UML_state_machine

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/watercraft/hsm"
)

const (
	CalcStateOn        hsm.State = "On"
	CalcStateOff       hsm.State = "Off"
	CalcStateOperand1  hsm.State = "Operand1"
	CalcStateOperand2  hsm.State = "Operand2"
	CalcStateOpEntered hsm.State = "OpEntered"
	CalcStateResult    hsm.State = "Result"
)

const (
	CalcEventOnButton  hsm.Event = "OnButton"
	CalcEventOffButton hsm.Event = "OffButton"
	CalcEventClear     hsm.Event = "Clear"
	CalcEventDigit     hsm.Event = "Digit"
	CalcEventOperator  hsm.Event = "Operator"
	CalcEventEqual     hsm.Event = "Equal"
)

type Calc struct {
	hsm.HSMBase
	log                *logrus.Logger
	operand1, operand2 int
	operator           byte
}

func (calc *Calc) Init(name string) {
	calc.log = logrus.StandardLogger()
	// calc.log.Level = logrus.DebugLevel
	calc.HSMBase.Init(name, calc)

	// OFF
	off := calc.NewState(CalcStateOff)
	off.AddTransitions([]hsm.Transition{
		{CalcEventOnButton, CalcStateOn, calc.ActionTurnOn},
	})

	// ON
	on := calc.NewState(CalcStateOn)
	on.AddTransitions([]hsm.Transition{
		{CalcEventOffButton, CalcStateOff, calc.ActionTurnOff},
		{CalcEventClear, CalcStateOn, nil}, // Clear with entry action to "on" state
		{CalcEventDigit, CalcStateOperand1, calc.ActionAppendOperand1},
	})
	// Operand1
	operand1 := calc.NewState(CalcStateOperand1)
	operand1.AddTransitions([]hsm.Transition{
		{CalcEventDigit, CalcStateOperand1, calc.ActionAppendOperand1},
		{CalcEventOperator, CalcStateOpEntered, calc.ActionSetOperator},
	})
	// OpEntered
	opEntered := calc.NewState(CalcStateOpEntered)
	opEntered.AddTransitions([]hsm.Transition{
		{CalcEventDigit, CalcStateOperand2, calc.ActionAppendOperand2},
	})
	// Operand2
	operand2 := calc.NewState(CalcStateOperand2)
	operand2.AddTransitions([]hsm.Transition{
		{CalcEventDigit, CalcStateOperand2, calc.ActionAppendOperand2},
		{CalcEventEqual, CalcStateResult, calc.ActionCompute},
	})
	// Result
	result := calc.NewState(CalcStateResult)
	result.AddTransitions([]hsm.Transition{
		{CalcEventDigit, CalcStateOperand1, calc.ActionSetOperand1},
		{CalcEventOperator, CalcStateOpEntered, calc.ActionSetOperator},
	})
	on.AddChildren(operand1, operand2, opEntered, result)
	on.AddEntryActions(calc.ActionClear)
}

func (calc *Calc) Log() *logrus.Logger {
	return calc.log
}

func (calc *Calc) LogTransition(tran *hsm.Transition, param interface{}) {
	if tran.NewState == CalcStateOff {
		return
	}
	if calc.operator == ' ' {
		fmt.Printf("%6d\n", calc.operand1)
		return
	}
	fmt.Printf("%6d %c %6d\n", calc.operand1, calc.operator, calc.operand2)
}

func intFromDigit(param interface{}) int {
	return int(param.(byte)) - int('0')
}

func (calc *Calc) ActionClear(param interface{}) {
	calc.operand1 = 0
	calc.operand2 = 0
	calc.operator = ' '
}

func (calc *Calc) ActionTurnOn(param interface{}) {
	fmt.Println("Calculator ON")
}

func (calc *Calc) ActionTurnOff(param interface{}) {
	fmt.Println("Calculator OFF")
}

func (calc *Calc) ActionSetOperand1(param interface{}) {
	calc.operand1 = intFromDigit(param)
}

func (calc *Calc) ActionAppendOperand1(param interface{}) {
	calc.operand1 = calc.operand1*10 + intFromDigit(param)
}

func (calc *Calc) ActionSetOperator(param interface{}) {
	calc.operator = param.(byte)
}

func (calc *Calc) ActionAppendOperand2(param interface{}) {
	calc.operand2 = calc.operand2*10 + intFromDigit(param)
}

func (calc *Calc) ActionCompute(param interface{}) {
	switch calc.operator {
	case '+':
		calc.operand1 += calc.operand2
	case '-':
		calc.operand1 -= calc.operand2
	case '*':
		calc.operand1 *= calc.operand2
	case '/':
		calc.operand1 /= calc.operand2
	}
	calc.operand2 = 0
	calc.operator = ' '
}

func SetTty() {
	err := exec.Command("/bin/stty", "-F", "/dev/tty", "-icanon", "min", "1", "-echo").Run()
	if err != nil {
		os.Exit(1)
	}
}

func ClearTty() {
	// Best effort to clear un-cooked input mode
	fmt.Println("Exit")
	_ = exec.Command("/bin/stty", "-F", "/dev/tty", "sane").Run()
}

func HandleSignals(sigCh chan os.Signal) {
	<-sigCh
	ClearTty()
	os.Exit(1)
}

func main() {
	// Set un-cooked input mode to read one character at a time from TTY
	SetTty()
	defer ClearTty()

	// Clear TTY on signals
	sigCh := make(chan os.Signal, 5)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	go func() { HandleSignals(sigCh) }()

	// Initialize calculator and input reader
	var myCalc Calc
	myCalc.Init("CALC")
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to the HSM Calculator!")
	fmt.Println("Perform integer arithmetic with +, -, *, /, =, o=on, f=off, c=clear, q=quit")
	fmt.Println("Calculator OFF")

	// Process input
forloop:
	for {
		c, err := reader.ReadByte()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		}
		// We ignore bad input and illegal transitions, just like your calculator
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			_ = myCalc.Inject(CalcEventDigit, c)
		case '+', '-', '*', '/':
			_ = myCalc.Inject(CalcEventOperator, c)
		case '=':
			_ = myCalc.Inject(CalcEventEqual, c)
		case 'o':
			_ = myCalc.Inject(CalcEventOnButton, c)
		case 'f':
			_ = myCalc.Inject(CalcEventOffButton, c)
		case 'c':
			_ = myCalc.Inject(CalcEventClear, c)
		case 'q':
			break forloop
		}
	}
}
