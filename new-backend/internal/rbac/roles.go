package rbac

// PlayerRole defines the role of a player within a specific league.
type PlayerRole string

const (
	PlayerRoleOwner     PlayerRole = "owner"
	PlayerRoleModerator PlayerRole = "moderator"
	PlayerRoleMember    PlayerRole = "member"
)

// IsValid checks if the PlayerRole is one of the defined valid roles.
func (pr PlayerRole) IsValid() bool {
	switch pr {
	case PlayerRoleOwner, PlayerRoleModerator, PlayerRoleMember:
		return true
	}
	return false
}

// HasPermission checks if the current role has a specific permission.
// This method relies on the rolePermissions map defined in permissions.go
func (pr PlayerRole) HasPermission(permission string) bool {
	if permissions, ok := rolePermissions[pr]; ok {
		return permissions[permission]
	}
	return false
}
