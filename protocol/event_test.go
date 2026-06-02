package protocol

import "testing"

func TestEventMarkers(t *testing.T) {
	(&StreamOpenedEvent{}).isEvent()
	(&StreamFeaturesEvent{}).isEvent()
	(&StanzaEvent{}).isEvent()
	(&ElementEvent{}).isEvent()
	(&StreamClosedEvent{}).isEvent()
	(&StreamErrorEvent{}).isEvent()
}
