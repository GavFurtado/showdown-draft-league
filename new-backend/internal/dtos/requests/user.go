package requests

type UserUpdateProfileRequestDTO struct {
	ShowdownName *string `json:"ShowdownName" binding:"omitempty"`
	// this is all we have for now
}
