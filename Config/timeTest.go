package Config

import (
	"fmt"
	"time"
)

type GlobalTimeTest struct {
	total    int
	timeUsed time.Duration
}

func (t *GlobalTimeTest) Add(ti time.Duration) {
	t.total++
	t.timeUsed += ti
}
func (t *GlobalTimeTest) Print() {
	fmt.Println(t.timeUsed, t.total)
}

var TimeTest = GlobalTimeTest{
	total:    0,
	timeUsed: 0,
}
