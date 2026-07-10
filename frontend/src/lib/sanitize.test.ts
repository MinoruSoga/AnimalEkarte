import { describe, expect, it } from "vitest";
import { sanitizeNullBytes } from "./sanitize";

const NULL_BYTE = String.fromCharCode(0);

describe("sanitizeNullBytes（BUG-067 / R-F20）", () => {
  it("文字列中の NULL バイトを除去する", () => {
    const input = `山田${NULL_BYTE}太郎`;

    expect(sanitizeNullBytes(input)).toBe("山田太郎");
  });

  it("NULL バイトを含まない文字列はそのまま返す", () => {
    expect(sanitizeNullBytes("山田太郎")).toBe("山田太郎");
  });

  it("配列内の各要素を再帰的にサニタイズする", () => {
    const input = [`ポチ${NULL_BYTE}`, "タマ"];

    expect(sanitizeNullBytes(input)).toEqual(["ポチ", "タマ"]);
  });

  it("ネストしたオブジェクトのフィールドを再帰的にサニタイズする", () => {
    const input = {
      name: `山田${NULL_BYTE}花子`,
      pets: [{ name: `ポチ${NULL_BYTE}` }],
      requestText: `爪切り${NULL_BYTE}希望`,
    };

    expect(sanitizeNullBytes(input)).toEqual({
      name: "山田花子",
      pets: [{ name: "ポチ" }],
      requestText: "爪切り希望",
    });
  });

  it("null / number / boolean はそのまま返す", () => {
    expect(sanitizeNullBytes(null)).toBeNull();
    expect(sanitizeNullBytes(42)).toBe(42);
    expect(sanitizeNullBytes(true)).toBe(true);
  });
});
