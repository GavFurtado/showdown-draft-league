// Wire model for GET /api/users/me (bare User, no preloads). PascalCase by design —
// the server data IS the model; subset contract: add fields as we consume them.
import { UUID } from '../types/branded-strings';
import { UserRole } from './enums/user-role';

export interface User {
  ID: UUID;
  DiscordID: string;
  DiscordUsername: string;
  DiscordAvatarURL: string;
  /** Optional in-game username; null until the user sets one. */
  ShowdownUsername: string | null;
  Role: UserRole;
}
