/**
 * POST /api/password-reset/confirm
 * Confirm password reset with code and new password
 */

const AUTH_API_URL = process.env.AUTH_API_URL;

export const POST = async (request: Request) => {
  try {
    const body = await request.json();
    const { code, password } = body;

    if (!code || !password) {
      return Response.json(
        { ok: false, error: "Code and password are required" },
        { status: 400 },
      );
    }

    const response = await fetch(`${AUTH_API_URL}/password-reset-confirm`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        code,
        password,
      }),
    });

    if (!response.ok) {
      return Response.json(
        {
          ok: false,
          error: `Auth API Error: ${response.status}`,
        },
        { status: response.status },
      );
    }

    const data = await response.json();
    return Response.json({
      ok: true,
      message: data.message,
    });
  } catch (error) {
    console.error("Password reset confirm error:", error);
    return Response.json(
      {
        ok: false,
        error: "Failed to confirm password reset",
      },
      { status: 500 },
    );
  }
};
