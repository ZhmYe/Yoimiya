package Record

import "time"

type Record struct {
	SplitTime time.Duration
	RunTime   time.Duration
}

func (r *Record) AddSplitTime(t time.Duration) {
	r.SplitTime += t
}
func (r *Record) AddRunTime(t time.Duration) {
	r.RunTime += t
}
func (r *Record) GetTime() (time.Duration, time.Duration) {
	return r.SplitTime, r.RunTime
}
func NewRecord() *Record {
	record := new(Record)
	record.SplitTime = time.Duration(0)
	record.RunTime = time.Duration(0)
	return record
}
func (r *Record) Clear() {
	r.SplitTime = time.Duration(0)
	r.RunTime = time.Duration(0)
}

var GlobalRecord = NewRecord()
