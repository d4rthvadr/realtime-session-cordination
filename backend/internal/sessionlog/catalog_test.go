package sessionlog

import "testing"

func TestRenderMessage(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		input     MessageInput
		want      string
	}{
		{
			name:      "session start",
			eventType: SessionStarted,
			input:     MessageInput{SessionTitle: "Town Hall"},
			want:      "Started session \"Town Hall\"",
		},
		{
			name:      "session adjust",
			eventType: SessionTimeAdjusted,
			input:     MessageInput{DeltaSeconds: 60},
			want:      "Adjusted session time by +60s",
		},
		{
			name:      "program item reorder with count",
			eventType: ProgramItemsReordered,
			input:     MessageInput{ReorderedItemCount: 4},
			want:      "Reordered 4 program items",
		},
		{
			name:      "program item adjust",
			eventType: ProgramItemTimeAdjusted,
			input: MessageInput{
				ProgramItemTitle: "Q&A",
				DeltaSeconds:     -30,
			},
			want: "Adjusted program item \"Q&A\" by -30s",
		},
		{
			name:      "cascade pause",
			eventType: CascadeProgramItemPausedBySession,
			input:     MessageInput{ProgramItemTitle: "Keynote"},
			want:      "Auto-paused program item \"Keynote\" because session was paused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderMessage(tt.eventType, tt.input)
			if got != tt.want {
				t.Fatalf("RenderMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderMessageFallback(t *testing.T) {
	unknown := EventType("CUSTOM_EVENT")
	got := RenderMessage(unknown, MessageInput{})
	if got != "Recorded event CUSTOM_EVENT" {
		t.Fatalf("RenderMessage() fallback = %q", got)
	}
}
