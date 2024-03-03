package ring

import (
	"testing"
	"time"
)

const MSG_UNEXPECTED_CLOSE string = "unexpected channel close"
const MSG_WRONG_MULTIPLE string = "output: idx=%v expected %v; received %v"
const MSG_WRONG_SINGLE string = "output: expected %v; received %v"
const MSG_DATA_UNEXPECTED string = "unexpected data before any send: %v"

func TestSingle(t *testing.T) {
	t.Run("overfill", func(t *testing.T) {
		const VAL_PLAN int = 42
		chTx := make(chan int)
		chRx := make(chan int)
		ringBufferLoop := MakeRingBufferLoop[int](1)
		go ringBufferLoop(chTx, chRx)

		chTx <- 1
		chTx <- VAL_PLAN
		valFact, ok := <-chRx

		if !ok {
			t.Error(MSG_UNEXPECTED_CLOSE)
		}
		if VAL_PLAN != valFact {
			t.Errorf(MSG_WRONG_SINGLE, VAL_PLAN, valFact)
		}
	})

	t.Run("starvation", func(t *testing.T) {
		const VAL_PLAN int = 42
		chTx := make(chan int)
		chRx := make(chan int)
		ringBufferLoop := MakeRingBufferLoop[int](1)
		go ringBufferLoop(chTx, chRx)

		var valFact int
		var ok bool

		select {
		case valFact, ok = <-chRx:
			if !ok {
				t.Error(MSG_UNEXPECTED_CLOSE)
			} else {
				t.Errorf(MSG_DATA_UNEXPECTED, valFact)
			}
		default:
		}

		go func() {
			time.Sleep(time.Millisecond * 50)
			chTx <- VAL_PLAN
		}()
		valFact, ok = <-chRx

		if !ok {
			t.Error(MSG_UNEXPECTED_CLOSE)
		}
		if VAL_PLAN != valFact {
			t.Errorf(MSG_WRONG_SINGLE, VAL_PLAN, valFact)
		}
	})
}

func TestMultiple(t *testing.T) {
	t.Run("overfill", func(t *testing.T) {
		slPlan := []int{7, 8, 9, 10, 11, 12}
		chTx := make(chan int)
		chRx := make(chan int)
		ringBufferLoop := MakeRingBufferLoop[int](6)
		go func() {
			defer close(chRx)
			ringBufferLoop(chTx, chRx)
		}()

		for _, val := range []int{1, 2, 3, 4, 5, 6} {
			chTx <- val
		}
		for _, val := range slPlan {
			chTx <- val
		}
		close(chTx)

		idx := 0
		for valFact := range chRx {
			if slPlan[idx] != valFact {
				t.Errorf(MSG_WRONG_MULTIPLE, idx, slPlan[idx], valFact)
			}
			idx += 1
		}
		if idx < len(slPlan) {
			t.Error(MSG_UNEXPECTED_CLOSE)
		}

	})

	t.Run("starvation", func(t *testing.T) {
		slPlan := []int{7, 8, 9, 10, 11, 12}
		slPlanCopy := make([]int, len(slPlan))
		copy(slPlanCopy, slPlan)

		chTx := make(chan int)
		chRx := make(chan int)
		ringBufferLoop := MakeRingBufferLoop[int](6)
		go func() {
			defer close(chRx)
			ringBufferLoop(chTx, chRx)
		}()

		var valFact int
		var ok bool

		slPrefix := []int{1, 2, 3, 4, 5, 6}
		for _, val := range slPrefix {
			chTx <- val
		}
		for idx := 0; idx < len(slPrefix); idx++ {
			valFact, ok = <-chRx
			if !ok {
				t.Error(MSG_UNEXPECTED_CLOSE)
			}
			if slPrefix[idx] != valFact {
				t.Errorf(MSG_WRONG_MULTIPLE, idx, slPlan[idx], valFact)
			}
		}

		select {
		case valFact, ok = <-chRx:
			if !ok {
				t.Error(MSG_UNEXPECTED_CLOSE)
			} else {
				t.Errorf(MSG_DATA_UNEXPECTED, valFact)
			}
		default:
		}

		go func() {
			time.Sleep(time.Millisecond * 50)
			for _, val := range slPlanCopy {
				chTx <- val
			}
			close(chTx)
		}()
		idx := 0
		for valFact := range chRx {
			if slPlan[idx] != valFact {
				t.Errorf(MSG_WRONG_MULTIPLE, idx, slPlan[idx], valFact)
			}
			idx += 1
		}
		if idx < len(slPlan) {
			t.Error(MSG_UNEXPECTED_CLOSE)
		}
	})
}
