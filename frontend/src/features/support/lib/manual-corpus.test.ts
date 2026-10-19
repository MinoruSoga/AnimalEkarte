import { describe, expect, it } from "vitest";
import { screenArticles, workflowArticles } from "@/lib/manual-index";
import { canonicalCorpusJson, baseCorpusRevision, verifyBaseCorpusRevision } from "./manual-corpus";

describe("manual corpus constants", () => {
  it("exposes generated base revision as baseCorpusRevision", () => {
    expect(baseCorpusRevision).toMatch(/^[0-9a-f]{64}$/);
    expect(baseCorpusRevision).not.toBe("");
  });

  it("bundled base corpus hashes to baseCorpusRevision", async () => {
    const base = [...screenArticles, ...workflowArticles];
    expect(base.length).toBeGreaterThan(0);
    await expect(verifyBaseCorpusRevision(base)).resolves.toBe(true);
  });

  it("returns false for a corpus that differs from the generated base", async () => {
    const articles = [
      {
        category: "screens" as const,
        slug: "test",
        title: "T",
        section: "S",
        content: "body",
        keywords: [],
        keywords_vi: [],
      },
    ];
    await expect(verifyBaseCorpusRevision(articles)).resolves.toBe(false);
  });
});

describe("canonicalCorpusJson", () => {
  // EMR-441: slug 整列は backend sortCorpusArticles (UTF-8 バイト順 = code point 順)
  // と同一。JS 既定の UTF-16 単位順だとサロゲートペア slug の位置が入れ替わり
  // revision hash が不一致 (全 chat で 409) になる。
  it("supplementary-plane slug と BMP slug の整列は UTF-8 順（code point 順）で一致する", () => {
    const mk = (slug: string) => ({
      category: "screens",
      slug,
      title: slug,
      order: 1,
      section: "",
      content: "",
      searchText: "",
      keywords: [],
      keywords_vi: [],
    });
    const json = canonicalCorpusJson([mk("𠮷"), mk("Ｚ")]);
    // UTF-8 バイト順: Ｚ (EF BC AA) < 𠮷 (F0 A0 AE B7)。JS UTF-16 順だと逆になる。
    expect(json.indexOf('"slug":"Ｚ"')).toBeGreaterThanOrEqual(0);
    expect(json.indexOf('"slug":"𠮷"')).toBeGreaterThanOrEqual(0);
    expect(json.indexOf('"slug":"Ｚ"')).toBeLessThan(json.indexOf('"slug":"𠮷"'));
  });

  it("escapes actual U+2028 but preserves literal backslash-u2028", () => {
    const json = canonicalCorpusJson([
      {
        category: "screens",
        slug: "x",
        title: "t",
        order: 1,
        section: "s",
        content: "a\u2028b" + "\\u2028" + "c\u2029d<>&😀é",
        searchText: "",
        keywords: [],
        keywords_vi: [],
      },
    ]);
    expect(json).toContain("\\u2028");
    expect(json).toContain("\\\\u2028");
    expect(json).toContain("\\u2029");
    expect(JSON.parse(json)[0].body).toBe("a\u2028b\\u2028c\u2029d<>&😀é");
  });
  it("sorted byte-lexically, fixed field order, no whitespace", () => {
    const json = canonicalCorpusJson([
      {
        category: "workflows",
        slug: "b",
        title: "",
        order: 1,
        section: "",
        content: "",
        searchText: "",
        keywords: [],
        keywords_vi: [],
      },
      {
        category: "screens",
        slug: "a",
        title: "",
        order: 1,
        section: "",
        content: "",
        searchText: "",
        keywords: [],
        keywords_vi: [],
      },
    ]);
    expect(json.startsWith('[{"category":"screens"')).toBe(true);
    expect(json).not.toContain(" ");
  });
});
