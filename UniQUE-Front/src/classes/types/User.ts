import type { Profile, ProfileData } from "../Profile";

export enum UserStatus {
  ESTABLISHED = "established",
  ACTIVE = "active",
  SUSPENDED = "suspended",
  ARCHIVED = "archived",
}

export interface UserData {
  id: string;
  customId: string;
  email: string;
  externalEmail: string | null;
  readonly pendingEmail: string | null;
  emailVerified: boolean;
  affiliationPeriod: string | null;
  isTotpEnabled: boolean | null;
  status: UserStatus;
  profile: ProfileData | Profile;
  createdAt: string | Date;
  updatedAt: string | Date;
  deletedAt: string | Date | null;
}
