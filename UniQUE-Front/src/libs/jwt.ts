import {
  importJWK,
  type JWTHeaderParameters,
  type JWTPayload,
  type JWTVerifyGetKey,
  jwtVerify,
} from "jose";
import { baseLogger } from "./logger";

const AUTH_ISSUER = process.env.AUTH_API_URL;
const JWKS_CACHE_TTL = 1000 * 60 * 60; // 1時間

export const ParseJwt = (token: string): { [key: string]: unknown } => {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) => {
          return `%${`00${c.charCodeAt(0).toString(16)}`.slice(-2)}`;
        })
        .join(""),
    );
    return JSON.parse(jsonPayload);
  } catch (error) {
    console.error("Failed to parse JWT:", error);
    return {};
  }
};

// JWKS キャッシュ
let jwksCache: {
  data: { keys: JWKSResponse["keys"] };
  timestamp: number;
} | null = null;

interface JWKSResponse {
  keys: Array<{
    kty: string;
    use: string;
    alg: string;
    kid: string;
    n: string;
    e: string;
  }>;
}

/**
 * JWKS エンドポイントから JSON Web Key Set を取得
 */
async function fetchJWKS(): Promise<JWKSResponse> {
  // キャッシュが有効かチェック
  if (jwksCache && Date.now() - jwksCache.timestamp < JWKS_CACHE_TTL) {
    return jwksCache.data;
  }

  try {
    const jwksUrl = `${AUTH_ISSUER}/.well-known/jwks.json`;
    const response = await fetch(jwksUrl);
    if (!response.ok) {
      throw new Error(`Failed to fetch JWKS: ${response.statusText}`);
    }
    const jwks: JWKSResponse = await response.json();
    jwksCache = {
      data: jwks,
      timestamp: Date.now(),
    };
    return jwks;
  } catch (error) {
    console.error("Error fetching JWKS:", error);
    throw error;
  }
}

/**
 * JWT トークンを検証する（JWKS から鍵を取得）
 * @param token JWT トークン文字列
 * @returns 検証結果 { valid: boolean, payload?: JWTPayload }
 */
export async function VerifyJwt(token: string): Promise<{
  valid: boolean;
  payload?: JWTPayload;
}> {
  try {
    // JWKS を取得
    const jwks = await fetchJWKS();

    // JWKS から鍵を探す関数
    const key: JWTVerifyGetKey = async (
      protectedHeader: JWTHeaderParameters,
    ) => {
      const jwk = jwks.keys.find((k) => k.kid === protectedHeader.kid);
      if (!jwk) {
        throw new Error(
          `Key with kid ${protectedHeader.kid} not found in JWKS`,
        );
      }
      return importJWK(jwk);
    };

    // JWT を検証
    const verified = await jwtVerify(token, key);
    const sub = verified.payload.sub;
    // sid_を外す
    const sid =
      typeof sub === "string" && sub.startsWith("SID_") ? sub.slice(4) : sub;
    const sessionValidateResponse = await fetch(
      `${AUTH_ISSUER}/internal/session_verify?jti=${sid}`,
    );
    if (!sessionValidateResponse.ok) {
      throw new Error(
        `Session validation failed: ${sessionValidateResponse.statusText}`,
      );
    }
    const data = await sessionValidateResponse.json();
    if (!data.valid) {
      throw new Error("Session is not valid");
    }
    return { valid: true, payload: verified.payload };
  } catch (error) {
    const logger = baseLogger.getLogger({ function: "VerifyJwt" });
    logger.error(error);
    return { valid: false };
  }
}
