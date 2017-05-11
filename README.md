# [HSM](https://github.com/watercraft/hsm)
HSM is a [Go](http://www.golang.org) library for creating [Hierarchical State Machines](https://en.wikipedia.org/wiki/UML_state_machine). Included in this package is an interactive calculator that follows this diagram:

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
func (calc *Calc) Init() {
        calc.HSMBase.Init(calc)
        ...
func (calc *Calc) Name() string {
        ...
func (calc *Calc) Log() *logrus.Logger {
        ...
func (calc *Calc) LogTransition(from hsm.State, tran *hsm.Transition, param interface{}) {
        ...
```
The *Init()* function constructs the state machine adding states, transitions, and actions see [calc.go](calc/calc.go).  *Name()* returns a name for the HSM instance, *Log()* returns the logrus.Logger for HSM to use and *LogTransition()* is called after each state transition. Actions are also member functions of the HSM instance. For example, this action adds a decimal digit to the member *operand1*:
```
func (calc *Calc) ActionAppendOperand1(param interface{}) error {
        calc.operand1 = calc.operand1*10 + intFromDigit(param)
	return nil
}
```
Exit and entry actions can also be specified. All exit actions from the current state to the ancestor with a matching transition are called in ascending order followed by entry actions for the new state and the transition action.  Entry actions for parents of the new state are not called. Any failure returned from an action will abort the transition.  Our calculator defines an entry action for the *On* state that clears the calculator for both the *OnButton* and *Clear* events:
```
	on.AddEntryActions(calc.ActionClear)
```
Once your machine is defined, you can then initialize your HSM and inject events.
```
        var myCalc Calc
        myCalc.Init("CALC")
        err = myCalc.Inject(CalcEventDigit, '5')
```
Here is an example run of the calculator.
```
$ calc
Welcome to the HSM Calculator!
Perform integer arithmetic with +, -, *, /, =, o=on, f=off, c=clear, q=quit
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
