"use server";

import { Announcement } from "@/classes/Announcement";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

export const pinAnnouncement = async (id: string, toPin: boolean) => {
  const announce = await Announcement.getById(id);

  if (!announce) {
    throw ResourceApiErrors.ResourceNotFound;
  }

  if (toPin) {
    await announce.pin();
  } else {
    await announce.unpin();
  }
};
