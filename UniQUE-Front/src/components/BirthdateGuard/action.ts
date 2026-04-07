"use server";

import { Session } from "@/classes/Session";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";
import { FormRequestErrors } from "@/errors/FormRequestErrors";
import { FrontendErrors } from "@/errors/FrontendErrors";
import { VerifyCSRFToken } from "@/libs/csrf";

export const updateBirthdate = async (
  prevState: { birthdate: string; status: FormStatus | null },
  formData: FormData,
): Promise<{ birthdate: string; status: FormStatus | null }> => {
  try {
    const csrfToken = formData.get("csrfToken") as string;
    const isVerified = VerifyCSRFToken(csrfToken);
    if (!isVerified) throw FormRequestErrors.CSRFTokenMismatch;

    const birthdate = formData.get("birthdate") as string;
    if (!birthdate) throw FrontendErrors.InvalidInput;

    const session = await Session.getCurrent();
    if (!session) {
      return {
        birthdate: prevState.birthdate,
        status: {
          status: "error",
          message: "セッションが見つかりません。",
        },
      };
    }

    const user = await session.getUser();
    user.profile.birthdate = birthdate;
    await user.save();

    return {
      birthdate,
      status: {
        status: "success",
        message: "生年月日を設定しました。",
      },
    };
  } catch (error) {
    return {
      birthdate: prevState.birthdate,
      status: {
        status: "error",
        message:
          error instanceof Error
            ? error.message
            : "生年月日の設定に失敗しました。",
      },
    };
  }
};
