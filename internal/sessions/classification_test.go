package sessions

import "testing"

func TestClassificationConstants(t *testing.T) {
	tests := []struct {
		name     string
		class    SessionClassification
		wantVal  string
	}{
		{"ClassFinished value", ClassFinished, "finished"},
		{"ClassNeedsResume value", ClassNeedsResume, "needs_resume"},
		{"ClassObsolete value", ClassObsolete, "obsolete"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.class) != tt.wantVal {
				t.Errorf("got %q, want %q", string(tt.class), tt.wantVal)
			}
		})
	}
}

func TestClassificationEmoji(t *testing.T) {
	tests := []struct {
		name  string
		class SessionClassification
		want  string
	}{
		{"finished emoji", ClassFinished, "🟢"},
		{"needs_resume emoji", ClassNeedsResume, "🔴"},
		{"obsolete emoji", ClassObsolete, "⚠️"},
		{"unknown emoji", SessionClassification("unknown"), "❓"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.class.ClassificationEmoji()
			if got != tt.want {
				t.Errorf("ClassificationEmoji() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassificationLabel(t *testing.T) {
	tests := []struct {
		name  string
		class SessionClassification
		want  string
	}{
		{"finished label", ClassFinished, "FINISHED"},
		{"needs_resume label", ClassNeedsResume, "NEEDS_RESUME"},
		{"obsolete label", ClassObsolete, "OBSOLETE"},
		{"unknown label", SessionClassification("unknown"), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.class.ClassificationLabel()
			if got != tt.want {
				t.Errorf("ClassificationLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassificationString(t *testing.T) {
	tests := []struct {
		name  string
		class SessionClassification
		want  string
	}{
		{"finished string", ClassFinished, "🟢 FINISHED"},
		{"needs_resume string", ClassNeedsResume, "🔴 NEEDS_RESUME"},
		{"obsolete string", ClassObsolete, "⚠️ OBSOLETE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.class.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
