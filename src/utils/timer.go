// timer
package utils

import (
	"fmt"
	"time"
)

type Timer struct {
	Start *time.Time
	List  []*TimePoint
}

type TimePoint struct {
	Timer *time.Time
	Msg   string
}

func NewTimer() *Timer {
	now := time.Now()
	timer := &Timer{
		Start: &now,
	}
	return timer
}

func (t *Timer) Add(msg string) {
	now := time.Now()
	timePoint := &TimePoint{&now, msg}
	t.List = append(t.List, timePoint)
}

func (t *Timer) Out() {
	fmt.Println("timer Out")
	for _, v := range t.List {
		fmt.Printf("time:%v, msg:%v \n", v.Timer.Format("2006-01-02 15:04:05"), v.Msg)
	}
}
