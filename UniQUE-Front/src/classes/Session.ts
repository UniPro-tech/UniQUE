import { cookies } from "next/headers";
import { apiGet, createApiClient } from "@/libs/apiClient";
import { ParseJwt, VerifyJwt } from "@/libs/jwt";
import { toCamelcase } from "@/libs/snakeCamelUtil";
import type { UserData } from "./types/User";
import { User } from "./User";

export interface SessionData {
  id: string;
  userId: string;
  ipAddress: string;
  userAgent: string;
  expiresAt: string | Date;
  lastLoginAt: string | Date;
  createdAt: string | Date;
  updatedAt: string | Date;
  deletedAt: string | Date | null;
}

interface SessionVerifyResponse {
  valid: boolean;
  userId: string;
  expiresAt: string;
}

export class Session {
  id: string;
  userId: string;
  ipAddress: string;
  userAgent: string;
  expiresAt: Date;
  lastLoginAt: Date;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;

  get isDeleted(): boolean {
    return this.deletedAt !== null;
  }

  constructor(data: SessionData) {
    this.id = data.id;
    this.userId = data.userId;
    this.ipAddress = data.ipAddress;
    this.userAgent = data.userAgent;
    this.expiresAt =
      data.expiresAt instanceof Date
        ? data.expiresAt
        : new Date(data.expiresAt);
    this.lastLoginAt =
      data.lastLoginAt instanceof Date
        ? data.lastLoginAt
        : new Date(data.lastLoginAt);
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

  static fromJson(data: SessionData): Session {
    return new Session(data);
  }

  toJson(): SessionData {
    return {
      id: this.id,
      userId: this.userId,
      ipAddress: this.ipAddress,
      userAgent: this.userAgent,
      expiresAt:
        this.expiresAt instanceof Date
          ? this.expiresAt.toISOString()
          : this.expiresAt,
      lastLoginAt:
        this.lastLoginAt instanceof Date
          ? this.lastLoginAt.toISOString()
          : this.lastLoginAt,
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

  static async getByUserId(userId: string): Promise<Session[]> {
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const queryParams = new URLSearchParams({ user_id: userId });
    const response = await apiClient.get(
      `/internal/sessions?${queryParams.toString()}`,
    );
    const resData = toCamelcase<{ data: SessionData[] }>(await response.json());
    return resData.data.map((sessionData) => Session.fromJson(sessionData));
  }

  static async deleteById(id: string): Promise<void> {
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const response = await apiClient.delete(`/internal/sessions/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to delete session with id ${id}`);
    }
  }

  static async isValidSessionJWT(token: string): Promise<boolean> {
    const isValid = VerifyJwt(token);
    if (!isValid) {
      return false;
    }
    const parsedToken = ParseJwt(token);
    if (!parsedToken || typeof parsedToken !== "object" || !parsedToken.sub) {
      return false;
    }
    const jti =
      typeof parsedToken.jti === "string"
        ? parsedToken.jti
        : String(parsedToken.jti);
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const queryParams = new URLSearchParams({ jti });
    const response = await apiClient.get(
      `/internal/session_verify?${queryParams.toString()}`,
    );
    const data = toCamelcase<SessionVerifyResponse>(await response.json());
    return data.valid;
  }

  static async getSessionFromJWT(token: string): Promise<Session | null> {
    const isValid = VerifyJwt(token);
    if (!isValid) {
      return null;
    }
    const apiClient = createApiClient(`${process.env.AUTH_API_URL}`);
    const parsedToken = ParseJwt(token);
    if (!parsedToken || typeof parsedToken !== "object" || !parsedToken.sub) {
      return null;
    }
    const subject = String(parsedToken.sub);
    if (!subject) {
      return null;
    }
    if (!subject.startsWith("SID_")) {
      return null;
    }
    const sessionId = subject.substring(4);
    const response = await apiClient.get(`/internal/sessions/${sessionId}`);
    if (!response.ok) {
      return null;
    }
    const sessionData = toCamelcase<SessionData>(await response.json());
    return Session.fromJson(sessionData);
  }

  static async getCurrent(): Promise<Session | null> {
    const cookieStore = await cookies();
    const token = cookieStore.get("session_jwt")?.value;
    if (!token) {
      return null;
    }
    return await Session.getSessionFromJWT(token);
  }

  // ------ Instance Methods ------

  async delete(): Promise<void> {
    await Session.deleteById(this.id);
  }

  async getUser(): Promise<User> {
    const response = await apiGet(`/users/${this.userId}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch user with id ${this.userId}`);
    }
    const userData = toCamelcase<UserData>(await response.json());
    return User.fromJson(userData);
  }
}
