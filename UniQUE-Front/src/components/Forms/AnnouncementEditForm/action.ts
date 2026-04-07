"use server";

import { Announcement } from "@/classes/Announcement";

export const updateAnnouncement = async ({
  id,
  title,
  content,
  isPinned,
}: {
  id: string;
  title: string;
  content: string;
  isPinned: boolean;
}) => {
  const res = await Announcement.updateById(id, {
    title,
    content,
    isPinned,
  });
  if (!res) {
    throw new Error("更新に失敗しました");
  }
};
