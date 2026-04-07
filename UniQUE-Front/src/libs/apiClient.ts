/* eslint-disable @typescript-eslint/no-explicit-any */
import { getAllCookies } from "./cookies";
import { toSnakecase } from "./snakeCamelUtil";

const BASE = process.env.RESOURCE_API_URL;
const CLIENT_BASE = process.env.NEXT_PUBLIC_RESOURCE_API_URL;
const COOKIE_NAME = "session_jwt";

/** API レスポンスの拡張型 */
export interface ExtendedResponse extends Response {
  /** エラー情報が含まれている場合 */
  errorDetails?: {
    errorCode: string;
    errorMessage: string;
    timestamp: number;
    path: string;
  };
}

const clientComponentCheck =
  typeof window === "undefined" ? "server" : "client";

/** Cookie から JWT トークンを取得する */
async function getJwtToken(): Promise<string | null> {
  if (typeof window === "undefined") {
    // Server side
    try {
      const { cookies } = await import("next/headers");
      return (await cookies()).get(COOKIE_NAME)?.value || null;
    } catch {
      return null;
    }
  } else {
    // Client side – Cookie は httpOnly なので JavaScript からは取得できない
    // credentials: "include" で自動送信される
    return null;
  }
}

/** 共通ヘッダーを構築 (JWT Authorization 含む) */
async function buildHeaders(
  headers: HeadersInit | undefined,
): Promise<Record<string, string>> {
  const base: Record<string, string> = { "Content-Type": "application/json" };
  if (headers) {
    if (headers instanceof Headers) {
      headers.forEach((v, k) => (base[k] = v));
    } else {
      Object.assign(base, headers as Record<string, string>);
    }
  }
  const jwt = await getJwtToken();
  if (jwt) {
    base.Authorization = `Bearer ${jwt}`;
  }
  return base;
}

export const apiFetch = async (path: string, init: RequestInit = {}) => {
  const cookie = await getAllCookies();
  const headers = await buildHeaders(init.headers);
  const res = await fetch(
    `${clientComponentCheck === "server" ? BASE : CLIENT_BASE}${path}`,
    {
      ...init,
      headers: {
        ...headers,
        cookie,
      },
      credentials: "include",
    },
  );

  return res as ExtendedResponse;
};

const jsonStringifySafe = (obj: any) =>
  JSON.stringify(obj, (_k, v) => (typeof v === "bigint" ? Number(v) : v));

export const stringJsonParseSafe = (str: string) =>
  // numberの上限を超える場合はbigintとして扱う
  JSON.parse(str, (_k, v) => {
    if (typeof v === "string" && /^\d+$/.test(v)) {
      const num = Number(v);
      if (num > Number.MAX_SAFE_INTEGER) {
        return BigInt(v);
      }
    }
    return v;
  });

export const apiGet = (path: string, init?: RequestInit) =>
  apiFetch(path, { method: "GET", ...init });

export const apiPost = (path: string, body?: any, init?: RequestInit) =>
  apiFetch(path, {
    method: "POST",
    body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
    ...init,
  });

export const apiPut = (path: string, body?: any, init?: RequestInit) =>
  apiFetch(path, {
    method: "PUT",
    body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
    ...init,
  });

export const apiPatch = (path: string, body?: any, init?: RequestInit) =>
  apiFetch(path, {
    method: "PATCH",
    body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
    ...init,
  });

export const apiDelete = (path: string, init?: RequestInit) =>
  apiFetch(path, { method: "DELETE", ...init });

export const createApiClient = (base?: string) => {
  const resolvedBase = base ?? BASE;
  const apiFetchWithBase = async (path: string, init: RequestInit = {}) => {
    const cookie = await getAllCookies();
    const headers = await buildHeaders(init.headers);
    const res = await fetch(`${resolvedBase}${path}`, {
      ...init,
      headers: {
        ...headers,
        cookie,
      },
      credentials: "include",
    });

    return res as ExtendedResponse;
  };

  return {
    get: (path: string, init?: RequestInit) =>
      apiFetchWithBase(path, { method: "GET", ...init }),
    post: (path: string, body?: any, init?: RequestInit) =>
      apiFetchWithBase(path, {
        method: "POST",
        body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
        ...init,
      }),
    put: (path: string, body?: any, init?: RequestInit) =>
      apiFetchWithBase(path, {
        method: "PUT",
        body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
        ...init,
      }),
    patch: (path: string, body?: any, init?: RequestInit) =>
      apiFetchWithBase(path, {
        method: "PATCH",
        body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
        ...init,
      }),
    delete: (path: string, init?: RequestInit) =>
      apiFetchWithBase(path, { method: "DELETE", ...init }),
  };
};
