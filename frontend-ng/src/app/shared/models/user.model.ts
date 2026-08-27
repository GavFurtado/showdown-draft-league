// Wire model for GET /api/users/me (bare User, no preloads). PascalCase by design —
// the server data IS the model; subset contract: add fields as we consume them.
export interface User {
  ID: string;
  DiscordID: string;
  DiscordUsername: string;
  DiscordAvatarURL: string;
  /** null until the user completes onboarding (new signups). */
  ShowdownUsername: string | null;
  Role: string;
}
