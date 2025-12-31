package timelatch

import (
	"testing"
	"time"
)

func TestTimeLatch(t *testing.T) {
	check := func(expected, actual bool) {
		t.Helper()
		if expected != actual {
			t.Errorf("unexpected trigger")
		}
	}
	{
		tl := NewAt(time.UnixMilli(1766990210010), time.UnixMilli(1766990210000))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210001)))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210005)))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210009)))
		check(true, tl.TriggeredAt(time.UnixMilli(1766990210010)))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210011)))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210012)))
	}
	{
		tl := NewAt(time.UnixMilli(1766990210010), time.UnixMilli(1766990210000))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210001)))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210005)))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210009)))
		check(true, tl.TriggeredAt(time.UnixMilli(1766990210011)))
		check(false, tl.TriggeredAt(time.UnixMilli(1766990210012)))
	}
	{
		now := time.UnixMilli(1766990210020)
		tl := NewAt(time.UnixMilli(1766990210010), now)
		count := 0
		for !tl.AddAt(1*time.Millisecond, now) {
			count++
		}
		if count != 10 {
			t.Errorf("expected: 10, got: %d", count)
		}
		if !tl.Time().Equal(time.UnixMilli(1766990210021)) {
			t.Errorf("expected: 1766990210021, got: %d", tl.Time().UnixMilli())
		}
	}
	{
		now := time.UnixMilli(1766990210020)
		tl := NewAt(time.UnixMilli(1766990210010), now)
		tl.AdvanceUntilFutureAt(1*time.Millisecond, now)
		if !tl.Time().Equal(time.UnixMilli(1766990210021)) {
			t.Errorf("expected: 1766990210021, got: %d", tl.Time().UnixMilli())
		}
	}
}
