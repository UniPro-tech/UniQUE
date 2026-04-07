"use server";

import { Session } from "@/classes/Session";

export interface UpdateProfileData {
  displayName?: string;
  bio?: string;
  websiteUrl?: string;
  twitterHandle?: string;
  birthdateVisible?: boolean;
}

export async function updateProfile(data: UpdateProfileData) {
  try {
    const session = await Session.getCurrent();

    const user = await session?.getUser();

    if (!user) {
      return {
        success: false,
        error: "ユーザーが見つかりませんでした",
      };
    }

    // プロフィール情報を更新
    if (data.displayName !== undefined)
      user.profile.displayName = data.displayName;
    if (data.bio !== undefined) user.profile.bio = data.bio;
    if (data.bio === "") user.profile.bio = null; // 空文字はnullに変換
    if (data.websiteUrl !== undefined)
      user.profile.websiteUrl = data.websiteUrl;
    if (data.websiteUrl === "") user.profile.websiteUrl = null; // 空文字はnullに変換
    if (data.twitterHandle !== undefined)
      user.profile.twitterHandle = data.twitterHandle;
    if (data.twitterHandle === "") user.profile.twitterHandle = null; // 空文字はnullに変換
    if (data.birthdateVisible !== undefined)
      user.profile.birthdateVisible = data.birthdateVisible;

    await user.save();

    return {
      success: true,
      message: "プロフィールを更新しました",
      profile: user.profile.toJson(),
    };
  } catch (err) {
    console.error("プロフィールの更新に失敗:", err);
    return {
      success: false,
      error: "プロフィールの更新に失敗しました",
    };
  }
}
