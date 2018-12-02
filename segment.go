package eggtimer

import (
	"fmt"
	"time"
)

// A Segment represents the start and stop time of an activity.
type Segment struct {
	// The name of the segment
	Name string
	// The start of the activity, relative to the overall workflow start.
	Start time.Duration
	// The finish of the activity, relative to teh overall workflow start.
	Finish time.Duration
	// Errors, if any
	Error error
}

// A SegmentDefinition processes lines and determines if they
// mark the start or end of a given tagged segment.
type SegmentDefinition interface {
	TypeName() string
	IsStart(line string) string
	IsFinish(line string) string
}

// A Segmenter is 
type Segmenter struct {
	segTypes []SegmentDefinition
}

// AddDefinitions adds a SegmentDefinition to a Segmenter.
func (this *Segmenter) AddDefinition(d SegmentDefinition) {
	this.segTypes = append(this.segTypes, d)
}

// Segment collects Events and, using its SegmentDefinitions, compiles
// a table mapping activity names to a Segment struct.
func (this Segmenter) Collect(events <-chan Event) (activity map[string]*Segment, err error) {
	activity = make(map[string]*Segment)
	for e := range events {
		err = e.Error
		if err != nil {
			return
		}
		for _, d := range this.segTypes {
			if tag := d.IsStart(e.What); tag != "" {
				name := fmt.Sprintf("%s:%s", d.TypeName(), tag)
				if segment, ok := activity[name]; !ok {
					activity[name] = &Segment{
						Name: name,
						Start: e.When,
					}
				} else {
					segment.Start = e.When
				}
			} else if tag := d.IsFinish(e.What); tag != "" {
				name := fmt.Sprintf("%s:%s", d.TypeName(), tag)
				if segment, ok := activity[name]; !ok {
					activity[name] = &Segment{
						Name: name,
						Finish:  e.When,
					}
				} else {
					segment.Finish = e.When
				}
			} else {
				// ignore events that are not entries or exits
			}
		}
	}
	return
}
