import { notFound } from "next/navigation";
import { Session } from "@/classes/Session";
import { User } from "@/classes/User";
import Profile from "@/components/Cards/Profile";
import { PermissionBitsFields } from "@/constants/Permission";

export const metadata = {
  title: "ユーザー詳細",
  description: "ユーザー詳細ページ",
};

export default async function UserDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let user: User | null = null;
  try {
    user = await User.getById(id);
  } catch (error) {
    console.error("Failed to fetch user:", error);
    notFound();
  }

  if (!user) {
    notFound();
  }

  const currentUser = await (await Session.getCurrent())?.getUser();
  const canUpdate = await currentUser?.hasPermission(PermissionBitsFields.USER_UPDATE);

  const userData = user.toJson();
  return (
    <Profile
      user={userData}
      variant={canUpdate ? "admin" : "detail"}
      showTimestamps
    />
  );
}
