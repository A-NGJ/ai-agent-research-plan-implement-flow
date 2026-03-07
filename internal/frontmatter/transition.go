package frontmatter

import "fmt"

// validTransitions defines the allowed status state machine.
var validTransitions = map[string][]string{
	"draft":    {"active", "superseded"},
	"active":   {"complete", "superseded"},
	"complete": {"archived", "superseded"},
}

// Transition validates and applies a status change.
// Returns a ValidationError if the transition is invalid.
// Missing status is treated as "draft".
func Transition(doc *Document, newStatus string) error {
	current, _ := doc.Frontmatter["status"].(string)
	if current == "" {
		current = "draft"
	}

	allowed, ok := validTransitions[current]
	if !ok {
		return &ValidationError{
			Message: fmt.Sprintf("cannot transition from status %q", current),
		}
	}

	for _, a := range allowed {
		if a == newStatus {
			doc.Frontmatter["status"] = newStatus
			return nil
		}
	}

	return &ValidationError{
		Message: fmt.Sprintf("invalid status transition: %s → %s (allowed: %s → %v)", current, newStatus, current, allowed),
	}
}

// ValidationError represents a validation failure (exit code 2).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
