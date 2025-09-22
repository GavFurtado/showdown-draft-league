package rbac

import "strings"

// PlayerRole represents the role of a player within a league.
type PlayerRole string

const (
	Owner     PlayerRole = "owner"
	Moderator PlayerRole = "moderator"
	Member    PlayerRole = "member"
)

// IsValid checks if the PlayerRole is one of the defined roles.
func (pr PlayerRole) IsValid() bool {
	switch pr {
	case Owner, Moderator, Member:
		return true
	}
	return false
}

// HasPermission checks if the role has a specific permission.
// This method will delegate to a centralized permission map.
func (pr PlayerRole) HasPermission(permission Permission) bool {
	// Permissions are defined in permissions.go
	// This function will be implemented once permissions.go is created
	return rolePermissions[pr][permission]
}

// IsOwner checks if the player role is Owner.
func (pr PlayerRole) IsOwner() bool {
	return pr == Owner
}

// IsModerator checks if the player role is Moderator (or Owner, as Owner implies Moderator).
func (pr PlayerRole) IsModerator() bool {
	return pr == Moderator || pr == Owner
}

// IsMember checks if the player role is Member (or Moderator/Owner, as they imply Member).
func (pr PlayerRole) IsMember() bool {
	return pr == Member || pr == Moderator || pr == Owner
}

// ParsePlayerRole parses a string into a PlayerRole, case-insensitively.
func ParsePlayerRole(roleStr string) (PlayerRole, bool) {
	switch strings.ToLower(roleStr) {
	case "owner":
		return Owner, true
	case "moderator":
		return Moderator, true
	case "member":
		return Member, true
	default:
		return "", false
	}
}
