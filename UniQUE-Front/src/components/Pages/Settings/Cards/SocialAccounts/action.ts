"use server";

import { ExternalIdentity } from "@/classes/ExternalIdentity";

export const deleteAction = async (userId: string, identityId: string) => {
  await ExternalIdentity.deleteByUserAndId(userId, identityId);
};
