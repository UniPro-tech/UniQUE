/**
 * Frontend structured logger.
 *
 * Mirrors the backend schema (see docs/logging.md): JSON entries with a
 * `component`, snake_case `msg`, correlation IDs and masked sensitive fields.
 */

export type LogLevel = "debug" | "info" | "warn" | "error";

export interface LogContext {
  requestId: string;
  traceId: string;
}

export const LOG_HEADERS = {
  requestId: "X-Request-ID",
  traceId: "X-Trace-ID",
  traceparent: "traceparent",
} as const;

const COMPONENT = "front";
const SENSITIVE_KEYS = new Set([
  "access_token",
  "refresh_token",
  "id_token",
  "token",
  "secret",
  "client_secret",
  "bot_token",
  "password",
  "authorization",
  "api_key",
  "cookie",
  "set-cookie",
  "jti",
  "sid",
  "session_id",
]);

const REDACTED = "[REDACTED]";
const SAFE_KEY = /^[a-z0-9_]{1,64}$/i;
const FORBIDDEN_KEYS = new Set(["__proto__", "constructor", "prototype"]);

function isSafeKey(key: string): boolean {
  return SAFE_KEY.test(key) && !FORBIDDEN_KEYS.has(key);
}

const RESERVED_KEYS = new Set([
  "timestamp",
  "level",
  "component",
  "msg",
  "message",
  "request_id",
  "trace_id",
  "error",
]);

let currentContext: LogContext | null = null;
let pageContext: LogContext | null = null;

function randomHex(bytes: number): string {
  const buffer = new Uint8Array(bytes);
  if (typeof globalThis.crypto?.getRandomValues === "function") {
    globalThis.crypto.getRandomValues(buffer);
  } else {
    for (let i = 0; i < buffer.length; i += 1) {
      buffer[i] = Math.floor(Math.random() * 256);
    }
  }
  return Array.from(buffer, (value) =>
    value.toString(16).padStart(2, "0"),
  ).join("");
}

export function newRequestContext(): LogContext {
  return { requestId: randomHex(16), traceId: randomHex(16) };
}

export function newClientContext(): LogContext {
  if (!pageContext) {
    pageContext = newRequestContext();
    pageContext.requestId = randomHex(16);
  }
  return { requestId: randomHex(16), traceId: pageContext.traceId };
}

export function traceparent(traceId: string): string {
  return `00-${traceId}-${randomHex(8)}-01`;
}

function traceIdFromTraceparent(value: string): string {
  const parts = value.split("-");
  if (parts.length >= 4 && /^[0-9a-f]{32}$/i.test(parts[1])) {
    return parts[1];
  }
  return "";
}

export function requestContextFromHeaders(headers: {
  get(name: string): string | null;
}): LogContext {
  const generated = newRequestContext();
  const traceId =
    headers.get(LOG_HEADERS.traceId) ||
    traceIdFromTraceparent(headers.get(LOG_HEADERS.traceparent) ?? "");
  return {
    requestId: headers.get(LOG_HEADERS.requestId) || generated.requestId,
    traceId: traceId || generated.traceId,
  };
}

export function getLogContext(): LogContext {
  if (currentContext) {
    return currentContext;
  }
  if (!pageContext) {
    pageContext = newRequestContext();
  }
  return pageContext;
}

export function setLogContext(context: LogContext | null): void {
  currentContext = context;
}

export async function withLogContext<T>(
  context: LogContext,
  fn: () => Promise<T> | T,
): Promise<T> {
  const previous = currentContext;
  currentContext = context;
  try {
    return await fn();
  } finally {
    currentContext = previous;
  }
}

function redactString(value: string): string {
  return value
    .replace(
      /"(access_token|refresh_token|id_token|token|secret|client_secret|bot_token|password|authorization|api_key|jti|sid|session_id)"\s*:\s*"[^"]*"/gi,
      '"$1":"[REDACTED]"',
    )
    .replace(/\b(Bearer|Bot)\s+[A-Za-z0-9._~+/=-]+/gi, "$1 [REDACTED]");
}

function sanitize(value: unknown, depth = 0): unknown {
  if (depth > 6) {
    return "[TRUNCATED]";
  }
  if (value instanceof Error) {
    return {
      name: value.name,
      message: redactString(value.message),
      ...(process.env.NODE_ENV === "production" ? {} : { stack: value.stack }),
    };
  }
  if (typeof value === "string") {
    return redactString(value);
  }
  if (Array.isArray(value)) {
    return value.map((item) => sanitize(item, depth + 1));
  }
  if (value && typeof value === "object") {
    const result: Record<string, unknown> = {};
    for (const [key, nested] of Object.entries(value)) {
      if (!isSafeKey(key)) {
        continue;
      }
      if (SENSITIVE_KEYS.has(key.toLowerCase())) {
        result[key] = REDACTED;
      } else {
        result[key] = sanitize(nested, depth + 1);
      }
    }
    return result;
  }
  return value;
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return (
    typeof value === "object" &&
    value !== null &&
    Object.getPrototypeOf(value) === Object.prototype
  );
}

function deriveEvent(message: string, fallback: string): string {
  let slug = "";
  let pendingSeparator = false;
  for (const char of message.toLowerCase()) {
    const isAlphanumeric =
      (char >= "a" && char <= "z") || (char >= "0" && char <= "9");
    if (isAlphanumeric) {
      if (pendingSeparator && slug.length > 0) {
        slug += "_";
      }
      slug += char;
      pendingSeparator = false;
    } else {
      pendingSeparator = true;
    }
    if (slug.length >= 60) {
      break;
    }
  }
  while (slug.endsWith("_")) {
    slug = slug.slice(0, -1);
  }
  return slug || fallback;
}

function normalizeArgs(args: unknown[]): {
  message: string;
  fields: Record<string, unknown>;
} {
  const messages: string[] = [];
  const extras: unknown[] = [];
  let error: unknown;

  args.forEach((arg, index) => {
    if (arg instanceof Error) {
      if (error === undefined) {
        error = sanitize(arg);
      } else {
        extras.push(sanitize(arg));
      }
    } else if (typeof arg === "string") {
      if (index === 0) {
        messages.push(arg);
      } else {
        extras.push(arg);
      }
    } else if (isPlainObject(arg)) {
      extras.push(sanitize(arg));
    } else {
      extras.push(sanitize(arg));
    }
  });

  const fields: Record<string, unknown> = {};
  if (error !== undefined) {
    fields.error = error;
  }
  if (extras.length === 1 && isPlainObject(extras[0])) {
    // Keep caller-provided fields at the top level, but never overwrite the
    // schema's reserved keys.
    for (const [key, value] of Object.entries(extras[0])) {
      if (!RESERVED_KEYS.has(key) && isSafeKey(key)) {
        fields[key] = value;
      }
    }
  } else if (extras.length === 1) {
    fields.detail = extras[0];
  } else if (extras.length > 1) {
    fields.details = extras;
  }

  return { message: messages.join(" "), fields };
}

function emit(level: LogLevel, fallbackEvent: string, args: unknown[]): void {
  const { message, fields } = normalizeArgs(args);
  const context = getLogContext();
  const entry = {
    timestamp: new Date().toISOString(),
    level,
    component: COMPONENT,
    msg: deriveEvent(message, fallbackEvent),
    ...(message ? { message } : {}),
    request_id: context.requestId,
    trace_id: context.traceId,
    ...fields,
  };
  const line = JSON.stringify(entry);

  switch (level) {
    case "error":
      console.error(line);
      break;
    case "warn":
      console.warn(line);
      break;
    case "debug":
      if (process.env.NODE_ENV !== "production") {
        console.debug(line);
      }
      break;
    default:
      console.log(line);
  }
}

/** Console-compatible structured logger. */
export const logger = {
  debug: (...args: unknown[]) => emit("debug", "front_debug", args),
  info: (...args: unknown[]) => emit("info", "front_info", args),
  log: (...args: unknown[]) => emit("info", "front_info", args),
  warn: (...args: unknown[]) => emit("warn", "front_warn", args),
  error: (...args: unknown[]) => emit("error", "front_error", args),
};
