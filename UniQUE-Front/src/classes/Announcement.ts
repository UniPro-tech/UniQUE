import { AuthorizationErrors } from "@/errors/AuthorizationErrors";
import { FrontendErrors } from "@/errors/FrontendErrors";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";
import { apiDelete, apiGet, apiPatch, apiPost } from "@/libs/apiClient";
import { toCamelcase } from "@/libs/snakeCamelUtil";
import type { ProfileData } from "./Profile";
import type { UserData } from "./types/User";

export interface AnnouncementData {
  id: string;
  title: string;
  content: string;
  isPinned: boolean;
  createdBy: Omit<
    UserData,
    | "email"
    | "externalEmail"
    | "emailVerified"
    | "status"
    | "createdAt"
    | "updatedAt"
    | "isTotpEnabled"
  > & { profile: Omit<ProfileData, "createdAt" | "updatedAt"> };
  createdAt: string | Date;
  updatedAt: string | Date;
  deletedAt: string | Date | null | undefined;
}

export class Announcement {
  id: string;
  title: string;
  content: string;
  isPinned: boolean;
  createdBy: Omit<
    UserData,
    | "email"
    | "externalEmail"
    | "emailVerified"
    | "status"
    | "createdAt"
    | "updatedAt"
    | "isTotpEnabled"
  > & { profile: Omit<ProfileData, "createdAt" | "updatedAt" | "joinedAt"> };
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;

  constructor(data: AnnouncementData) {
    this.id = data.id;
    this.title = data.title;
    this.content = data.content;
    this.isPinned = data.isPinned;
    this.createdBy = data.createdBy;
    this.createdAt =
      data.createdAt instanceof Date
        ? data.createdAt
        : new Date(data.createdAt);
    this.updatedAt =
      data.updatedAt instanceof Date
        ? data.updatedAt
        : new Date(data.updatedAt);
    this.deletedAt =
      data.deletedAt === null || data.deletedAt === undefined
        ? null
        : data.deletedAt instanceof Date
          ? data.deletedAt
          : new Date(data.deletedAt);
  }

  get isDeleted(): boolean {
    return this.deletedAt !== null;
  }

  // ------- Converte Methods --------

  static fromJson(data: AnnouncementData): Announcement {
    return new Announcement(data);
  }

  toJson(): AnnouncementData {
    return {
      id: this.id,
      title: this.title,
      content: this.content,
      isPinned: this.isPinned,
      createdBy: this.createdBy,
      createdAt: this.createdAt.toISOString(),
      updatedAt: this.updatedAt.toISOString(),
      deletedAt: this.deletedAt ? this.deletedAt.toISOString() : null,
    };
  }

  // ------- CRUD Methods --------

  static async create(
    data: Omit<
      AnnouncementData,
      "id" | "createdBy" | "createdAt" | "updatedAt" | "deletedAt"
    >,
  ): Promise<Announcement> {
    const response = await apiPost("/announcements", {
      ...data,
    });
    const responseData = await response.json();
    return new Announcement({
      ...responseData,
    });
  }

  static async getAll(): Promise<Announcement[]> {
    const response = await apiGet("/announcements");
    const responseData = toCamelcase<{ data: AnnouncementData[] }>(
      await response.json(),
    );
    return responseData.data.map(
      (data: AnnouncementData) => new Announcement(data),
    );
  }

  static async getById(id: string): Promise<Announcement | null> {
    const response = await apiGet(`/announcements/${id}`);
    if (!response.ok) {
      if (response.status === 404) {
        return null;
      }
      switch (response.status) {
        case 400:
          throw FrontendErrors.InvalidInput;
        case 401:
          throw AuthorizationErrors.Unauthorized;
        case 403:
          throw AuthorizationErrors.AccessDenied;
        default: {
          const data = await response.json();
          console.log("API Error:", data);
          throw ResourceApiErrors.ApiServerInternalError;
        }
      }
    }
    const responseData = toCamelcase<AnnouncementData>(await response.json());
    return new Announcement({ ...responseData, id: id });
  }

  static async updateById(
    id: string,
    data: Partial<
      Omit<
        AnnouncementData,
        "id" | "createdBy" | "createdAt" | "updatedAt" | "deletedAt"
      >
    >,
  ): Promise<Announcement> {
    const response = await apiPatch(`/announcements/${id}`, {
      ...data,
    });
    if (!response.ok) {
      switch (response.status) {
        case 400:
          throw FrontendErrors.InvalidInput;
        case 401:
          throw AuthorizationErrors.Unauthorized;
        case 404:
          throw ResourceApiErrors.ResourceNotFound;
        case 403:
          throw AuthorizationErrors.AccessDenied;
        default:
          throw ResourceApiErrors.ApiServerInternalError;
      }
    }
    const responseData = toCamelcase<{ data: AnnouncementData }>(
      await response.json(),
    );
    return new Announcement({ ...responseData.data, id: id });
  }

  static async deleteById(id: string): Promise<void> {
    const response = await apiDelete(`/announcements/${id}`);
    if (!response.ok) {
      switch (response.status) {
        case 400:
          throw FrontendErrors.InvalidInput;
        case 401:
          throw AuthorizationErrors.Unauthorized;
        case 403:
          throw AuthorizationErrors.AccessDenied;
        default:
          throw ResourceApiErrors.ApiServerInternalError;
      }
    }
  }

  // ------- Instance Methods --------

  async pin(): Promise<void> {
    await Announcement.updateById(this.id, { isPinned: true });
    this.isPinned = true;
  }

  async unpin(): Promise<void> {
    await Announcement.updateById(this.id, { isPinned: false });
    this.isPinned = false;
  }

  async delete(): Promise<void> {
    await Announcement.deleteById(this.id);
    this.deletedAt = new Date();
  }

  async save(): Promise<void> {
    if (this.isDeleted) {
      throw new Error("Cannot save a deleted announcement");
    }
    if (!this.id) {
      const newAnnouncement = await Announcement.create({
        title: this.title,
        content: this.content,
        isPinned: this.isPinned,
      });
      this.id = newAnnouncement.id;
      this.createdAt = newAnnouncement.createdAt;
      this.updatedAt = newAnnouncement.updatedAt;
    } else {
      const updatedAnnouncement = await Announcement.updateById(this.id, {
        title: this.title,
        content: this.content,
        isPinned: this.isPinned,
      });
      this.updatedAt = updatedAnnouncement.updatedAt;
    }
  }
}
