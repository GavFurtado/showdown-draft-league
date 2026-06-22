package types

// DiscordWebhookPayload represents the structure for sending messages to Discord webhooks.
type DiscordWebhookPayload struct {
	Content   string                `json:"Content" gorm:"column:content"`
	Username  string                `json:"Username" gorm:"column:username"`
	AvatarURL string                `json:"AvatarURL" gorm:"column:avatar_url"`
	Embeds    []DiscordWebhookEmbed `json:"Embeds" gorm:"column:embeds"`
}

// DiscordWebhookEmbed represents an embed object within a Discord webhook payload.
type DiscordWebhookEmbed struct {
	Title       string `json:"Title" gorm:"column:title"`
	Description string `json:"Description" gorm:"column:description"`
	Color       int    `json:"Color"`
}
