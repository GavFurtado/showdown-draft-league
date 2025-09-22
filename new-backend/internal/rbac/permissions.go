package rbac

// Permission represents a granular action:resource permission.
type Permission string

// Define permission constants
const (
	// League Permissions
	CreateLeague Permission = "create:league"
	ReadLeague   Permission = "read:league"
	UpdateLeague Permission = "update:league"
	DeleteLeague Permission = "delete:league"

	// Player Permissions
	CreatePlayer Permission = "create:player"
	ReadPlayer   Permission = "read:player"
	UpdatePlayer Permission = "update:player"
	DeletePlayer Permission = "delete:player"

	// Draft Permissions
	CreateDraft Permission = "create:draft"
	ReadDraft   Permission = "read:draft"
	UpdateDraft Permission = "update:draft"
	DeleteDraft Permission = "delete:draft"

	// LeaguePokemon Permissions
	CreateLeaguePokemon Permission = "create:league_pokemon"
	ReadLeaguePokemon   Permission = "read:league_pokemon"
	UpdateLeaguePokemon Permission = "update:league_pokemon"
	DeleteLeaguePokemon Permission = "delete:league_pokemon"

	// DraftedPokemon Permissions
	CreateDraftedPokemon Permission = "create:drafted_pokemon"
	ReadDraftedPokemon   Permission = "read:drafted_pokemon"
	UpdateDraftedPokemon Permission = "update:drafted_pokemon"
	DeleteDraftedPokemon Permission = "delete:drafted_pokemon"

	// Game Permissions
	CreateGame Permission = "create:game"
	ReadGame   Permission = "read:game"
	UpdateGame Permission = "update:game"
	DeleteGame Permission = "delete:game"

	// User Permissions (for admin-like actions on users)
	ReadUser   Permission = "read:user"
	UpdateUser Permission = "update:user"
	DeleteUser Permission = "delete:user"

	// PlayerRoster Permissions
	CreatePlayerRoster Permission = "create:player_roster"
	ReadPlayerRoster   Permission = "read:player_roster"
	UpdatePlayerRoster Permission = "update:player_roster"
	DeletePlayerRoster Permission = "delete:player_roster"

	// PokemonSpecies Permissions (likely read-only for most roles)
	ReadPokemonSpecies Permission = "read:pokemon_species"
)

// rolePermissions maps each PlayerRole to a set of permissions it possesses.
// This map is initialized in the init() function to handle inheritance.
var rolePermissions = make(map[PlayerRole]map[Permission]bool)

func init() {
	// Initialize permissions for each role
	rolePermissions[Member] = make(map[Permission]bool)
	rolePermissions[Moderator] = make(map[Permission]bool)
	rolePermissions[Owner] = make(map[Permission]bool)

	// Member permissions
	setPermissions(Member,
		ReadLeague,
		ReadPlayer,
		ReadDraft,
		ReadLeaguePokemon,
		ReadDraftedPokemon,
		ReadGame,
		ReadUser,
		ReadPlayerRoster,
		ReadPokemonSpecies,
	)

	// Moderator permissions inherit from Member and add more
	inheritPermissions(Moderator, Member)
	setPermissions(Moderator,
		UpdateLeague,
		UpdatePlayer,
		UpdateDraft,
		UpdateLeaguePokemon,
		UpdateDraftedPokemon,
		UpdateGame,
		CreatePlayer,
		CreateDraftedPokemon,
		CreateGame,
		DeleteGame,
	)

	// Owner permissions inherit from Moderator and add more
	inheritPermissions(Owner, Moderator)
	setPermissions(Owner,
		CreateLeague,
		DeleteLeague,
		DeletePlayer,
		DeleteDraft,
		DeleteLeaguePokemon,
		DeleteDraftedPokemon,
		UpdateUser,
		DeleteUser,
		CreateLeaguePokemon,
		CreateDraft,
		CreatePlayerRoster,
		DeletePlayerRoster,
		UpdatePlayerRoster,
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
