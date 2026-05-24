"use server";

import { Session, type SessionData } from "@/classes/Session";
import { FrontendErrors } from "@/errors/FrontendErrors";
import type { FormStatus } from "../../Base";

export const logoutSession = async (
  prevState: { sessions: SessionData[]; status: FormStatus | null },
  formData: FormData,
): Promise<{ sessions: SessionData[]; status: FormStatus }> => {
  try {
    const sessionId = formData.get("sessionId") as string;
    if (!sessionId) {
      throw FrontendErrors.InvalidInput;
    }

    // 個別のセッションを削除
    try {
      await Session.deleteById(sessionId);
    } catch {
      return {
        sessions: prevState.sessions,
        status: {
          status: "error",
          message: "セッションの削除に失敗しました。",
        },
      };
    }

    // セッションに削除済みフラグを付ける（削除せずにマークする）
    const updatedSessions = prevState.sessions.map((s) =>
      s.id === sessionId ? { ...s, deletedAt: new Date() } : s,
    );

    const status: FormStatus = {
      status: "success",
      message: "セッションをログアウトしました。",
    };
    return { sessions: updatedSessions, status };
  } catch (error) {
    const sessions = prevState.sessions;
    const status: FormStatus = {
      status: "error",
      message:
        error instanceof Error
          ? error.message
          : "セッションのログアウト中に予期せぬエラーが発生しました。",
    };
    return { sessions, status };
  }
};
