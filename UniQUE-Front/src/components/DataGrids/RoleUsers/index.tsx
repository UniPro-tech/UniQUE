import { use } from "react";
import type { Role } from "@/classes/Role";
import type { UserData } from "@/classes/types/User";
import RoleUsersDataGridClient from "./Client";

export default function RoleUsersDataGrid({ role }: { role: Role }) {
  const users = use(role.getUsers());
  const rows: UserData[] = users.map((user) => user.toJson());
  return <RoleUsersDataGridClient rows={rows} role={role.toJson()} />;
}
