package sessions

import "strings"

// SessionClassification represents the 3-state classification of a session.
type SessionClassification string

const (
	ClassFinished    SessionClassification = "finished"
	ClassNeedsResume SessionClassification = "needs_resume"
	ClassObsolete    SessionClassification = "obsolete"
)

// ClassificationEmoji returns the display emoji for the classification.
func (c SessionClassification) ClassificationEmoji() string {
	switch c {
	case ClassFinished:
		return "🟢"
	case ClassNeedsResume:
		return "⏸️"
	case ClassObsolete:
		return "🪦"
	default:
		return "❓"
	}
}

// ClassificationLabel returns the display label for the classification.
func (c SessionClassification) ClassificationLabel() string {
	switch c {
	case ClassFinished:
		return "FINISHED"
	case ClassNeedsResume:
		return "NEEDS_RESUME"
	case ClassObsolete:
		return "OBSOLETE"
	default:
		return strings.ToUpper(string(c))
	}
}

// String implements the Stringer interface.
func (c SessionClassification) String() string {
	return c.ClassificationEmoji() + " " + c.ClassificationLabel()
}
