package example_test

import (
	"testing"

	e "github.com/ckbaldy/hsm/example"
	"github.com/smartystreets/goconvey/convey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStateTransition(t *testing.T) {

	sm := e.NewHSM("tstHSM")

	Convey("CASE: Run All Transitions", t, func() {
		Convey("1. Intial Transition\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\tOn; Transition from Inital Pseudo state into Top State")
			err := sm.On()
			Print("\t  Final State: ", sm.CurrentState)
			So(err, convey.ShouldBeNil)
			So(sm.CurrentState, convey.ShouldEqual, "s11")
		})

		Convey("2. Self Transition\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\t->EventA")
			sm.Inject(e.EventA, nil)
			Print("\t  Final State: ", sm.CurrentState)
		})

		Convey("3. Direct Transition to Different State Tree\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\t->EventE")
			sm.Inject(e.EventE, nil)
			Print("\t  Final State: ", sm.CurrentState)
		})

		Convey("4. Direct Transition to Child in Same Tree\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\t->EventE")
			sm.Inject(e.EventE, nil)
			Print("\t  Final State: ", sm.CurrentState)
		})

		Convey("5. Unhandled Event\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\t->EventA")
			sm.Inject(e.EventA, nil)
			Println("\t  EventA not handled since it is not in the tree")
			Print("\t  Final State: ", sm.CurrentState)
		})

		Convey("6. Guarded Transition, Guard Returns False\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\t->EventH")
			// Additionallly verifies passing/processing single param through
			// transition action.
			sm.Inject(e.EventH, !sm.Foo)
			Print("\t  Final State: ", sm.CurrentState)
		})

		Convey("7. Guarded Transition, Guard Returns True\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\t->EventH")
			sm.Inject(e.EventH, !sm.Foo)
			Println("\t  EventH guarded")
			Print("\t  Final State: ", sm.CurrentState)
		})

		Convey("8. Transition to Target with Default Child State\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\t->EventC")
			sm.Inject(e.EventC, nil)
			Print("\t  Final State: ", sm.CurrentState)
		})

		Convey("9. Internal Transition\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\t->EventH")
			sm.Inject(e.EventH, sm.Foo)
			Print("\t  Foo ", sm.Foo)
			So(sm.Foo, convey.ShouldBeFalse)
			Print("\n\t  Final State: ", sm.CurrentState)
		})

		Convey("10. Exit Transition to Top State (off)\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			err := sm.Off()
			Print("\t  Final State: ", sm.CurrentState)
			So(err, convey.ShouldBeNil)
			So(sm.CurrentState, convey.ShouldEqual, "TopState")
		})

		Convey("11. Intial Transition, turn back on\n", func() {
			Println("\tCurrent State: ", sm.CurrentState)
			Println("\tOn; Transition from Inital Pseudo state into Top State")
			err := sm.On()
			Print("\t  Final State: ", sm.CurrentState)
			So(err, convey.ShouldBeNil)
			So(sm.CurrentState, convey.ShouldEqual, "s11")
		})

	})
}
