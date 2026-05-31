import type { User } from "@/classes/User";
import ProfileClient, { type ExternalIdentityFetchResult } from "./Client";

interface ProfileProps {
  user: User;
  variant?: "self" | "detail" | "admin";
  onBack?: () => void;
  showTimestamps?: boolean;
}

export default async function Profile(props: ProfileProps) {
  const externalIdentities = await props.user
    .getExternalIdentities()
    .then((items) => {
      const itemData = items.map((data) => data.toJson());
      return {
        externalIdentities: itemData,
        isError: false,
      } as ExternalIdentityFetchResult;
    })
    .catch(() => ({
      externalIdentities: [],
      isError: true,
    }));
  return (
    <ProfileClient
      {...props}
      user={props.user.toJson()}
      externalIdentities={externalIdentities}
    />
  );
}
