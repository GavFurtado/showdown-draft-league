import { User } from '../../../shared/models/user.model';

export interface LeagueMember {
  ID: string;
  LeagueID: string;
  UserID: string;
  DisplayName: string;
  InLeagueName?: string | null;
  TeamName?: string | null;
  Role: 'OWNER' | 'MODERATOR' | 'MEMBER';
  IsActive: boolean;
  JoinedAt: string;
  // Preloaded by GET /api/leagues/:id (Members.User) — optional per subset contract.
  User?: User;
}
