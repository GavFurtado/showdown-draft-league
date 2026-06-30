package rbac

import "strings"

type MemberRole string

const (
	MRoleOwner     MemberRole = "OWNER"
	MRoleModerator MemberRole = "MODERATOR"
	MRoleMember    MemberRole = "MEMBER"
)

func (role MemberRole) IsValid() bool {
	switch role {
	case MRoleOwner, MRoleModerator, MRoleMember:
		return true
	}
	return false
}

func (role MemberRole) HasPermission(permission Permission) bool {
	return rolePermissions[role][permission]
}

func (role MemberRole) IsOwner() bool {
	return role == MRoleOwner
}

func (role MemberRole) IsModerator() bool {
	return role == MRoleModerator || role == MRoleOwner
}

func (role MemberRole) IsMember() bool {
	return role == MRoleMember || role == MRoleModerator || role == MRoleOwner
}

func ParseMemberRole(roleStr string) (MemberRole, bool) {
	switch strings.ToLower(roleStr) {
	case "owner":
		return MRoleOwner, true
	case "moderator":
		return MRoleModerator, true
	case "member":
		return MRoleMember, true
	default:
		return "", false
	}
}
