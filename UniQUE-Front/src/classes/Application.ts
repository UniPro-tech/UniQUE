import { ResourceApiErrors } from "@/errors/ResourceApiErrors";
import { apiDelete, apiGet, apiPatch, apiPost } from "@/libs/apiClient";
import { toCamelcase } from "@/libs/snakeCamelUtil";
import type { UserData } from "./types/User";
import { User } from "./User";

export interface ApplicationData {
  id: string;
  name: string;
  description: string | null;
  websiteUrl: string | null;
  privacyPolicyUrl: string | null;
  userId?: string;
  owner?: User | UserData;
  clientSecret: string | null;
  publicClient?: boolean;
  createdAt: string | Date;
  updatedAt: string | Date;
  deletedAt: string | Date | null;
}

export class Application {
  id: string;
  name: string;
  description: string | null;
  websiteUrl: string | null;
  privacyPolicyUrl: string | null;
  private ownerId: string;
  clientSecret: string | null;
  publicClient: boolean;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;

  get isDeleted(): boolean {
    return this.deletedAt !== null;
  }

  constructor(data: ApplicationData) {
    this.id = data.id;
    this.name = data.name;
    this.description = data.description;
    this.websiteUrl = data.websiteUrl;
    this.privacyPolicyUrl = data.privacyPolicyUrl;
    if (!data.userId && !data.owner) {
      throw new Error("Either ownerId or owner must be provided");
    }
    if (data.owner) {
      this.ownerId = data.owner.id;
    } else if (data.userId) {
      this.ownerId = data.userId;
    } else {
      throw new Error("Either ownerId or owner must be provided");
    }
    this.clientSecret = data.clientSecret || null;
    this.publicClient = data.publicClient || false;
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

  // This static method allows you to create an Application instance from a plain object, which is useful when parsing JSON data.
  static fromJson(data: ApplicationData): Application {
    return new Application(data);
  }

  // This method converts the Application instance back to a plain object that can be easily serialized to JSON.
  async toJson(): Promise<ApplicationData> {
    const owner = await this.getOwner().then((user) => user.toJson());
    return {
      id: this.id,
      name: this.name,
      description: this.description,
      websiteUrl: this.websiteUrl,
      privacyPolicyUrl: this.privacyPolicyUrl,
      userId: this.ownerId,
      owner,
      clientSecret: this.clientSecret,
      publicClient: this.publicClient,
      createdAt: this.createdAt.toISOString(),
      updatedAt: this.updatedAt.toISOString(),
      deletedAt: this.deletedAt ? this.deletedAt.toISOString() : null,
    };
  }

  // ------ CRUD Static Methods ------

  static async create(
    applicationData: Omit<
      ApplicationData,
      "id" | "createdAt" | "updatedAt" | "deletedAt"
    >,
  ): Promise<Application> {
    if (!applicationData.userId && !applicationData.owner) {
      throw new Error("Either userId or owner must be provided");
    }
    const response = await apiPost("/applications", applicationData);

    if (!response.ok) {
      throw new Error(`Failed to create application: ${response.statusText}`);
    }

    const responseJson = toCamelcase<ApplicationData>(await response.json());
    return Application.fromJson(responseJson);
  }

  static async getById(id: string): Promise<Application> {
    const response = await apiGet(`/applications/${id}`);

    if (!response.ok) {
      if (response.status === 404) {
        throw ResourceApiErrors.ResourceNotFound;
      }
      throw new Error(`Failed to fetch application: ${response.statusText}`);
    }

    const responseJson = toCamelcase<ApplicationData>(await response.json());

    return Application.fromJson(responseJson);
  }

  static async updateById(
    id: string,
    applicationData: Partial<
      Omit<ApplicationData, "id" | "createdAt" | "updatedAt" | "deletedAt">
    >,
  ): Promise<Application> {
    const response = await apiPatch(`/applications/${id}`, applicationData);

    if (!response.ok) {
      throw new Error(`Failed to update application: ${response.statusText}`);
    }

    const responseJson = toCamelcase<ApplicationData>(await response.json());

    return Application.fromJson(responseJson);
  }

  static async deleteById(id: string): Promise<void> {
    const response = await apiDelete(`/applications/${id}`);

    if (!response.ok) {
      throw new Error(`Failed to delete application: ${response.statusText}`);
    }
  }

  static async getAll(): Promise<Application[]> {
    const response = await apiGet("/applications");

    if (!response.ok) {
      throw new Error(`Failed to fetch applications: ${response.statusText}`);
    }

    const responseJson = toCamelcase<{ data: ApplicationData[] }>(
      await response.json(),
    );

    return responseJson.data.map((data) => Application.fromJson(data));
  }

  // ------ Instance Methods ------

  async save(
    applicationData?: Partial<Omit<ApplicationData, "id">>,
  ): Promise<void> {
    const updatedApplication = await Application.updateById(
      this.id,
      applicationData || {
        name: this.name,
        description: this.description,
        websiteUrl: this.websiteUrl,
        privacyPolicyUrl: this.privacyPolicyUrl,
        userId: this.ownerId,
        clientSecret: this.clientSecret,
      },
    );
    Object.assign(this, updatedApplication);
  }

  async delete(): Promise<void> {
    await Application.deleteById(this.id);
    this.deletedAt = new Date();
  }

  async getOwner(): Promise<User> {
    const user = await User.getById(this.ownerId);
    if (!user) {
      throw new Error(`Owner with id ${this.ownerId} not found`);
    }
    return user;
  }

  async getRedirectUris(): Promise<string[]> {
    const res = await apiGet(`/applications/${this.id}/redirect_uris`);
    const data = toCamelcase<{
      data: [
        {
          applicationId: string;
          uri: string;
          createdAt: string;
          updatedAt: string;
        },
      ];
    }>(await res.json());
    return data.data.map((item) => item.uri);
  }

  async addRedirectUri(uri: string): Promise<void> {
    const response = await apiPost(`/applications/${this.id}/redirect_uris`, {
      uri,
    });
    if (!response.ok) {
      throw new Error(`Failed to add redirect URI: ${response.statusText}`);
    }
  }

  async removeRedirectUri(uri: string): Promise<void> {
    const uriQueryParams = new URLSearchParams({ uri });
    const response = await apiDelete(
      `/applications/${this.id}/redirect_uris?${uriQueryParams.toString()}`,
    );

    if (!response.ok) {
      console.log("Failed to remove redirect URI:", {
        status: response.status,
        statusText: response.statusText,
        response: await response.text(),
        uri,
      });
      throw new Error(`Failed to remove redirect URI: ${response.statusText}`);
    }
  }
}
