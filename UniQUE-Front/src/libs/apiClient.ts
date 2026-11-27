/* eslint-disable @typescript-eslint/no-explicit-any */
import { getAllCookies } from "./cookies";
import {
  LOG_HEADERS,
  type LogContext,
  logger,
  newClientContext,
  newRequestContext,
  traceparent,
  withLogContext,
} from "./logger";
import { toSnakecase } from "./snakeCamelUtil";

const BASE = process.env.RESOURCE_API_URL;
const CLIENT_BASE = process.env.NEXT_PUBLIC_RESOURCE_API_URL;
const COOKIE_NAME = "session_jwt";
const MAX_LOGGED_BODY = 2048;

/** API レスポンスの拡張型 */
export interface ExtendedResponse extends Response {
  /** リクエスト相関 ID */
  requestId?: string;
  /** トレース ID */
  traceId?: string;
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
      headers.forEach((v, k) => {
        base[k] = v;
      });
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

function resolveLogContext(): LogContext {
  // Server requests have no shared page context; generate a fresh pair.
  return typeof window === "undefined"
    ? newRequestContext()
    : newClientContext();
}

async function loggedFetch(
  url: string,
  init: RequestInit,
): Promise<ExtendedResponse> {
  const context = resolveLogContext();
  const method = (init.method ?? "GET").toUpperCase();
  const path = url.replace(BASE ?? "", "").replace(CLIENT_BASE ?? "", "");
  const startedAt = Date.now();

  return withLogContext(context, async () => {
    const headers = new Headers(init.headers);
    headers.set(LOG_HEADERS.requestId, context.requestId);
    headers.set(LOG_HEADERS.traceId, context.traceId);
    headers.set(LOG_HEADERS.traceparent, traceparent(context.traceId));

    logger.debug("api_request", { method, path });

    const response = await fetch(url, { ...init, headers });
    const latencyMs = Date.now() - startedAt;
    logger.debug("api_response", {
      method,
      path,
      status: response.status,
      latency_ms: latencyMs,
    });

    const extended = response as ExtendedResponse;
    extended.requestId = context.requestId;
    extended.traceId = context.traceId;

    if (!response.ok) {
      let detail: unknown;
      try {
        const body = await response.clone().text();
        const bounded = body.slice(0, MAX_LOGGED_BODY);
        try {
          detail = JSON.parse(bounded);
        } catch {
          detail = bounded;
        }
      } catch {
        detail = undefined;
      }
      logger.error("api_error", {
        method,
        path,
        status: response.status,
        latency_ms: latencyMs,
        response_body: detail,
      });
    }

    return extended;
  });
}

export const apiFetch = async (path: string, init: RequestInit = {}) => {
  const cookie = await getAllCookies();
  const headers = await buildHeaders(init.headers);
  const url = `${clientComponentCheck === "server" ? BASE : CLIENT_BASE}${path}`;
  return loggedFetch(url, {
    ...init,
    headers: {
      ...headers,
      cookie,
    },
    credentials: "include",
  });
};

const jsonStringifySafe = (obj: object) =>
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

export const apiPost = (path: string, body?: object, init?: RequestInit) =>
  apiFetch(path, {
    method: "POST",
    body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
    ...init,
  });

export const apiPut = (path: string, body?: object, init?: RequestInit) =>
  apiFetch(path, {
    method: "PUT",
    body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
    ...init,
  });

export const apiPatch = (path: string, body?: object, init?: RequestInit) =>
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
    const url = `${resolvedBase}${path}`;
    return loggedFetch(url, {
      ...init,
      headers: {
        ...headers,
        cookie,
      },
      credentials: "include",
    });
  };

  return {
    get: (path: string, init?: RequestInit) =>
      apiFetchWithBase(path, { method: "GET", ...init }),
    post: (path: string, body?: object, init?: RequestInit) =>
      apiFetchWithBase(path, {
        method: "POST",
        body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
        ...init,
      }),
    put: (path: string, body?: object, init?: RequestInit) =>
      apiFetchWithBase(path, {
        method: "PUT",
        body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
        ...init,
      }),
    patch: (path: string, body?: object, init?: RequestInit) =>
      apiFetchWithBase(path, {
        method: "PATCH",
        body: body ? jsonStringifySafe(toSnakecase(body)) : undefined,
        ...init,
      }),
    delete: (path: string, init?: RequestInit) =>
      apiFetchWithBase(path, { method: "DELETE", ...init }),
  };
};
