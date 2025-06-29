package common

// represents the basic information retrieved from Discord's /users/@me endpoint.
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

// Response Structs
