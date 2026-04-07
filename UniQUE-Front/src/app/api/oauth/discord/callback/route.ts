import type { NextRequest } from "next/server";
import { ExternalIdentity } from "@/classes/ExternalIdentity";
import { Session } from "@/classes/Session";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

// guild invite and role assignment are performed by backend API

/**
 * Discord OAuthコールバックエンドポイント
 * Discordからのリダイレクトを受け取り、外部アイデンティティを登録する
 */
export const GET = async (request: NextRequest) => {
  const searchParams = request.nextUrl.searchParams;
  const code = searchParams.get("code");
  const error = searchParams.get("error");
  const stateParam = searchParams.get("state");

  // stateをデコードしてfromとcodeを取得
  let from = "settings";
  let emailVerifyCode: string | undefined;
  try {
    if (stateParam) {
      const stateObj = JSON.parse(
        Buffer.from(stateParam, "base64").toString("utf-8"),
      );
      from = stateObj.from || "settings";
      emailVerifyCode = stateObj.code;
    }
  } catch (e) {
    console.error("Failed to parse state:", e);
  }

  // エラーチェック
  if (error) {
    const sanitizedError = String(error).replace(/\r|\n/g, "");
    console.error("Discord OAuth error:", sanitizedError);
    let redirectPath: string;
    if (from === "email_verify" && emailVerifyCode) {
      redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_error=true`;
    } else if (from === "signup") {
      redirectPath = `/signup?completed=true&oauth=discord&status=error`;
    } else {
      redirectPath = `/dashboard/settings?oauth=Discord&status=error`;
    }
    return Response.redirect(new URL(redirectPath, request.nextUrl.origin));
  }

  if (!code) {
    console.error("Discord OAuth code is missing");
    let redirectPath: string;
    if (from === "email_verify" && emailVerifyCode) {
      redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_error=true`;
    } else if (from === "signup") {
      redirectPath = `/signup?completed=true&oauth=discord&status=error`;
    } else {
      redirectPath = `/dashboard/settings?oauth=Discord&status=error`;
    }
    return Response.redirect(new URL(redirectPath, request.nextUrl.origin));
  }

  // 環境変数の確認
  const clientId = process.env.DISCORD_CLIENT_ID;
  const clientSecret = process.env.DISCORD_CLIENT_SECRET;
  const redirectUri = process.env.DISCORD_REDIRECT_URI;

  if (!clientId || !clientSecret || !redirectUri) {
    console.error("Discord OAuth環境変数が設定されていません");
    let redirectPath: string;
    if (from === "email_verify" && emailVerifyCode) {
      redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_error=true`;
    } else if (from === "signup") {
      redirectPath = `/signup?completed=true&oauth=discord&status=error`;
    } else {
      redirectPath = `/dashboard/settings?oauth=Discord&status=error`;
    }
    return Response.redirect(new URL(redirectPath, request.nextUrl.origin));
  }

  try {
    // 1. アクセストークンを取得
    const tokenResponse = await fetch("https://discord.com/api/oauth2/token", {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: new URLSearchParams({
        client_id: clientId,
        client_secret: clientSecret,
        grant_type: "authorization_code",
        code,
        redirect_uri: redirectUri,
      }),
    });

    if (!tokenResponse.ok) {
      const tokenErr = await tokenResponse.text().catch(() => null);
      console.error("Failed to get Discord access token:", tokenErr);
      let redirectPath: string;
      if (from === "email_verify" && emailVerifyCode) {
        redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_error=true`;
      } else if (from === "signup") {
        redirectPath = `/signup?completed=true&oauth=discord&status=error`;
      } else {
        redirectPath = `/dashboard/settings?oauth=Discord&status=error`;
      }
      return Response.redirect(new URL(redirectPath, request.nextUrl.origin));
    }

    const tokenData = await tokenResponse.json();
    const {
      access_token,
      refresh_token,
      expires_in,
    }: {
      access_token: string;
      refresh_token: string;
      expires_in: number;
    } = tokenData;

    // 2. Discordユーザー情報を取得
    const userResponse = await fetch("https://discord.com/api/users/@me", {
      headers: {
        Authorization: `Bearer ${access_token}`,
      },
    });

    if (!userResponse.ok) {
      const userErr = await userResponse.text().catch(() => null);
      console.error("Failed to get Discord user info:", userErr);
      let redirectPath: string;
      if (from === "email_verify" && emailVerifyCode) {
        redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_error=true`;
      } else if (from === "signup") {
        redirectPath = `/signup?completed=true&oauth=discord&status=error`;
      } else {
        redirectPath = `/dashboard/settings?oauth=Discord&status=error`;
      }
      return Response.redirect(new URL(redirectPath, request.nextUrl.origin));
    }

    const discordUser = await userResponse.json();
    const discordUserId = discordUser.id;

    // 3. 外部アイデンティティをAPIに登録
    const expiresAt = new Date(Date.now() + expires_in * 1000).toISOString();

    const session = await Session.getCurrent();
    const user = await session?.getUser();

    try {
      if (user) {
        await user.addExternalIdentity({
          provider: "discord",
          externalUserId: discordUserId,
          accessToken: access_token,
          refreshToken: refresh_token,
          tokenExpiresAt: expiresAt,
          idToken: null,
        });
      } else if (from === "email_verify" && emailVerifyCode) {
        await ExternalIdentity.createByEmailVerificationCode(emailVerifyCode, {
          provider: "discord",
          externalUserId: discordUserId,
          accessToken: access_token,
          refreshToken: refresh_token,
          tokenExpiresAt: expiresAt,
          idToken: null,
        });
      } else {
        console.error("Could not determine user ID");
        return Response.redirect(
          new URL(
            `/signin?redirect=/dashboard/settings`,
            request.nextUrl.origin,
          ),
        );
      }
    } catch (e) {
      // 409 Conflict エラーの場合、既に連携されているアカウント
      if (e === ResourceApiErrors.ResourceAlreadyExists) {
        let redirectPath: string;
        if (from === "email_verify" && emailVerifyCode) {
          redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_error=conflict`;
        } else if (from === "signup") {
          redirectPath = `/signup?completed=true&oauth=discord&status=warning&reason=conflict`;
        } else {
          redirectPath = `/dashboard/settings?oauth=Discord&status=warning&reason=conflict`;
        }
        return Response.redirect(new URL(redirectPath, request.nextUrl.origin));
      }

      let redirectPath: string;
      if (from === "email_verify" && emailVerifyCode) {
        redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_error=true`;
      } else if (from === "signup") {
        redirectPath = `/signup?completed=true&oauth=discord&status=error`;
      } else {
        redirectPath = `/dashboard/settings?oauth=Discord&status=error`;
      }
      return Response.redirect(new URL(redirectPath, request.nextUrl.origin));
    }

    // guild invite and role assignment are handled by backend; nothing to do here

    // 7. 成功したらリダイレクト
    let redirectPath: string;
    if (from === "email_verify" && emailVerifyCode) {
      // メール認証ページに戻る
      redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_linked=true`;
    } else if (from === "signup") {
      redirectPath = `/signup?completed=true&oauth=discord&status=success`;
    } else {
      redirectPath = `/dashboard/settings?oauth=Discord&status=success`;
    }
    return Response.redirect(
      new URL(
        redirectPath,
        process.env.NEXT_PUBLIC_FRONTEND_URL || request.nextUrl.origin,
      ),
    );
  } catch (error) {
    console.error("Discord OAuth callback error:", error);
    let redirectPath: string;
    if (from === "email_verify" && emailVerifyCode) {
      redirectPath = `/email-verify?code=${encodeURIComponent(emailVerifyCode)}&discord_error=true`;
    } else if (from === "signup") {
      redirectPath = `/signup?completed=true&oauth=discord&status=error`;
    } else {
      redirectPath = `/dashboard/settings?oauth=Discord&status=error`;
    }
    return Response.redirect(
      new URL(
        redirectPath,
        process.env.NEXT_PUBLIC_FRONTEND_URL || request.nextUrl.origin,
      ),
    );
  }
};
