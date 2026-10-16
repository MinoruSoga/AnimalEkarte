import { describe, expect, it } from "vitest";
import { canonicalCorpusJson } from "./manual-corpus";

describe("canonicalCorpusJson", () => {
  it("escapes actual U+2028 but preserves literal backslash-u2028", () => {
    const json = canonicalCorpusJson([
      {
        category: "s",
        slug: "x",
        title: "t",
        section: "s",
        body: "a\u2028b" + "\\u2028" + "c\u2029d<>&😀é",
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
        section: "",
        body: "",
        keywords: [],
        keywords_vi: [],
      },
      {
        category: "screens",
        slug: "a",
        title: "",
        section: "",
        body: "",
        keywords: [],
        keywords_vi: [],
      },
    ]);
    expect(json.startsWith('[{"category":"screens"')).toBe(true);
    expect(json).not.toContain(" ");
  });
});
