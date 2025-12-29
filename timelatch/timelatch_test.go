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
}
