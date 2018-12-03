package eggtimer

import (
	"testing"
	"time"
)

func sendData(data []Event, inputs chan<- Event) {
	defer close(inputs)
	for _, d := range data {
		inputs <- d
	}
}

func duration(text string) time.Duration {
	value, err := time.ParseDuration(text)
	if err != nil {
		panic(err.Error())
	}
	return value
}

func TestCollect(t *testing.T) {
	var segmenter Segmenter
	segmenter.AddDefinition(NewRegexpDef("RType", `^Start\s+(\w+)`, `^Finish\s+(\w+)`))
	inputs := make(chan Event)
	data := []Event {
		{
			When: duration("0s"),
			What: "Ignore this string",
		},
		{
			When: duration("1s"),
			What: "Start x",
		},
		{
			When: duration("9s"),
			What: "Finish y",
		},
		{
			When: duration("3s"),
			What: "Start  y",
		},
		{
			When: duration("2s"),
			What: "Finish x",
		},
	}
	go sendData(data, inputs)
	table, err := segmenter.Collect(inputs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(table) != 2 {
		t.Errorf("bad table length: %d", len(table))
	}
	wants := map[string]*Segment {
		"RType:x": &Segment{
			Name: "RType:x",
			Start: duration("1s"),
			Finish: duration("2s"),
		},
		"RType:y": &Segment{
			Name: "RType:y",
			Start: duration("3s"),
			Finish: duration("9s"),
		},
	}
	for key, want := range wants {
		got, ok := table[key]
		if !ok {
			t.Errorf("missing key: %s", key)
		}
		if got.Name != want.Name {
			t.Errorf("%s name mismatch: %s vs %s", key, got.Name, want.Name)
		}
		if got.Start != want.Start {
			t.Errorf("%s start mismatch: %v vs %v", key, got.Start, want.Start)
		}
		if got.Finish != want.Finish {
			t.Errorf("%s finish mismatch: %v vs %v", key, got.Finish, want.Finish)
		}
	}
}
