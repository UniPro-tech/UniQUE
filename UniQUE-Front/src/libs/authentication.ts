import { cookies, headers } from "next/headers";
import { Session } from "@/classes/Session";
import { AuthenticationErrors } from "@/errors/AuthenticationErrors";
import { AuthServerErrors } from "@/errors/AuthServerErrors";
import { FrontendErrors } from "@/errors/FrontendErrors";
import { toCamelcase } from "@/libs/snakeCamelUtil";
import { createApiClient } from "./apiClient";
import { ClearSessionCookie } from "./cookies";
import { ParseJwt } from "./jwt";
import { getRealIPAddress } from "./request";

export interface Credentials {
  username?: string;
  password?: string;
  code?: string;
  remember?: boolean;
}

enum AuthenticationType {
  Password = "password",
  TOTP = "totp",
}

interface AuthenticationRequest {
  code?: string;
  ipAddress: string;
  password?: string;
  remember: boolean;
  type: AuthenticationType;
  userAgent: string;
  username: string;
}

enum MultifactorAuthenticationType {
  TOTP = "totp",
}

export interface AuthenticationResponse {
  mfaType?: MultifactorAuthenticationType[];
  requireMfa?: boolean;
  sessionJwt?: string;
}

export const AuthenticationRequest = async (
  credentials: Credentials,
): Promise<AuthenticationResponse> => {
  const { username, password, code } = credentials;

  const ipAddress = await getRealIPAddress();
  const userAgent = (await headers()).get("user-agent") || "Unknown";

  const requestBody: AuthenticationRequest = {
    type: password ? AuthenticationType.Password : AuthenticationType.TOTP,
    username: username || "",
    password,
    code,
    ipAddress,
    userAgent,
    remember: credentials.remember || false,
  };

  const apiClient = createApiClient(process.env.AUTH_API_URL);
  const response = await apiClient.post(
    "/internal/authentication",
    requestBody,
  );

  if (!response.ok) {
    switch (response.status) {
      case 401: {
        const errorData = await response.json();
        if (errorData.reason === "invalid_credentials") {
          throw AuthenticationErrors.InvalidCredentials;
        } else if (errorData.reason === "user_inactive") {
          throw AuthenticationErrors.AccountLocked;
        }
        throw AuthenticationErrors.InvalidCredentials;
      }
      case 400:
        throw FrontendErrors.InvalidInput;
      default:
        throw AuthServerErrors.InternalServerError;
    }
  }

  const data = toCamelcase<AuthenticationResponse>(await response.json());

  return data;
};

export const Logout = async () => {
  const cookieStore = await cookies();
  const sessionJwt = cookieStore.get("session_jwt")?.value;
  if (!sessionJwt) {
    return;
  }
  const sessionJwtPayload = ParseJwt(sessionJwt);
  const sub = sessionJwtPayload.sub;
  if (typeof sub !== "string") {
    return;
  }
  const sid = sub.startsWith("SID_") ? sub.slice(4) : sub;

  Session.deleteById(sid);

  ClearSessionCookie();
};
