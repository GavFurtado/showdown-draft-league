package requests

// Dev-only request payloads. These routes are only registered when ENV=dev.

type DevCreateUserRequestDTO struct {
	Name             string  `json:"Name" binding:"required"`
	ShowdownUsername *string `json:"ShowdownUsername"`
}

type DevImpersonateRequestDTO struct {
	UserID string `json:"UserId" binding:"required,uuid"`
}

type DevUpsertMembershipRequestDTO struct {
	LeagueID string `json:"LeagueId" binding:"required,uuid"`
	UserID   string `json:"UserId" binding:"required,uuid"`
	Role     string `json:"Role" binding:"required,oneof=OWNER MODERATOR MEMBER"`
}
