"use server";

import { Announcement } from "@/classes/Announcement";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

export const deleteAnnouncement = async (id: string) => {
  const announce = await Announcement.getById(id);

  if (!announce) {
    throw ResourceApiErrors.ResourceNotFound;
  }

  await announce.delete();
};
