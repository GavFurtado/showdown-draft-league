package rbac

// Permission constants
const (
	// League permissions
	ActionCreateLeague       = "create:league"
	ActionReadLeague         = "read:league"
	ActionUpdateLeague       = "update:league"
	ActionDeleteLeague       = "delete:league"
	ActionManageStatusLeague = "manage_status:league"

	// Player permissions
	ActionCreatePlayer      = "create:player"
	ActionReadPlayer        = "read:player"
	ActionUpdatePlayer      = "update:player"
	ActionDeletePlayer      = "delete:player"
	ActionAssignRolePlayer  = "assign_role:player"
	ActionUpdateStatsPlayer = "update_stats:player" // For wins/losses, draft points, position

	// Draft permissions
	ActionStartDraft         = "start:draft"
	ActionMakePickDraft      = "make_pick:draft"
	ActionSkipTurnDraft      = "skip_turn:draft"
	ActionManagePeriodsDraft = "manage_periods:draft" // For starting/ending trading/free agency
	ActionReadDraft          = "read:draft"           // For viewing draft status/history

	// Game permissions
	ActionCreateGame       = "create:game"
	ActionReadGame         = "read:game"
	ActionUpdateGame       = "update:game"
	ActionDeleteGame       = "delete:game"
	ActionReportResultGame = "report_result:game"

	// League Pokemon (draft pool) permissions
	ActionCreateLeaguePokemon = "create:league_pokemon"
	ActionReadLeaguePokemon   = "read:league_pokemon"
	ActionUpdateLeaguePokemon = "update:league_pokemon" // For cost, availability
	ActionDeleteLeaguePokemon = "delete:league_pokemon"

	// Drafted Pokemon (player's pokemon) permissions
	ActionReadDraftedPokemon    = "read:drafted_pokemon"
	ActionReleaseDraftedPokemon = "release:drafted_pokemon" // Releasing own pokemon
	ActionTradeDraftedPokemon   = "trade:drafted_pokemon"
	ActionDeleteDraftedPokemon  = "delete:drafted_pokemon" // Soft delete
)

// rolePermissions defines the permissions for each PlayerRole.
// This map is initialized once at program startup using the init() function.
var rolePermissions = map[PlayerRole]map[string]bool{}

func init() {
	// Define Member permissions
	memberPermissions := []string{
		ActionReadLeague,
		ActionReadPlayer,
		ActionMakePickDraft,
		ActionReadDraft,
		ActionReadGame,
		ActionReportResultGame,
		ActionReadLeaguePokemon,
		ActionReadDraftedPokemon,
		ActionReleaseDraftedPokemon,
	}
	rolePermissions[PlayerRoleMember] = make(map[string]bool)
	for _, perm := range memberPermissions {
		rolePermissions[PlayerRoleMember][perm] = true
	}

	// copy Member permissions
	rolePermissions[PlayerRoleModerator] = make(map[string]bool)
	for perm := range rolePermissions[PlayerRoleMember] {
		rolePermissions[PlayerRoleModerator][perm] = true
	}
	// Add Moderator-specific permissions
	moderatorPermissions := []string{
		ActionUpdateLeague,
		ActionManageStatusLeague,
		ActionCreatePlayer,
		ActionUpdatePlayer,
		ActionDeletePlayer,
		ActionUpdateStatsPlayer,
		ActionStartDraft,
		ActionSkipTurnDraft,
		ActionManagePeriodsDraft,
		ActionCreateGame,
		ActionUpdateGame,
		ActionDeleteGame,
		ActionCreateLeaguePokemon,
		ActionUpdateLeaguePokemon,
		ActionDeleteLeaguePokemon,
		ActionTradeDraftedPokemon,
		ActionDeleteDraftedPokemon,
	}
	for _, perm := range moderatorPermissions {
		rolePermissions[PlayerRoleModerator][perm] = true
	}

	// copy Moderator permissions to owner
	rolePermissions[PlayerRoleOwner] = make(map[string]bool)
	for perm := range rolePermissions[PlayerRoleModerator] {
		rolePermissions[PlayerRoleOwner][perm] = true
	}
	// Add Owner-specific permissions
	ownerPermissions := []string{
		ActionDeleteLeague,
		ActionAssignRolePlayer,
	}
	for _, perm := range ownerPermissions {
		rolePermissions[PlayerRoleOwner][perm] = true
	}
}
