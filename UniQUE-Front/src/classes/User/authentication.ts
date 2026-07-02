"use server";

import "server-only";
import { cookies, headers } from "next/headers";
import { Session } from "@/classes/Session";
import { AuthServerErrors } from "@/errors/auth-server-errors";
import { AuthenticationErrors } from "@/errors/authentication-errors";
import { UniQUE_Error } from "@/errors/base";
import { FrontendErrors } from "@/errors/frontend-errors";
import { toCamelcase } from "@/libs/snakeCamelUtil";
import { createApiClient } from "../../libs/apiClient";
import { ClearSessionCookie, SetCookie } from "../../libs/cookies";
import { ParseJwt } from "../../libs/jwt";
import { getRealIPAddress } from "../../libs/request";
import { UserStatus } from "../types/User";
import { User } from ".";

export interface Credentials {
  username?: string;
  password?: string;
  code?: string;
  is_remember?: boolean;
}

enum AuthenticationType {
  Password = "password",
  TOTP = "totp",
}

interface AuthenticationRequest {
  code?: string;
  ipAddress: string;
  password?: string;
  is_remember: boolean;
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
  sessionJwt: string;
}

export const AuthenticationRequest = async (
  credentials: Credentials,
): Promise<AuthenticationResponse | UniQUE_Error | Error> => {
  try {
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
      is_remember: credentials.is_remember || false,
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

    SetCookie("session_jwt", data.sessionJwt);
    if (data.requireMfa) {
      if (data.mfaType) SetCookie("signin_mfa_type", data.mfaType.toString());
      else throw FrontendErrors.InvalidInput;
    }

    return data;
  } catch (e) {
    if (e instanceof UniQUE_Error || e instanceof Error) {
      return e;
    }
    return new Error(e as string);
  }
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

export interface MigrateRequest {
  displayName: string;
  password: string;
  birthdate: string;
  email: string;
  external_email: string;
}

export const MigrateRequest = async (request: MigrateRequest) => {
  const { email, external_email, displayName, password, birthdate } = request;
  const period = email.split("@")[0].split(".")[0];
  const username = email.split("@")[0].split(".").slice(1).join("."); // periodを除いた部分をusernameとして使用
  let internalEmail = email;
  if (period) {
    internalEmail = `${period.toUpperCase()}.${username}@uniproject.jp`;
  }

  const verifyRes = await fetch(
    `${process.env.GAS_MIGRATE_API_URL}?external_email=${encodeURIComponent(
      external_email,
    )}&internal_email=${encodeURIComponent(internalEmail)}`,
    { method: "GET" },
  );
  const verifyData = await verifyRes.json();
  if (!verifyRes.ok || verifyData.status !== 200) {
    return AuthenticationErrors.MigrationError;
  }

  const joinedAt = new Date(verifyData.joined_at);

  const res = await User.create(
    {
      email,
      customId: username,
      externalEmail: external_email,
      affiliationPeriod: period.toUpperCase(),
      status: UserStatus.ACTIVE,
      profile: {
        displayName,
        joinedAt: joinedAt.toISOString(),
        birthdate,
      },
    },
    password,
  );
  return res;
};
