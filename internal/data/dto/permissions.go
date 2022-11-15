package dto

import "github.com/kientink26/go-json-api/internal/validator"

// Define a Permissions slice, which we will use to will hold the permission codes for a single user.
type Permissions []string

var (
	MoviesWrite      = "movies:write"
	CommentsWrite    = "comments:write"
	UsersRead        = "users:read"
	PermissionsRead  = "permissions:read"
	PermissionsWrite = "permissions:write"
	PermissionList   = Permissions{CommentsWrite, MoviesWrite, UsersRead, PermissionsRead, PermissionsWrite}
)

func ValidatePermissions(v *validator.Validator, p Permissions) {
	v.Check(len(p) >= 1, "permissions", "must contain at least 1 code")
	v.Check(len(p) <= 5, "permissions", "must not contain more than 5 code")
	v.Check(validator.Unique(p), "permissions", "must not contain duplicate values")
	for _, code := range p {
		v.Check(validator.In(code, PermissionList...), "permission", "invalid permission code value")
	}
}

// Add a helper method to check whether the Permissions slice contains a specific
// permission code.
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}
