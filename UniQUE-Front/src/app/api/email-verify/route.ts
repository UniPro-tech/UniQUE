import { type NextRequest, NextResponse } from "next/server";
import { User } from "@/classes/User";

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const code = searchParams.get("code");

  if (!code) {
    return NextResponse.json({ error: "Code is required" }, { status: 400 });
  }

  try {
    const result = await User.emailVerify(code);

    if (!result) {
      return NextResponse.json({ valid: false, type: null }, { status: 200 });
    }

    if ("error" in result) {
      // エラーメッセージを含めて返す
      return NextResponse.json(result, { status: 200 });
    }

    return NextResponse.json(result, { status: 200 });
  } catch (error) {
    console.error("Email verification error:", error);
    return NextResponse.json({ error: "Verification failed" }, { status: 500 });
  }
}
