import { type NextRequest, NextResponse } from "next/server";
import { apiGet } from "@/libs/apiClient";

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const code = searchParams.get("code");

  if (!code) {
    return NextResponse.json({ error: "Code is required" }, { status: 400 });
  }

  try {
    const res = await apiGet(
      `/internal/users/email_verify/${encodeURIComponent(code)}`,
    );

    if (!res.ok) {
      const errorData = await res.json().catch(() => ({}));
      return NextResponse.json(
        { error: errorData.error || "Code not found" },
        { status: res.status },
      );
    }

    const data = await res.json();
    return NextResponse.json(data, { status: 200 });
  } catch (error) {
    console.error("Email verify code lookup error:", error);
    return NextResponse.json(
      { error: "Verification lookup failed" },
      { status: 500 },
    );
  }
}
