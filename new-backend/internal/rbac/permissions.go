package rbac

// Permission represents a granular action:resource permission.
type Permission string

// Define permission constants
const (
	// League Permissions
	// PermissionCreateLeague is redundant; League is created then Players
	//	reating a League is tied to User and not a Player or their role. Kept in to understand the roles (ig)
	PermissionCreateLeague Permission = "create:league"
	PermissionReadLeague   Permission = "read:league"
	PermissionUpdateLeague Permission = "update:league"
	PermissionDeleteLeague Permission = "delete:league"

	// Player Permissions
	PermissionCreatePlayer Permission = "create:player"
	PermissionReadPlayer   Permission = "read:player"
	PermissionUpdatePlayer Permission = "update:player"
	PermissionDeletePlayer Permission = "delete:player"

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
	PermissionCreateGame Permission = "create:game"
	PermissionReadGame   Permission = "read:game"
	PermissionUpdateGame Permission = "update:game"
	PermissionDeleteGame Permission = "delete:game"
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

// rolePermissions maps each PlayerRole to a set of permissions it possesses.
// This map is initialized in the init() function to handle inheritance.
var rolePermissions = make(map[PlayerRole]map[Permission]bool)

func init() {
	// Initialize permissions for each role
	rolePermissions[PRoleMember] = make(map[Permission]bool)
	rolePermissions[PRoleModerator] = make(map[Permission]bool)
	rolePermissions[PRoleOwner] = make(map[Permission]bool)

	// Member permissions
	setPermissions(PRoleMember,
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
	)

	// Moderator permissions inherit from Member and add more
	inheritPermissions(PRoleModerator, PRoleMember)
	setPermissions(PRoleModerator,
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
	)

	// Owner permissions inherit from Moderator and add more
	inheritPermissions(PRoleOwner, PRoleModerator)
	setPermissions(PRoleOwner,
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
	)
}

// setPermissions is a helper to assign multiple permissions to a role.
func setPermissions(role PlayerRole, perms ...Permission) {
	for _, perm := range perms {
		rolePermissions[role][perm] = true
	}
}

// inheritPermissions copies all permissions from a parent role to a child role.
func inheritPermissions(child, parent PlayerRole) {
	for perm, has := range rolePermissions[parent] {
		if has {
			rolePermissions[child][perm] = true
		}
	}
}
