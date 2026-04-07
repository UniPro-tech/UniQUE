import type { NextRequest } from "next/server";

/**
 * Discord OAuth認証の開始エンドポイント
 * ユーザーをDiscordの認証ページにリダイレクトする
 */
export const GET = async (request: NextRequest) => {
  const clientId = process.env.DISCORD_CLIENT_ID;
  const redirectUri = process.env.DISCORD_REDIRECT_URI;
  const searchParams = request.nextUrl.searchParams;
  const from = searchParams.get("from"); // signup | settings | email_verify
  const code = searchParams.get("code");
  const stateParam = searchParams.get("state");

  if (!clientId || !redirectUri) {
    console.error("Discord OAuth環境変数が設定されていません");
    return new Response(
      JSON.stringify({ error: "Discord OAuth is not configured" }),
      { status: 500, headers: { "Content-Type": "application/json" } },
    );
  }

  // state パラメータを生成（CSRF対策）
  // fromパラメータをstateに含める
  const stateObj: { nonce: string; from: string; code?: string } = {
    nonce: crypto.randomUUID(),
    from: from || "settings",
  };

  if (code) {
    stateObj.code = code;
  }

  // email_verifyからの場合、stateパラメータに追加情報が含まれている
  if (stateParam) {
    try {
      const incomingState = JSON.parse(
        Buffer.from(stateParam, "base64").toString("utf-8"),
      );
      if (incomingState.from) {
        stateObj.from = incomingState.from;
      }
      if (incomingState.code) {
        stateObj.code = incomingState.code;
      }
    } catch (e) {
      console.error("Failed to parse incoming state:", e);
    }
  }

  const state = Buffer.from(JSON.stringify(stateObj)).toString("base64");

  // Discord OAuth URL を構築
  const params = new URLSearchParams({
    client_id: clientId,
    redirect_uri: redirectUri,
    response_type: "code",
    scope: "openid identify guilds.join",
    state,
  });

  const authUrl = `https://discord.com/api/oauth2/authorize?${params.toString()}`;

  return Response.redirect(authUrl);
};
