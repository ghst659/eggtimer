package eggtimer

import (
	"fmt"
	"sync"
	"time"
)

// A SegmentDefinition processes lines and determines if they
// mark the start or end of a given tagged segment.
type SegmentDefinition interface {
	Name() string
	IsStart(line string) string
	IsStop(line string) string
}

type Segmenter struct {
	segTypes []SegmentDefinition
}

type TagSegment struct {
	Tag string
	Opening time.Duration
	Closing time.Duration
}

func (this Segmenter) segment(events <-chan Event) (result map[string]TagSegment) {
	var wg sync.WaitGroup
	wg.Add(1)
	for e := range events {
		for i, d := range this.segTypes {
			if tag := d.IsStart(e.What); tag != "" {
				fmt.Println(i)
			}
		}
	}
	return
}
