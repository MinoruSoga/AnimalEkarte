import { describe, it, expect } from "vitest";

import { parseNormalRange } from "./get-exam-type-fields";

describe("parseNormalRange", () => {
  describe("空入力 / 不正入力", () => {
    it.each([
      ["", "空文字"],
      ["   ", "空白のみ"],
      ["abc", "数値以外"],
      ["3.5", "単一数値（範囲ではない）"],
      ["3.5g/dL以上", "単位混入"],
      ["3.5以上 7.0", "以下なし"],
      ["3.5-", "片方欠落 (-)"],
      ["~7.0", "片方欠落 (~)"],
      ["<", "数値なし不等号"],
      [">", "数値なし不等号"],
      ["<>7.0", "二重不等号"],
      ["3.5から7.0", "未対応の区切り"],
    ])("%s (%s) は {} を返す", (input) => {
      expect(parseNormalRange(input)).toEqual({});
    });
  });

  describe("既存挙動の維持 — 範囲表記 [-~〜]", () => {
    it('"3.5-7.0" → refMin=3.5, refMax=7.0', () => {
      expect(parseNormalRange("3.5-7.0")).toEqual({ refMin: 3.5, refMax: 7.0 });
    });

    it('"3.5 - 7.0" (スペースあり) → refMin=3.5, refMax=7.0', () => {
      expect(parseNormalRange("3.5 - 7.0")).toEqual({ refMin: 3.5, refMax: 7.0 });
    });

    it('"3.5〜7.0" (全角チルダ) → refMin=3.5, refMax=7.0', () => {
      expect(parseNormalRange("3.5〜7.0")).toEqual({ refMin: 3.5, refMax: 7.0 });
    });

    it('"3.5~7.0" (半角チルダ) → refMin=3.5, refMax=7.0', () => {
      expect(parseNormalRange("3.5~7.0")).toEqual({ refMin: 3.5, refMax: 7.0 });
    });

    it('"70-110" (整数) → refMin=70, refMax=110', () => {
      expect(parseNormalRange("70-110")).toEqual({ refMin: 70, refMax: 110 });
    });

    it('前後空白あり "  3.5-7.0  " → refMin=3.5, refMax=7.0', () => {
      expect(parseNormalRange("  3.5-7.0  ")).toEqual({ refMin: 3.5, refMax: 7.0 });
    });

    it('"-2.0--1.0" (負の範囲) → refMin=-2.0, refMax=-1.0', () => {
      expect(parseNormalRange("-2.0--1.0")).toEqual({ refMin: -2.0, refMax: -1.0 });
    });
  });

  describe("不等号表記 — 上限のみ", () => {
    it('"<7.0" → refMax=7.0 のみ (refMin は undefined)', () => {
      expect(parseNormalRange("<7.0")).toEqual({ refMax: 7.0 });
    });

    it('"< 7.0" (スペースあり) → refMax=7.0', () => {
      expect(parseNormalRange("< 7.0")).toEqual({ refMax: 7.0 });
    });

    it('"＜7.0" (全角不等号) → refMax=7.0', () => {
      expect(parseNormalRange("＜7.0")).toEqual({ refMax: 7.0 });
    });

    it('"<110" (整数) → refMax=110', () => {
      expect(parseNormalRange("<110")).toEqual({ refMax: 110 });
    });
  });

  describe("不等号表記 — 下限のみ", () => {
    it('">3.5" → refMin=3.5 のみ (refMax は undefined)', () => {
      expect(parseNormalRange(">3.5")).toEqual({ refMin: 3.5 });
    });

    it('"> 3.5" (スペースあり) → refMin=3.5', () => {
      expect(parseNormalRange("> 3.5")).toEqual({ refMin: 3.5 });
    });

    it('"＞3.5" (全角不等号) → refMin=3.5', () => {
      expect(parseNormalRange("＞3.5")).toEqual({ refMin: 3.5 });
    });

    it('">70" (整数) → refMin=70', () => {
      expect(parseNormalRange(">70")).toEqual({ refMin: 70 });
    });
  });

  describe("日本語表記 — 両端", () => {
    it('"3.5以上7.0以下" → refMin=3.5, refMax=7.0', () => {
      expect(parseNormalRange("3.5以上7.0以下")).toEqual({ refMin: 3.5, refMax: 7.0 });
    });

    it('"3.5 以上 7.0 以下" (スペースあり) → refMin=3.5, refMax=7.0', () => {
      expect(parseNormalRange("3.5 以上 7.0 以下")).toEqual({ refMin: 3.5, refMax: 7.0 });
    });

    it('"70以上110以下" (整数) → refMin=70, refMax=110', () => {
      expect(parseNormalRange("70以上110以下")).toEqual({ refMin: 70, refMax: 110 });
    });
  });

  describe("日本語表記 — 片側", () => {
    it('"3.5以上" → refMin=3.5 のみ (refMax は undefined)', () => {
      expect(parseNormalRange("3.5以上")).toEqual({ refMin: 3.5 });
    });

    it('"7.0以下" → refMax=7.0 のみ (refMin は undefined)', () => {
      expect(parseNormalRange("7.0以下")).toEqual({ refMax: 7.0 });
    });

    it('"70以上" (整数) → refMin=70', () => {
      expect(parseNormalRange("70以上")).toEqual({ refMin: 70 });
    });

    it('"110以下" (整数) → refMax=110', () => {
      expect(parseNormalRange("110以下")).toEqual({ refMax: 110 });
    });
  });

  describe("片側のみ抽出時の保存挙動 — refMin / refMax のいずれかが undefined", () => {
    it('"<7.0" は refMin プロパティ自体を持たない (Object.hasOwn=false)', () => {
      const result = parseNormalRange("<7.0");
      expect(Object.hasOwn(result, "refMin")).toBe(false);
      expect(result.refMax).toBe(7.0);
    });

    it('">3.5" は refMax プロパティ自体を持たない (Object.hasOwn=false)', () => {
      const result = parseNormalRange(">3.5");
      expect(result.refMin).toBe(3.5);
      expect(Object.hasOwn(result, "refMax")).toBe(false);
    });

    it('"3.5以上" は refMax プロパティ自体を持たない', () => {
      const result = parseNormalRange("3.5以上");
      expect(result.refMin).toBe(3.5);
      expect(Object.hasOwn(result, "refMax")).toBe(false);
    });

    it('"7.0以下" は refMin プロパティ自体を持たない', () => {
      const result = parseNormalRange("7.0以下");
      expect(Object.hasOwn(result, "refMin")).toBe(false);
      expect(result.refMax).toBe(7.0);
    });
  });
});
