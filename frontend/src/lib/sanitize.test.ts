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

  it("PR #186 review (P2-1): FormData は空オブジェクト化せずファイルパートを保持する", () => {
    const fd = new FormData();
    fd.append("file", new File(["a,b\n1,2"], "friends.csv", { type: "text/csv" }));

    const result = sanitizeNullBytes(fd) as FormData;

    expect(result instanceof FormData).toBe(true);
    expect(result.get("file")).toBeInstanceOf(File);
    expect((result.get("file") as File).name).toBe("friends.csv");
  });

  it("セキュリティレビュー指摘: FormData に同居するテキストフィールドの NULL バイトは除去し、file パートは無傷で通す", () => {
    const fd = new FormData();
    fd.append("file", new File(["content"], "a.csv", { type: "text/csv" }));
    fd.append("purpose", `line_upload${NULL_BYTE}`);
    fd.append("owner_id", "123");

    const result = sanitizeNullBytes(fd) as FormData;

    expect(result.get("purpose")).toBe("line_upload");
    expect(result.get("owner_id")).toBe("123");
    expect(result.get("file")).toBeInstanceOf(File);
    expect((result.get("file") as File).name).toBe("a.csv");
  });

  it("PR #186 review (P2-1): Blob/File 単体もそのまま返す", () => {
    const file = new File(["content"], "a.csv", { type: "text/csv" });
    const blob = new Blob(["content"], { type: "text/plain" });

    expect(sanitizeNullBytes(file)).toBe(file);
    expect(sanitizeNullBytes(blob)).toBe(blob);
  });
});
