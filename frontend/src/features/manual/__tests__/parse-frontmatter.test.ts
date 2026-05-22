import { describe, it, expect } from "vitest";

import { parseFrontmatter } from "../lib/manual-index";

describe("parseFrontmatter", () => {
  it("A: 標準的な frontmatter を解析する", () => {
    const raw = `---
title: ログイン
order: 1
section: 基本
---

# ログイン

本文です。
`;
    const { meta, body } = parseFrontmatter(raw);
    expect(meta.title).toBe("ログイン");
    expect(meta.order).toBe(1);
    expect(meta.section).toBe("基本");
    expect(body).toContain("# ログイン");
    expect(body).toContain("本文です。");
  });

  it("B: frontmatter が無い場合は meta 空・body 全文", () => {
    const raw = "# 見出しのみ\n\n本文";
    const { meta, body } = parseFrontmatter(raw);
    expect(meta).toEqual({});
    expect(body).toBe(raw);
  });

  it("C: ダブルクォート/シングルクォートを除去する", () => {
    const raw = `---
title: "クォート付き"
section: 'シングル'
---

本文`;
    const { meta } = parseFrontmatter(raw);
    expect(meta.title).toBe("クォート付き");
    expect(meta.section).toBe("シングル");
  });

  it("D: order は数値に変換される", () => {
    const raw = `---
title: テスト
order: 42
---

本文`;
    const { meta } = parseFrontmatter(raw);
    expect(meta.order).toBe(42);
    expect(typeof meta.order).toBe("number");
  });

  it("E: 未知のキーは無視される（型安全性）", () => {
    const raw = `---
title: テスト
unknown_key: 何か
---

本文`;
    const { meta } = parseFrontmatter(raw);
    expect(meta.title).toBe("テスト");
    expect("unknown_key" in meta).toBe(false);
  });

  it("F: コロンが値に含まれる場合も最初のコロンで分割", () => {
    const raw = `---
title: 11:00 開始
---

本文`;
    const { meta } = parseFrontmatter(raw);
    expect(meta.title).toBe("11:00 開始");
  });

  it("G: 終端 --- がない場合は frontmatter として扱わない", () => {
    const raw = `---
title: 未完
本文だけ`;
    const { meta, body } = parseFrontmatter(raw);
    expect(meta).toEqual({});
    expect(body).toBe(raw);
  });
});
