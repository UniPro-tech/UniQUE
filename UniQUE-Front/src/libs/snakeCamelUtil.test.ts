/// <reference types="bun" />

import { describe, expect, test } from "bun:test";
import { toCamelcase, toSnakecase } from "./snakeCamelUtil";

type Converter = <T>(obj: unknown) => T;

describe("snakeCamelUtil", () => {
  test("converts nested keys and arrays to camel case", () => {
    expect(
      toCamelcase<Record<string, unknown>>({
        user_name: "test user",
        profile_data: [{ created_at: "2026-09-30" }],
      }),
    ).toEqual({
      userName: "test user",
      profileData: [{ createdAt: "2026-09-30" }],
    });
  });

  test("converts nested keys and arrays to snake case", () => {
    expect(
      toSnakecase<Record<string, unknown>>({
        userName: "test user",
        profileData: [{ createdAt: "2026-09-30" }],
      }),
    ).toEqual({
      user_name: "test user",
      profile_data: [{ created_at: "2026-09-30" }],
    });
  });

  test.each<Converter>([toCamelcase, toSnakecase])(
    "does not copy properties that can alter the prototype chain",
    (convert) => {
      const maliciousInput = JSON.parse(
        '{"__proto__":{"polluted":true},"constructor":{"prototype":{"polluted":true}},"prototype":{"polluted":true},"safe":1}',
      );

      const result = convert<Record<string, unknown>>(maliciousInput);

      expect(result).toEqual({ safe: 1 });
      expect(Object.hasOwn(result, "__proto__")).toBe(false);
      expect(Object.hasOwn(result, "constructor")).toBe(false);
      expect(Object.hasOwn(result, "prototype")).toBe(false);
      expect(({} as { polluted?: boolean }).polluted).toBeUndefined();
    },
  );
});
