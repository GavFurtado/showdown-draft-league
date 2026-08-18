export interface LeagueMember {
  ID: string;
  LeagueID: string;
  UserID: string;
  DisplayName: string;
  Role: 'OWNER' | 'MODERATOR' | 'MEMBER';
  IsActive: boolean;
  JoinedAt: string;
}
