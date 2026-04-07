import type { Scope } from "@/constants/Scope";
import { createApiClient } from "@/libs/apiClient";

export interface ConsentsData {
  id: string;
  userId: string;
  applicationId: string;
  scopes: string[];
  createdAt: string | Date;
  updatedAt: string | Date;
  deletedAt: string | Date | null;
}

export class Consents {
  id: string;
  userId: string;
  applicationId: string;
  scopes: string[];
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;

  get isDeleted(): boolean {
    return this.deletedAt !== null;
  }

  constructor(data: ConsentsData) {
    this.id = data.id;
    this.userId = data.userId;
    this.applicationId = data.applicationId;
    this.scopes = data.scopes;
    this.createdAt =
      data.createdAt instanceof Date
        ? data.createdAt
        : new Date(data.createdAt);
    this.updatedAt =
      data.updatedAt instanceof Date
        ? data.updatedAt
        : new Date(data.updatedAt);
    this.deletedAt = data.deletedAt
      ? data.deletedAt instanceof Date
        ? data.deletedAt
        : new Date(data.deletedAt)
      : null;
  }

  // ------ Converter Methods ------

  static fromJson(data: ConsentsData): Consents {
    return new Consents(data);
  }

  toJson(): ConsentsData {
    return {
      id: this.id,
      userId: this.userId,
      applicationId: this.applicationId,
      scopes: this.scopes,
      createdAt:
        this.createdAt instanceof Date
          ? this.createdAt.toISOString()
          : this.createdAt,
      updatedAt:
        this.updatedAt instanceof Date
          ? this.updatedAt.toISOString()
          : this.updatedAt,
      deletedAt: this.deletedAt
        ? this.deletedAt instanceof Date
          ? this.deletedAt.toISOString()
          : this.deletedAt
        : null,
    };
  }

  // ------ Static Methods ------

  static async getByUserId(userId: string): Promise<Consents[]> {
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const response = await apiClient.get(
      `/consents?user_id=${encodeURIComponent(userId)}`,
    );
    if (!response.ok) {
      throw new Error(`Failed to fetch consents for user ${userId}`);
    }
    const data: ConsentsData[] = await response.json();
    return data.map((item) => Consents.fromJson(item));
  }

  static async getByApplicationId(applicationId: string): Promise<Consents[]> {
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const response = await apiClient.get(
      `/consents?application_id=${encodeURIComponent(applicationId)}`,
    );
    if (!response.ok) {
      throw new Error(
        `Failed to fetch consents for application ${applicationId}`,
      );
    }
    const data: ConsentsData[] = await response.json();
    return data.map((item) => Consents.fromJson(item));
  }

  static async revokeById(id: string): Promise<void> {
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const response = await apiClient.delete(
      `/consents/${encodeURIComponent(id)}`,
    );
    if (!response.ok) {
      throw new Error(`Failed to revoke consent with id ${id}`);
    }
  }

  static async isConsented(
    applicationId: string,
    userId: string,
    scope: Scope[],
  ): Promise<boolean> {
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const response = await apiClient.get(
      `/consents?application_id=${encodeURIComponent(
        applicationId,
      )}&user_id=${encodeURIComponent(userId)}&scope=${encodeURIComponent(
        scope.join(","),
      )}`,
    );
    if (!response.ok) {
      throw new Error(
        `Failed to check consent for application ${applicationId} and user ${userId}`,
      );
    }
    const data: { isConsented: boolean } = await response.json();
    return data.isConsented;
  }

  // ------ Instance Methods ------

  async revoke(): Promise<void> {
    await Consents.revokeById(this.id);
    this.deletedAt = new Date();
  }
}
