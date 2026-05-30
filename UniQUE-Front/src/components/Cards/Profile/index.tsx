import type { User } from "@/classes/User";
import ProfileClient from "./Client";

interface ProfileProps {
  user: User;
  variant?: "self" | "detail" | "admin";
  onBack?: () => void;
  showTimestamps?: boolean;
}

export default async function Profile(props: ProfileProps) {
  const externalIdentities = (await props.user.getExternalIdentities()).map(
    (data) => data.toJson(),
  );
  return (
    <ProfileClient
      {...props}
      user={props.user.toJson()}
      externalIdentities={externalIdentities}
    />
  );
}
