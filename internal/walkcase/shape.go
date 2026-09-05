package walkcase

import "github.com/blisspixel/fartapp/internal/strictjson"

// Go's JSON decoder also accepts case-insensitive aliases and silently accepts
// null for several concrete types. Reject both before typed interpretation.
// Syntax, duplicate-member, byte, and depth limits have already been checked.
func inspectDocumentShape(data []byte) *Diagnostic {
	if issue := strictjson.InspectShape[requestDocument](data); issue != nil {
		return schema(issue.Path, string(issue.Kind))
	}
	return nil
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}
