package archive

import "testing"

func TestExtractTimeline(t *testing.T) {
	entry := &Entry{Data: []byte(`<?xml version="1.0" encoding="utf-8"?>
<Session SID="42" BitFlags="19">
  <SessionTimers ClientBeginRequest="2026-04-08T15:42:15.1413311+08:00" ClientDoneResponse="2026-04-08T15:42:15.8764964+08:00" />
</Session>`)}
	info := ExtractTimeline(entry, 42)
	if !info.HasTimeline {
		t.Fatal("expected timeline to be present")
	}
	if info.Begin == "" || info.End == "" {
		t.Fatalf("unexpected timeline values: %#v", info)
	}
}

func TestSortTimeline(t *testing.T) {
	infos := []TimelineInfo{
		{SessionID: 2, Begin: "2026-04-08T15:42:15.1413311+08:00", HasTimeline: true},
		{SessionID: 1, Begin: "2026-04-08T15:41:58.4766015+08:00", HasTimeline: true},
		{SessionID: 3, Begin: "", HasTimeline: false},
	}
	sorted := SortTimeline(infos)
	if sorted[0].SessionID != 1 || sorted[1].SessionID != 2 || sorted[2].SessionID != 3 {
		t.Fatalf("unexpected sort order: %#v", sorted)
	}
}
