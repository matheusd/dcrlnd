package chainscan

import (
	"testing"
)

// TestTargetSlice tests whether the targetSlice struct behaves correctly.
func TestTargetSlice(t *testing.T) {

	t1 := &target{version: 1}
	t2 := &target{version: 2}
	t3 := &target{version: 3}

	ts := &targetSlice{}

	assertContains := func(want []*target) {
		t.Helper()
		got := *ts
		if len(want) != len(got) {
			t.Fatalf("different slice lengths; want=%v got=%v", want, got)
		}

		for i := range want {
			found := false
			for j := range got {
				if want[i] == got[j] {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("could not find wanted element %d. "+
					"want=%v got=%v", i, want, got)
			}

		}
	}

	ts.add(t1)
	ts.add(t2)
	ts.add(t3)
	assertContains([]*target{t1, t2, t3})

	ts.del(t2)
	assertContains([]*target{t1, t3})

	ts.del(t1)
	assertContains([]*target{t3})

	ts.del(t3)
	assertContains([]*target{})
}
