package rbac

import "strings"

type MemberRole string

const (
	MRoleRegular            MemberRole = "REGULAR"
	MRoleModerator          MemberRole = "MODERATOR"
	MRoleOwner              MemberRole = "OWNER"
	MRoleSpectator          MemberRole = "SPECTATOR"
	MRoleSpectatorModerator MemberRole = "SPECTATOR_MODERATOR"
	MRoleSpectatorOwner     MemberRole = "SPECTATOR_OWNER"
)

func (role MemberRole) IsValid() bool {
	switch role {
	case MRoleRegular, MRoleModerator, MRoleOwner,
		MRoleSpectator, MRoleSpectatorModerator, MRoleSpectatorOwner:
		return true
	}
	return false
}

func (role MemberRole) HasPermission(permission Permission) bool {
	return rolePermissions[role][permission]
}

func (role MemberRole) IsOwner() bool {
	return role == MRoleOwner || role == MRoleSpectatorOwner
}

func (role MemberRole) IsModerator() bool {
	switch role {
	case MRoleModerator, MRoleOwner, MRoleSpectatorModerator, MRoleSpectatorOwner:
		return true
	}
	return false
}

func (role MemberRole) IsParticipant() bool {
	switch role {
	case MRoleRegular, MRoleModerator, MRoleOwner:
		return true
	}
	return false
}

func (role MemberRole) IsMember() bool {
	return role.IsValid()
}

func ParseMemberRole(roleStr string) (MemberRole, bool) {
	switch strings.ToUpper(roleStr) {
	case "REGULAR":
		return MRoleRegular, true
	case "MODERATOR":
		return MRoleModerator, true
	case "OWNER":
		return MRoleOwner, true
	case "SPECTATOR":
		return MRoleSpectator, true
	case "SPECTATOR_MODERATOR":
		return MRoleSpectatorModerator, true
	case "SPECTATOR_OWNER":
		return MRoleSpectatorOwner, true
	default:
		return "", false
	}
}
