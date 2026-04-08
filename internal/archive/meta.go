package archive

import (
	"encoding/xml"
	"sort"
)

type sessionXML struct {
	Timers sessionTimers `xml:"SessionTimers"`
}

type sessionTimers struct {
	ClientBeginRequest string `xml:"ClientBeginRequest,attr"`
	ClientDoneResponse string `xml:"ClientDoneResponse,attr"`
}

type TimelineInfo struct {
	SessionID   int
	Begin       string
	End         string
	HasTimeline bool
}

func ExtractTimeline(meta *Entry, sessionID int) TimelineInfo {
	info := TimelineInfo{SessionID: sessionID}
	if meta == nil || len(meta.Data) == 0 {
		return info
	}
	var parsed sessionXML
	if err := xml.Unmarshal(meta.Data, &parsed); err != nil {
		return info
	}
	info.Begin = parsed.Timers.ClientBeginRequest
	info.End = parsed.Timers.ClientDoneResponse
	info.HasTimeline = info.Begin != "" || info.End != ""
	return info
}

func SortTimeline(infos []TimelineInfo) []TimelineInfo {
	out := append([]TimelineInfo(nil), infos...)
	sort.SliceStable(out, func(i, j int) bool {
		ai := out[i]
		aj := out[j]
		if ai.Begin == aj.Begin {
			return ai.SessionID < aj.SessionID
		}
		if ai.Begin == "" {
			return false
		}
		if aj.Begin == "" {
			return true
		}
		return ai.Begin < aj.Begin
	})
	return out
}
