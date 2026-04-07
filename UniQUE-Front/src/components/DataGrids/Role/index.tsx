import { Role, type RoleData } from "@/classes/Role";

export default async function RolesDataGrid({
  rows,
}: {
  rows?: RoleData[] | Role[];
}) {
  if (!rows) {
    const fetched = await Role.getAll();
    rows = fetched.map((role) => role.toJson());
  } else {
    rows = rows.map((role) => (role instanceof Role ? role.toJson() : role));
  }
  const RolesDataGridClient = (await import("../Roles/Client")).default;
  return <RolesDataGridClient rows={rows as RoleData[]} />;
}
