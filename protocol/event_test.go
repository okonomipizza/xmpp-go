package protocol

import "testing"

func TestEventMarkers(t *testing.T) {
	(&StreamOpenedEvent{}).isEvent()
	(&StreamFeaturesEvent{}).isEvent()
	(&StartTLSProceedEvent{}).isEvent()
	(&StartTLSFailureEvent{}).isEvent()
	(&SASLChallengeEvent{}).isEvent()
	(&SASLSuccessEvent{}).isEvent()
	(&SASLFailureEvent{}).isEvent()
	(&BindSuccessEvent{}).isEvent()
	(&BindFailureEvent{}).isEvent()
	(&MessageEvent{}).isEvent()
	(&PresenceEvent{}).isEvent()
	(&StanzaEvent{}).isEvent()
	(&ElementEvent{}).isEvent()
	(&StreamClosedEvent{}).isEvent()
	(&StreamErrorEvent{}).isEvent()
}
