import { type NextRequest, NextResponse } from "next/server";
import { Session } from "@/classes/Session";
import { User } from "@/classes/User";

/**
 * Discord連携状態をチェックするエンドポイント
 * GET /api/auth/check-discord-link
 */
export const GET = async (_request: NextRequest) => {
  try {
    // セッションからカスタムIDを取得
    const session = await Session.getCurrent();
    if (!session) {
      return NextResponse.json(
        { linked: false, hasSession: false },
        { status: 200 },
      );
    }

    // ユーザーの外部アイデンティティを取得
    const externalIdentities = await User.getById(session.userId).then(
      async (user) => (user ? await user.getExternalIdentities() : []),
    );

    // Discord連携があるかチェック
    const hasDiscord = externalIdentities.some(
      (identity) => identity.provider === "discord",
    );

    return NextResponse.json(
      { linked: hasDiscord, hasSession: true },
      { status: 200 },
    );
  } catch (error) {
    console.error("Discord連携状態の確認エラー:", error);
    return NextResponse.json(
      { linked: false, hasSession: false },
      { status: 200 },
    );
  }
};
