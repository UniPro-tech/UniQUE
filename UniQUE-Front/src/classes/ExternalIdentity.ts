import { ResourceApiErrors } from "@/errors/ResourceApiErrors";
import { apiDelete, apiGet, apiPost } from "@/libs/apiClient";
import { toCamelcase } from "@/libs/snakeCamelUtil";

export interface ExternalIdentityData {
  id: string;
  username?: string;
  displayName?: string;
  avatarUrl: string;
  email?: string;
  externalUserId: string;
  idTokenClaims: unknown | null;
  provider: string;
  providerData: unknown;
  userId: string;
  createdAt: string;
  updatedAt: string;
}

export interface ExternalIdentityCreateData {
  accessToken: string;
  externalUserId: string;
  idToken?: string | null;
  provider: string;
  refreshToken: string;
  tokenExpiresAt: string;
}

export class ExternalIdentity {
  id: string;
  username?: string;
  displayName?: string;
  avatarUrl: string;
  email?: string;
  externalUserId: string;
  idTokenClaims: unknown | null;
  provider: string;
  providerData: unknown;
  readonly userId: string;
  createdAt: string;
  updatedAt: string;

  constructor(data: ExternalIdentityData) {
    this.id = data.id;
    this.username = data.username;
    this.displayName = data.displayName;
    this.avatarUrl = data.avatarUrl;
    this.email = data.email;
    this.externalUserId = data.externalUserId;
    this.idTokenClaims = data.idTokenClaims;
    this.provider = data.provider;
    this.providerData = data.providerData;
    this.userId = data.userId;
    this.createdAt = data.createdAt;
    this.updatedAt = data.updatedAt;
  }

  // ------ Converter Methods ------

  static fromJson(data: ExternalIdentityData): ExternalIdentity {
    return new ExternalIdentity(data);
  }

  toJson(): ExternalIdentityData {
    return {
      id: this.id,
      username: this.username,
      displayName: this.displayName,
      avatarUrl: this.avatarUrl,
      email: this.email,
      externalUserId: this.externalUserId,
      idTokenClaims: this.idTokenClaims,
      provider: this.provider,
      providerData: this.providerData,
      userId: this.userId,
      createdAt: this.createdAt,
      updatedAt: this.updatedAt,
    };
  }

  // ------ CRUD Methods ------

  static async getByUserId(userId: string): Promise<ExternalIdentity[]> {
    // API から userId に紐づく外部IDを取得する実装例
    const response = await apiGet(`/users/${userId}/external_identities`);
    if (!response.ok) {
      throw new Error(
        `Failed to fetch external identities: ${response.statusText}`,
      );
    }
    const data = toCamelcase<{ data: ExternalIdentityData[] }>(
      await response.json(),
    );
    return data.data.map((item: ExternalIdentityData) =>
      ExternalIdentity.fromJson(item),
    );
  }

  static async create(
    userId: string,
    data: ExternalIdentityCreateData,
  ): Promise<ExternalIdentity> {
    // API を呼び出して新しい外部IDを作成する実装例
    const response = await apiPost(
      `/users/${userId}/external_identities`,
      data,
    );
    if (!response.ok) {
      if (response.status === 409) {
        throw ResourceApiErrors.ResourceAlreadyExists;
      }
      throw new Error(
        `Failed to create external identity: ${response.statusText}`,
      );
    }
    const responseJson = toCamelcase<ExternalIdentityData>(
      await response.json(),
    );
    return ExternalIdentity.fromJson(responseJson);
  }

  static async createByEmailVerificationCode(
    emailVerifyCode: string,
    data: ExternalIdentityCreateData,
  ): Promise<ExternalIdentity> {
    if (data.provider !== "discord") {
      throw new Error("Unsupported provider for email verification linking");
    }
    const response = await apiPost(
      `/internal/users/email_verify/discord_link`,
      {
        code: emailVerifyCode,
        external_user_id: data.externalUserId,
        access_token: data.accessToken,
        refresh_token: data.refreshToken,
        token_expires_at: data.tokenExpiresAt,
      },
    );
    if (!response.ok) {
      if (response.status === 409) {
        throw ResourceApiErrors.ResourceAlreadyExists;
      }
      throw new Error(
        `Failed to link external identity via email verification: ${response.statusText}`,
      );
    }
    const responseJson = toCamelcase<ExternalIdentityData>(
      await response.json(),
    );
    return ExternalIdentity.fromJson(responseJson);
  }

  static async deleteByUserAndId(
    userid: string,
    identityId: string,
  ): Promise<void> {
    // API を呼び出して identityId に紐づく外部IDを取得する実装例
    const response = await apiDelete(
      `/users/${userid}/external_identities/${identityId}`,
    );
    if (!response.ok) {
      throw new Error(
        `Failed to delete external identity: ${response.statusText}`,
      );
    }
  }

  // ------ Instance Methods ------

  async delete(): Promise<void> {
    // API を呼び出してこの外部IDを削除する実装例
    const response = await apiDelete(
      `/users/${this.userId}/external_identities/${this.id}`,
    );
    if (!response.ok) {
      throw new Error(
        `Failed to remove external identity: ${response.statusText}`,
      );
    }
  }
}
