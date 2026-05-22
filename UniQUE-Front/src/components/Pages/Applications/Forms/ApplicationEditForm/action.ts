"use server";

import { Application, type ApplicationData } from "@/classes/Application";

export const changeAction = async (
  data: Partial<
    Omit<ApplicationData, "createdAt" | "updatedAt" | "deletedAt">
  > & { id: string },
) => {
  await Application.updateById(data.id, data);
};
