import { describe, expect, it } from "vitest";

import { getSafeMarkdownHref } from "./ManualContent";

describe("getSafeMarkdownHref", () => {
  it("許可された外部リンクと相対リンクをそのまま返す", () => {
    expect(getSafeMarkdownHref("https://example.com/docs")).toBe("https://example.com/docs");
    expect(getSafeMarkdownHref("http://example.com/docs")).toBe("http://example.com/docs");
    expect(getSafeMarkdownHref("mailto:support@example.com")).toBe("mailto:support@example.com");
    expect(getSafeMarkdownHref("/manual/screens/01-login")).toBe("/manual/screens/01-login");
    expect(getSafeMarkdownHref("./images/help.png")).toBe("./images/help.png");
    expect(getSafeMarkdownHref("#section")).toBe("#section");
  });

  it("実行系・ローカルファイル系・protocol-relative URL を拒否する", () => {
    expect(getSafeMarkdownHref("javascript:alert(1)")).toBeUndefined();
    expect(getSafeMarkdownHref("JaVaScRiPt:alert(1)")).toBeUndefined();
    expect(getSafeMarkdownHref("javascript%3Aalert(1)")).toBeUndefined();
    expect(getSafeMarkdownHref("java\nscript:alert(1)")).toBeUndefined();
    expect(getSafeMarkdownHref("data:text/html,<script>alert(1)</script>")).toBeUndefined();
    expect(getSafeMarkdownHref("vbscript:msgbox(1)")).toBeUndefined();
    expect(getSafeMarkdownHref("file:///etc/passwd")).toBeUndefined();
    expect(getSafeMarkdownHref("//evil.example/path")).toBeUndefined();
  });

  it("空値や文字列以外は拒否する", () => {
    expect(getSafeMarkdownHref("")).toBeUndefined();
    expect(getSafeMarkdownHref("   ")).toBeUndefined();
    expect(getSafeMarkdownHref(undefined)).toBeUndefined();
    expect(getSafeMarkdownHref(null)).toBeUndefined();
  });
});
