import { UUID, ISODateTimeString } from '../../../shared/types/branded-strings';
import { User } from '../../../shared/models/user.model';
import { MemberRole } from './enums/member-role';

export interface LeagueMember {
  ID: UUID;
  LeagueID: UUID;
  UserID: UUID;
  InLeagueName?: string | null;
  TeamName?: string | null;
  Role: MemberRole;
  IsActive: boolean;
  JoinedAt: ISODateTimeString;
  User?: User;
}
