package requests

type UserUpdateProfileRequestDTO struct {
	ShowdownName *string `json:"ShowdownName" validate:"omitempty"`
	// this is all we have for now
}
