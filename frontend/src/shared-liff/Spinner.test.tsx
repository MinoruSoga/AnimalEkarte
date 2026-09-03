import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Spinner } from "./Spinner";

describe("Spinner（FE5-22: liff/line-reserve 3箇所のコピペ統合）", () => {
  it("デフォルト（liff 系: LoadingPage/LiffLinkPage 相当）は旧マークアップと等価な DOM を出力する", () => {
    const { container } = render(<Spinner />);
    const el = container.firstElementChild as HTMLElement;

    expect(el.tagName).toBe("DIV");
    expect(el.getAttribute("aria-hidden")).toBe("true");
    // 旧マークアップ: "w-10 h-10 border-4 border-liff-brand border-t-transparent rounded-full animate-spin mx-auto mb-4"
    for (const cls of [
      "w-10",
      "h-10",
      "border-4",
      "border-liff-brand",
      "border-t-transparent",
      "rounded-full",
      "animate-spin",
      "mx-auto",
      "mb-4",
    ]) {
      expect(el.classList.contains(cls)).toBe(true);
    }
    expect(el.classList.length).toBe(9);
  });

  it('size="sm" + noah-teal（line-reserve/App.tsx 相当）は旧マークアップと等価な DOM を出力する', () => {
    const { container } = render(<Spinner size="sm" borderColorClassName="border-noah-teal" />);
    const el = container.firstElementChild as HTMLElement;

    expect(el.tagName).toBe("DIV");
    expect(el.getAttribute("aria-hidden")).toBe("true");
    // 旧マークアップ: "w-8 h-8 border-4 border-noah-teal border-t-transparent rounded-full animate-spin mx-auto mb-3"
    for (const cls of [
      "w-8",
      "h-8",
      "border-4",
      "border-noah-teal",
      "border-t-transparent",
      "rounded-full",
      "animate-spin",
      "mx-auto",
      "mb-3",
    ]) {
      expect(el.classList.contains(cls)).toBe(true);
    }
    expect(el.classList.length).toBe(9);
  });
});
