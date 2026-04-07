/**
 * POST /api/password-reset/request
 * Request password reset by email
 */

const AUTH_API_URL = process.env.AUTH_API_URL;

export const POST = async (request: Request) => {
  try {
    const body = await request.json();
    const { email } = body;

    if (!email) {
      return Response.json(
        { ok: false, error: "Email is required" },
        { status: 400 },
      );
    }

    const response = await fetch(`${AUTH_API_URL}/password-reset-request`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        email,
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
    console.error("Password reset request error:", error);
    return Response.json(
      {
        ok: false,
        error: "Failed to request password reset",
      },
      { status: 500 },
    );
  }
};
