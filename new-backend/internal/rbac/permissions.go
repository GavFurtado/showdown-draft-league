package rbac

import "slices"

// Permission represents a granular action:resource permission.
type Permission string

// Define permission constants
const (
	// League Permissions
	PermissionReadLeague   Permission = "read:league"
	PermissionUpdateLeague Permission = "update:league"
	PermissionDeleteLeague Permission = "delete:league"
	// This PermissionCreateLeague is redundant; Permissions are assigned to member (player/spectator) and not to Users.
	// and creating a League is something the User does and not a member.
	// A League is first created then a member. Keeping this in to understand the roles
	PermissionCreateLeague Permission = "create:league"

	// PoolEntry Permissions
	PermissionCreatePoolEntry Permission = "create:pool_entry"
	PermissionReadPoolEntry   Permission = "read:pool_entry"
	PermissionUpdatePoolEntry Permission = "update:pool_entry"
	PermissionDeletePoolEntry Permission = "delete:pool_entry"

	// Player Permissions
	PermissionCreatePlayer      Permission = "create:player"
	PermissionReadPlayer        Permission = "read:player"
	PermissionUpdatePlayer      Permission = "update:player"
	PermissionUpdateOwnPlayer   Permission = "update:own_player"
	PermissionUpdatePlayerScore Permission = "update:member_score"
	PermissionDeletePlayer      Permission = "delete:player"

	// Spectator Permissions
	PermissionCreateSpectator    Permission = "create:spectator"
	PermissionReadSpectator      Permission = "read:spectator"
	PermissionUpdateSpectator    Permission = "update:spectator"
	PermissionUpdateOwnSpectator Permission = "update:own_spectator"
	PermissionDeleteSpectator    Permission = "delete:spectator"

	// Role Permissions
	PermissionUpdateRole Permission = "update:role"

	// DraftPick Permissions
	PermissionReadDraftPick   Permission = "read:draft_pick"
	PermissionCreateDraftPick Permission = "create:draft_pick"

	// Claim Permissions
	PermissionReadClaim   Permission = "read:claim"
	PermissionCreateClaim Permission = "create:claim"
	PermissionUpdateClaim Permission = "update:claim"

	// Draft Permissions
	PermissionCreateDraft Permission = "create:draft"
	PermissionReadDraft   Permission = "read:draft"
	PermissionUpdateDraft Permission = "update:draft"
	PermissionDeleteDraft Permission = "delete:draft"

	// Transfer Period Permissions
	PermissionStartTransferPeriod Permission = "start:transfer_period"
	PermissionEndTransferPeriod   Permission = "end:transfer_period"

	// LeaguePokemon Permissions
	PermissionCreateLeaguePokemon Permission = "create:league_pokemon"
	PermissionReadLeaguePokemon   Permission = "read:league_pokemon"
	PermissionUpdateLeaguePokemon Permission = "update:league_pokemon"
	PermissionDeleteLeaguePokemon Permission = "delete:league_pokemon"

	// DraftedPokemon Permissions
	PermissionReleaseDraftedPokemon Permission = "release:drafted_pokemon"
	PermissionCreateDraftedPokemon  Permission = "create:drafted_pokemon"
	PermissionReadDraftedPokemon    Permission = "read:drafted_pokemon"
	PermissionUpdateDraftedPokemon  Permission = "update:drafted_pokemon"
	PermissionDeleteDraftedPokemon  Permission = "delete:drafted_pokemon"

	// Game Permissions
	PermissionCreateGame   Permission = "create:game"
	PermissionReadGame     Permission = "read:game"
	PermissionUpdateGame   Permission = "update:game"
	PermissionDeleteGame   Permission = "delete:game"
	PermissionReportGame   Permission = "report:game"
	PermissionFinalizeGame Permission = "finalize:game"

	// User Permissions
	PermissionReadUser   Permission = "read:user"
	PermissionUpdateUser Permission = "update:user"
	PermissionDeleteUser Permission = "delete:user"

	// PlayerRoster Permissions
	PermissionCreatePlayerRoster Permission = "create:player_roster"
	PermissionReadPlayerRoster   Permission = "read:player_roster"
	PermissionUpdatePlayerRoster Permission = "update:player_roster"
	PermissionDeletePlayerRoster Permission = "delete:player_roster"

	// PokemonSpecies Permissions
	PermissionReadPokemonSpecies Permission = "read:pokemon_species"
)

// participationPermissions are Player-only: actions tied to actively playing
// (drafting, claiming, picking up/set Pokémon, editing your own player profile).
// Spectator roles are derived from their player counterparts by removing exactly
// these, so a spectator's role can never grant participation.
var participationPermissions = []Permission{
	PermissionCreateDraftPick,
	PermissionCreateClaim,
	PermissionUpdateClaim,
	PermissionCreateDraftedPokemon,
	PermissionReleaseDraftedPokemon,
	PermissionUpdateOwnPlayer,
}

// readPermissions are what a Viewer (expiring read-only spectator) may do.
var readPermissions = []Permission{
	PermissionReadLeague,
	PermissionReadPoolEntry,
	PermissionReadPlayer,
	PermissionReadSpectator,
	PermissionReadDraftPick,
	PermissionReadClaim,
	PermissionReadDraft,
	PermissionReadLeaguePokemon,
	PermissionReadDraftedPokemon,
	PermissionReadGame,
	PermissionReadUser,
	PermissionReadPlayerRoster,
	PermissionReadPokemonSpecies,
}

func IsParticipationPermission(permission Permission) bool {
	return slices.Contains(participationPermissions, permission)
}

func IsReadPermission(permission Permission) bool {
	return slices.Contains(readPermissions, permission)
}

// basePermissions are shared standing grants: everything a REGULAR Player can do.
var basePermissions = []Permission{
	PermissionReadLeague,
	PermissionReadPoolEntry,
	PermissionReadPlayer,
	PermissionReadSpectator,
	PermissionReadDraftPick,
	PermissionReadClaim,
	PermissionReadDraft,
	PermissionReadLeaguePokemon,
	PermissionReadDraftedPokemon,
	PermissionReadGame,
	PermissionReadUser,
	PermissionReadPlayerRoster,
	PermissionReadPokemonSpecies,
	PermissionReportGame,
	PermissionUpdateRole,
	PermissionUpdateOwnSpectator,
}

// moderatorPermissions are the standing grants MODERATOR adds over REGULAR.
var moderatorPermissions = []Permission{
	PermissionUpdatePlayer,
	PermissionUpdatePlayerScore,
	PermissionUpdateSpectator,
	PermissionCreateSpectator,
	PermissionDeleteSpectator,
	PermissionUpdateLeague,
	PermissionUpdateDraft,
	PermissionUpdateLeaguePokemon,
	PermissionUpdateDraftedPokemon,
	PermissionUpdateGame,
	PermissionCreatePlayer,
	PermissionCreateLeaguePokemon,
	PermissionCreateGame,
	PermissionDeleteGame,
	PermissionStartTransferPeriod,
	PermissionEndTransferPeriod,
	PermissionFinalizeGame,
	PermissionCreatePoolEntry,
	PermissionUpdatePoolEntry,
}

// ownerPermissions are the standing grants OWNER adds over MODERATOR.
var ownerPermissions = []Permission{
	PermissionCreateLeague,
	PermissionDeleteLeague,
	PermissionDeletePlayer,
	PermissionDeleteDraft,
	PermissionDeleteLeaguePokemon,
	PermissionDeleteDraftedPokemon,
	PermissionUpdateUser,
	PermissionDeleteUser,
	PermissionCreateDraft,
	PermissionCreatePlayerRoster,
	PermissionDeletePlayerRoster,
	PermissionUpdatePlayerRoster,
	PermissionDeletePoolEntry,
}

var rolePermissions = make(map[MemberRole]map[Permission]bool)

func init() {
	// Initialize permissions for each role
	for _, role := range []MemberRole{
		MRoleRegular, MRoleModerator, MRoleOwner,
		MRoleSpectator, MRoleSpectatorModerator, MRoleSpectatorOwner,
	} {
		rolePermissions[role] = make(map[Permission]bool)
	}

	// Player roles: establish standing grants first.
	setPermissions(MRoleRegular, basePermissions)
	setPermissions(MRoleRegular, participationPermissions)
	inheritPermissions(MRoleModerator, MRoleRegular)
	setPermissions(MRoleModerator, moderatorPermissions)
	inheritPermissions(MRoleOwner, MRoleModerator)
	setPermissions(MRoleOwner, ownerPermissions)

	// Spectator roles: Derivative of the matching player role, minus
	// participation, a spectator's role can never grant drafting/claims.
	inheritPermissions(MRoleSpectator, MRoleRegular)
	removePermissions(MRoleSpectator, participationPermissions)
	inheritPermissions(MRoleSpectatorModerator, MRoleModerator)
	removePermissions(MRoleSpectatorModerator, participationPermissions)
	inheritPermissions(MRoleSpectatorOwner, MRoleOwner)
	removePermissions(MRoleSpectatorOwner, participationPermissions)
}

func setPermissions(role MemberRole, perms []Permission) {
	for _, perm := range perms {
		rolePermissions[role][perm] = true
	}
}

func removePermissions(role MemberRole, perms []Permission) {
	for _, perm := range perms {
		delete(rolePermissions[role], perm)
	}
}

func inheritPermissions(child, parent MemberRole) {
	for perm, has := range rolePermissions[parent] {
		if has {
			rolePermissions[child][perm] = true
		}
	}
}
