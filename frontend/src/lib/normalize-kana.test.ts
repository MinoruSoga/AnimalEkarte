import { describe, it, expect } from "vitest";
import { normalizeKana } from "./normalize-kana";

describe("normalizeKana", () => {
  it("ひらがなは変換しない", () => {
    expect(normalizeKana("あいうえお")).toBe("あいうえお");
  });

  it("カタカナをひらがなに変換する", () => {
    expect(normalizeKana("アイウエオ")).toBe("あいうえお");
  });

  it("ぴ はひらがなのまま (ピーター の name_kana ヒット用)", () => {
    expect(normalizeKana("ぴ")).toBe("ぴ");
  });

  it("ピ をひらがな ぴ に変換 (ピーター 検索用)", () => {
    expect(normalizeKana("ピ")).toBe("ぴ");
  });

  it("ピーター 全体をひらがなに変換", () => {
    expect(normalizeKana("ピーター")).toBe("ぴーたー");
  });

  it("カタカナひらがな混在: ポチたろう", () => {
    expect(normalizeKana("ポチたろう")).toBe("ぽちたろう");
  });

  it("漢字は変換しない", () => {
    expect(normalizeKana("山田花子")).toBe("山田花子");
  });

  it("漢字とカタカナ混在: 山田ハナコ", () => {
    expect(normalizeKana("山田ハナコ")).toBe("山田はなこ");
  });

  it("小文字カタカナ ァィゥェォ をひらがなに変換", () => {
    expect(normalizeKana("ァィゥェォ")).toBe("ぁぃぅぇぉ");
  });

  it("ASCII は変換しない", () => {
    expect(normalizeKana("abc123")).toBe("abc123");
  });

  it("空文字は空文字のまま", () => {
    expect(normalizeKana("")).toBe("");
  });

  it("ヴ → ゔ", () => {
    expect(normalizeKana("ヴ")).toBe("ゔ");
  });

  it("ヵヶ → ゕゖ", () => {
    expect(normalizeKana("ヵヶ")).toBe("ゕゖ");
  });
});
