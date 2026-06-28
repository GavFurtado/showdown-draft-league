package rbac

// Permission represents a granular action:resource permission.
type Permission string

// Define permission constants
const (
	// League Permissions
	// PermissionCreateLeague is redundant; League is created then Players
	// creating a League is tied to User and not a Player or their role. Kept in to understand the roles
	PermissionCreateLeague Permission = "create:league"
	PermissionReadLeague   Permission = "read:league"
	PermissionUpdateLeague Permission = "update:league"
	PermissionDeleteLeague Permission = "delete:league"

	// Player Permissions
	PermissionCreatePlayer      Permission = "create:player"
	PermissionReadPlayer        Permission = "read:player"
	PermissionUpdatePlayer      Permission = "update:player_info"
	PermissionUpdatePlayerScore Permission = "update:player_score"
	PermissionDeletePlayer      Permission = "delete:player"

	// PoolEntry Permissions
	PermissionCreatePoolEntry Permission = "create:pool_entry"
	PermissionReadPoolEntry   Permission = "read:pool_entry"
	PermissionUpdatePoolEntry Permission = "update:pool_entry"
	PermissionDeletePoolEntry Permission = "delete:pool_entry"

	// Member Permissions
	PermissionCreateMember      Permission = "create:member"
	PermissionReadMember        Permission = "read:member"
	PermissionUpdateMember      Permission = "update:member"
	PermissionUpdateMemberScore Permission = "update:member_score"
	PermissionDeleteMember      Permission = "delete:member"

	// DraftPick Permissions
	PermissionReadDraftPick Permission = "read:draft_pick"

	// Claim Permissions
	PermissionReadClaim Permission = "read:claim"

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

	// User Permissions (for admin-like actions on users)
	PermissionReadUser   Permission = "read:user"
	PermissionUpdateUser Permission = "update:user"
	PermissionDeleteUser Permission = "delete:user"

	// PlayerRoster Permissions
	PermissionCreatePlayerRoster Permission = "create:player_roster"
	PermissionReadPlayerRoster   Permission = "read:player_roster"
	PermissionUpdatePlayerRoster Permission = "update:player_roster"
	PermissionDeletePlayerRoster Permission = "delete:player_roster"

	// PokemonSpecies Permissions (likely read-only for most roles)
	PermissionReadPokemonSpecies Permission = "read:pokemon_species"
)

var rolePermissions = make(map[MemberRole]map[Permission]bool)

func init() {
	// Initialize permissions for each role
	rolePermissions[MRoleMember] = make(map[Permission]bool)
	rolePermissions[MRoleModerator] = make(map[Permission]bool)
	rolePermissions[MRoleOwner] = make(map[Permission]bool)

	setPermissions(MRoleMember,
		PermissionReadLeague,
		PermissionReadPlayer,
		PermissionReadDraft,
		PermissionReadLeaguePokemon,
		PermissionReadDraftedPokemon,
		PermissionReadGame,
		PermissionReadUser,
		PermissionReadPlayerRoster,
		PermissionReadPokemonSpecies,
		PermissionReleaseDraftedPokemon,
		PermissionUpdatePlayer,
		PermissionCreateDraftedPokemon,
		PermissionReportGame,

		PermissionReadPoolEntry,
		PermissionReadMember,
		PermissionUpdateMember,
		PermissionReadDraftPick,
		PermissionReadClaim,
	)

	inheritPermissions(MRoleModerator, MRoleMember)
	setPermissions(MRoleModerator,
		PermissionUpdatePlayerScore,
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
		PermissionCreateMember,
		PermissionUpdateMemberScore,
	)

	inheritPermissions(MRoleOwner, MRoleModerator)
	setPermissions(MRoleOwner,
		PermissionCreateLeague,
		PermissionDeleteLeague,
		PermissionDeletePlayer,
		PermissionDeleteDraft,
		PermissionDeleteLeaguePokemon,
		PermissionDeleteDraftedPokemon,
		PermissionUpdateUser,
		PermissionDeleteUser,
		PermissionCreateLeaguePokemon,
		PermissionCreateDraft,
		PermissionCreatePlayerRoster,
		PermissionDeletePlayerRoster,
		PermissionUpdatePlayerRoster,

		PermissionDeletePoolEntry,
		PermissionDeleteMember,
	)
}

func setPermissions(role MemberRole, perms ...Permission) {
	for _, perm := range perms {
		rolePermissions[role][perm] = true
	}
}

func inheritPermissions(child, parent MemberRole) {
	for perm, has := range rolePermissions[parent] {
		if has {
			rolePermissions[child][perm] = true
		}
	}
}
