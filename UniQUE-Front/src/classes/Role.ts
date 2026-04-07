import type { PermissionTexts } from "@/constants/Permission";
import {
  apiDelete,
  apiGet,
  apiPatch,
  apiPost,
  apiPut,
  stringJsonParseSafe,
} from "@/libs/apiClient";
import { toCamelcase } from "@/libs/snakeCamelUtil";
import { User } from "./User";

export interface RoleData {
  id: string;
  name: string;
  customId: string;
  description: string;
  permissionBitmask: number;
  isDefault: boolean;
  createdAt: string | Date;
  updatedAt: string | Date;
  deletedAt: string | Date | null;
}

export class Role {
  id: string;
  name: string;
  customId: string;
  description: string;
  permissionBitmask: bigint;
  isDefault: boolean;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;

  get isDeleted(): boolean {
    return this.deletedAt !== null;
  }

  get permissons(): PermissionTexts[] {
    return [];
  }

  constructor(
    data: Omit<RoleData, "permissionBitmask"> & {
      permissionBitmask: bigint | number;
    },
  ) {
    this.id = data.id;
    this.name = data.name;
    this.customId = data.customId;
    this.description = data.description;
    this.permissionBitmask = BigInt(data.permissionBitmask);
    this.isDefault = data.isDefault;
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

  static fromJson(data: RoleData): Role {
    return new Role(data);
  }

  toJson(): RoleData {
    return {
      id: this.id,
      name: this.name,
      customId: this.customId,
      description: this.description,
      permissionBitmask: Number(this.permissionBitmask),
      isDefault: this.isDefault,
      createdAt:
        this.createdAt instanceof Date
          ? this.createdAt.toISOString()
          : this.createdAt,
      updatedAt:
        this.updatedAt instanceof Date
          ? this.updatedAt.toISOString()
          : this.updatedAt,
      deletedAt: this.deletedAt ? this.deletedAt.toISOString() : null,
    };
  }

  // ------ Static Methods ------

  static async create(
    roleData: Omit<
      RoleData,
      "id" | "createdAt" | "updatedAt" | "deletedAt" | "permissionBitmask"
    > & {
      permissionBitmask: bigint;
    },
  ): Promise<Role> {
    const response = await apiPost("/roles", roleData);
    if (!response.ok) {
      throw new Error(`Failed to create role: ${response.statusText}`);
    }

    const responseData = toCamelcase<RoleData>(await response.json());
    return Role.fromJson(responseData);
  }

  static async getAll(): Promise<Role[]> {
    const response = await apiGet("/roles");
    if (!response.ok) {
      throw new Error(`Failed to fetch roles: ${response.statusText}`);
    }
    const responseJson = toCamelcase<{ data: RoleData[] }>(
      await response.json(),
    );
    return responseJson.data.map((roleData) => Role.fromJson(roleData));
  }

  static async getById(id: string): Promise<Role | null> {
    const response = await apiGet(`/roles/${id}`);
    if (!response.ok) {
      if (response.status === 404) {
        return null;
      }
      throw new Error(`Failed to fetch role: ${response.statusText}`);
    }
    const responseData = await response.text().then((text) => {
      try {
        return toCamelcase<RoleData>(stringJsonParseSafe(text));
      } catch (error) {
        throw new Error(`Failed to parse role data: ${text} - ${error}`);
      }
    });
    return Role.fromJson(responseData);
  }

  static async updateById(
    id: string,
    roleData: Partial<
      Omit<RoleData, "id" | "createdAt" | "updatedAt" | "deletedAt">
    >,
  ): Promise<Role> {
    const response = await apiPatch(`/roles/${id}`, roleData);
    if (!response.ok) {
      throw new Error(`Failed to update role: ${response.statusText}`);
    }
    const responseData = await response.text().then((text) => {
      try {
        return toCamelcase<RoleData>(
          stringJsonParseSafe(text)
            .then((json: never) => json)
            .catch(() => {
              throw new Error(`Failed to parse role data: ${text}`);
            }),
        );
      } catch (error) {
        throw new Error(`Failed to parse role data: ${text} - ${error}`);
      }
    });
    return Role.fromJson(responseData);
  }

  static async deleteById(id: string): Promise<void> {
    const response = await apiDelete(`/roles/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to delete role: ${response.statusText}`);
    }
  }

  // ------ Instance Methods ------

  async save(
    roleData?: Partial<
      Omit<RoleData, "id" | "createdAt" | "updatedAt" | "deletedAt">
    >,
  ): Promise<void> {
    if (this.isDeleted) {
      throw new Error("Cannot save a deleted role.");
    }
    const updatedRole = await Role.updateById(this.id, {
      name: roleData?.name ?? this.name,
      customId: roleData?.customId ?? this.customId,
      description: roleData?.description ?? this.description,
      permissionBitmask:
        roleData?.permissionBitmask !== undefined
          ? Number(roleData.permissionBitmask)
          : Number(this.permissionBitmask),
      isDefault: roleData?.isDefault ?? this.isDefault,
    });
    Object.assign(this, updatedRole);
  }

  async delete(): Promise<void> {
    await Role.deleteById(this.id);
    this.deletedAt = new Date();
  }

  async getUsers(): Promise<User[]> {
    const response = await apiGet(`/roles/${this.id}/users`);
    if (!response.ok) {
      throw new Error(`Failed to fetch users for role: ${response.statusText}`);
    }
    const responseJson = toCamelcase<{ data: User[] }>(await response.json());
    return responseJson.data.map((userData) => User.fromJson(userData));
  }

  async addUser(userId: string): Promise<void> {
    const response = await apiPut(`/roles/${this.id}/users/${userId}`);
    if (!response.ok) {
      throw new Error(`Failed to add user to role: ${response.statusText}`);
    }
  }

  async removeUser(userId: string): Promise<void> {
    const response = await apiDelete(`/roles/${this.id}/users/${userId}`);
    if (!response.ok) {
      throw new Error(
        `Failed to remove user from role: ${response.statusText}`,
      );
    }
  }

  async asignAllUsers(): Promise<void> {
    const response = await apiPost(`/roles/${this.id}/assign_all`);
    if (!response.ok) {
      throw new Error(
        `Failed to assign all users to role: ${response.statusText}`,
      );
    }
  }
}
