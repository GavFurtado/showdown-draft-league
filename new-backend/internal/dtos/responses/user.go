package responses

type DiscordUserResponse struct {
	ID            string `json:"ID"`
	Username      string `json:"Username"`
	Discriminator string `json:"Discriminator"`
	Avatar        string `json:"Avatar"`
}
