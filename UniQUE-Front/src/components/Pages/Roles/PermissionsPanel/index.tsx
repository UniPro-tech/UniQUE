import type { Role } from "@/classes/Role";
import PermissionsPanelClient from "./Client";

export default async function PermissionsPanel({ role }: { role: Role }) {
  return (
    <PermissionsPanelClient
      role={role.toJson()}
      permissionBitmask={role.permissionBitmask}
    />
  );
}
