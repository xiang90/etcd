package fileSystem

import (
	"testing"
)

// TestEventQueue tests a queue with capacity = 100
// Add 200 events into that queue, and test if the
// previous 100 events have been swapped out.
func TestEventQueue(t *testing.T) {

	eh := newEventHistory(100)

	// Add
	for i := 0; i < 200; i++ {
		e := newEvent(Create, "/foo", uint64(i), 0)
		eh.addEvent(e)
	}

	// Test
	j := 100
	i := eh.Queue.front
	n := eh.Queue.size
	for ; n > 0; n-- {
		e := eh.Queue.events[i]
		if e.Index != uint64(j) {
			t.Fatalf("queue error!")
		}
		j++
		i = (i + 1) % eh.Queue.capacity

	}

}

func TestScanHistory(t *testing.T) {
	eh := newEventHistory(100)

	// Add
	eh.addEvent(newEvent(Create, "/foo", 1, 0))
	eh.addEvent(newEvent(Create, "/foo/bar", 2, 0))
	eh.addEvent(newEvent(Create, "/foo/foo", 3, 0))
	eh.addEvent(newEvent(Create, "/foo/bar/bar", 4, 0))
	eh.addEvent(newEvent(Create, "/foo/foo/foo", 5, 0))

	e, err := eh.scan("/foo", 1)
	if err != nil || e.Index != 1 {
		t.Fatalf("scan error [/foo] [1] %v", e.Index)
	}

	e, err = eh.scan("/foo/bar", 1)

	if err != nil || e.Index != 2 {
		t.Fatalf("scan error [/foo/bar] [2] %v", e.Index)
	}

	e, err = eh.scan("/foo/bar", 3)

	if err != nil || e.Index != 4 {
		t.Fatalf("scan error [/foo/bar/bar] [4] %v", e.Index)
	}

	e, err = eh.scan("/foo/bar", 6)

	if e != nil {
		t.Fatalf("bad index shoud reuturn nil")
	}

}
