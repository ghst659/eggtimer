package eggtimer

import (
	"io/ioutil"
	"os/exec"
	"fmt"
	"strings"
	"testing"
	"time"
)

type mockClock struct {
	index int
	points []time.Time
}

func (m *mockClock) Now() time.Time {
	if m.index >= len(m.points) {
		panic(fmt.Sprintf("mockClock index out of range: %q", m))
	}
	defer func() {m.index++}()
	return m.points[m.index]
}

func TestRun(t *testing.T) {
	data := []struct {
		t time.Time
		s string
	}{
		{
			t: time.Date(2018, time.December, 1, 0, 0, 10, 0, time.UTC),
			s: "Line 1\n",
		},
		{
			t: time.Date(2018, time.December, 1, 0, 0, 20, 0, time.UTC),
			s: "Line 2\n",
		},
		{
			t: time.Date(2018, time.December, 1, 0, 0, 30, 0, time.UTC),
			s: "Line 3\n",
		},
	}
	m := &mockClock{}
	var buf strings.Builder
	m.points = append(m.points, time.Date(2018, time.December, 1, 0, 0, 0, 0, time.UTC))
	for i := 0; i < len(data); i++ {
		m.points = append(m.points, data[i].t)
		buf.WriteString(data[i].s)
	}
	dataPath := "/tmp/data.txt"
	ioutil.WriteFile(dataPath, []byte(buf.String()), 0777)
	cmd := exec.Command("/bin/cat", dataPath)
	rut := NewRunner(m)
	events := make(chan Event)
	go rut.Run(cmd, events)
	i := 1
	for e := range events {
		t.Logf("event: %q", e)
		wantTime, err := time.ParseDuration(fmt.Sprintf("%ds", i * 10))
		if err != nil {
			t.Errorf("setup mismatch: %q", err)
		}
		wantText := fmt.Sprintf("Line %d", i)
		if e.When != wantTime || e.What != wantText {
			t.Errorf("mismatch: %q vs %q, %q vs %q", e.When, wantTime, e.What, wantText)
		}
		i++
	}	
}
