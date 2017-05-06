# [HSM](https://github.com/watercraft/hsm)
HSM is a [Go](http://www.golang.org) library for creating [Hierachical State Machines](https://en.wikipedia.org/wiki/UML_state_machine). Included in this package is an interactive calculator that follows this diagram:

![Calculator HSM](CalculatorHSM.png "Calculator Hierachical State Machine")
## Building
```
go install ...
```
## Usage
Your HSM is a struct that includes hsm.HSMBase and implements the instance interface.
```
import "github.com/watercraft/hsm"
type Calc struct {
        hsm.HSMBase
        ...
func (calc *Calc) Init(name string) {
        calc.HSMBase.Init(name, calc)
        ...
func (calc *Calc) Log() *logrus.Logger {
        ...
func (calc *Calc) LogTransition(tran *hsm.Transition, param interface{}) {
        ...
```
The Init() function is constructs the state machine adding states, transitions, and actions (see calc.go).  Log() returns the logrus.Logger for HSM to use if debugging is enabled and LogTransition() is a call out that is done after all actions and before the new state is applied. Actions are also member functions of the HSM instance. This action adds a decimal digit to operand1:
```
func (calc *Calc) ActionAppendOperand1(param interface{}) {
        calc.operand1 = calc.operand1*10 + intFromDigit(param)
}
```
You can then initialize your HSM and inject events.
```
        var myCalc Calc
        myCalc.Init("CALC")
        err = myCalc.Inject(CalcEventDigit, '5')
```
Here is an example run of the calculator.
```
$ calc
Welcome to HSM Calculator! Perform integer arithmetic with +, -, *, /, =, o=on, f=off, c=clear, q=quit
Calculator OFF
Calculator ON
     0
     2
    25
    25 +      0
    25 +      3
    28
    28 /      0
    28 /      4
     7
     0
Calculator OFF
Exit
```
