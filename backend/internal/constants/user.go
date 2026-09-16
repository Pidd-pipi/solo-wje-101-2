package constants

// UserRole enumerates platform roles.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// ValidRoles returns all accepted roles.
func ValidRoles() []string {
	return []string{RoleUser, RoleAdmin}
}
