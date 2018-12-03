package eggtimer

import (
	"fmt"
	"regexp"
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
	// Errors, if any, found during segment identification.
	Error error
}

// A SegmentDefinition processes lines and determines if they
// mark the start or end of a given tagged segment.
type SegmentDefinition interface {
	// TypeName returns the string name for this type of segment.
	TypeName() string
	// IsStart determines if the line is a start for this kind segment,
	// and returns a non-empty string for the name of this instance of this
	// type of segment.  An empty return value indicates it is not a start.
	IsStart(line string) string
	// IsFinish determines if the line is a finish for this kind segment,
	// and returns a non-empty string for the name of this instance of this
	// type of segment.  An empty return value indicates it is not a finish.
	IsFinish(line string) string
}

// A RegexpDef is a regexp-based SegmentDefinition, which uses regular
// expressions to identify start and finish Events.
type RegexpDef struct {
	name string
	reStart *regexp.Regexp
	reFinish *regexp.Regexp
}

// NewRegexpDef creates a new RegexpDef given a type name, and start and finish
// regexps.  Each regexp's first sub-expression which identifies the tag
// of a given segment instance.
func NewRegexpDef(typeName, startExpr, finishExpr string) *RegexpDef {
	return &RegexpDef {
		name: typeName,
		reStart: regexp.MustCompile(startExpr),
		reFinish: regexp.MustCompile(finishExpr),
	}
}

// TypeName returns the string name for this segment type.
func (d *RegexpDef) TypeName() string {
	return d.name
}

// IsStart determines if the line is a start for this kind segment,
// and returns a non-empty string for the name of this instance of this
// type of segment.  An empty return value indicates it is not a start.
func (d *RegexpDef) IsStart(line string) string {
	matches := d.reStart.FindStringSubmatch(line)
	if matches == nil {
		return ""
	}
	return matches[1]
}

// IsFinish determines if the line is a finish for this kind segment,
// and returns a non-empty string for the name of this instance of this
// type of segment.  An empty return value indicates it is not a finish.
func (d *RegexpDef) IsFinish(line string) string {
	matches := d.reFinish.FindStringSubmatch(line)	
	if matches == nil {
		return ""
	}
	return matches[1]
}

// A Segmenter processes Events, using a set of SegmentDefinitions to
// identify segments that have start and finish events within the processed set.
type Segmenter struct {
	segTypes []SegmentDefinition
}

// AddDefinitions adds a SegmentDefinition to a Segmenter.
func (this *Segmenter) AddDefinition(d SegmentDefinition) {
	this.segTypes = append(this.segTypes, d)
}

// Collect processes a channel of Events and, using its SegmentDefinitions, compiles
// a table mapping activity names to a map of segment names to Segment structs.
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
