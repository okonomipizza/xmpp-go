package protocol

import "testing"

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateInitial, "Initial"},
		{StateAwaitServerStream, "AwaitServerStream"},
		{StateAwaitFeatures, "AwaitFeatures"},
		{StateNegotiating, "Negotiating"},
		{StateAwaitTLSProceed, "AwaitTLSProceed"},
		{StateAwaitSASLOutcome, "AwaitSASLOutcome"},
		{StateAwaitBind, "AwaitBind"},
		{StateReady, "Ready"},
		{StateClosing, "Closing"},
		{StateClosed, "Closed"},
		{State(99), "State(99)"},
	}
	for _, tc := range tests {
		if got := tc.state.String(); got != tc.want {
			t.Fatalf("state %d: got %q want %q", tc.state, got, tc.want)
		}
	}
}
