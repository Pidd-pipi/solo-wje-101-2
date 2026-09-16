package constants

// ProcessMethod enumerates coffee processing methods.
const (
	ProcessWashed   = "washed"
	ProcessNatural  = "natural"
	ProcessHoney    = "honey"
	ProcessAnaerobic = "anaerobic"
)

// ValidProcessMethods returns all accepted methods.
func ValidProcessMethods() []string {
	return []string{ProcessWashed, ProcessNatural, ProcessHoney, ProcessAnaerobic}
}

// IsValidProcessMethod reports whether a method is known.
func IsValidProcessMethod(s string) bool {
	for _, v := range ValidProcessMethods() {
		if v == s {
			return true
		}
	}
	return false
}
