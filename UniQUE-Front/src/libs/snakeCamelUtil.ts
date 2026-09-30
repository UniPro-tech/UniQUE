const unsafePropertyNames = new Set(["__proto__", "constructor", "prototype"]);

const transformObject = <T>(
  obj: object,
  transformKey: (key: string) => string,
  transformValue: (value: unknown) => unknown,
): T => {
  const entries: [string, unknown][] = [];

  for (const [key, value] of Object.entries(obj)) {
    const transformedKey = transformKey(key);

    // Prevent untrusted input from modifying an object's prototype chain.
    if (
      unsafePropertyNames.has(key) ||
      unsafePropertyNames.has(transformedKey)
    ) {
      continue;
    }

    entries.push([transformedKey, transformValue(value)]);
  }

  return Object.fromEntries(entries) as T;
};

export const toCamelcase = <T>(obj: unknown): T => {
  if (Array.isArray(obj)) {
    return obj.map((item) => toCamelcase(item)) as unknown as T;
  } else if (obj !== null && typeof obj === "object") {
    return transformObject<T>(
      obj,
      (key) =>
        key.replace(/_([a-z])/g, (_, letter: string) => letter.toUpperCase()),
      toCamelcase,
    );
  }
  return obj as T;
};

export const toSnakecase = <T>(obj: unknown): T => {
  if (Array.isArray(obj)) {
    return obj.map((item) => toSnakecase(item)) as unknown as T;
  } else if (obj !== null && typeof obj === "object") {
    return transformObject<T>(
      obj,
      (key) => key.replace(/([A-Z])/g, "_$1").toLowerCase(),
      toSnakecase,
    );
  }
  return obj as T;
};
