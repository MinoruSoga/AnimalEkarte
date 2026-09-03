import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { Pagination } from "./Pagination";

const noop = () => {};

describe("Pagination", () => {
  it("totalCount=0 のとき何もレンダリングしない", () => {
    render(
      <Pagination
        currentPage={1}
        totalPages={0}
        totalCount={0}
        startIndex={0}
        endIndex={0}
        onPageChange={noop}
        onPrev={noop}
        onNext={noop}
      />
    );
    expect(screen.queryByRole("navigation")).not.toBeInTheDocument();
  });

  it("件数情報を表示する", () => {
    render(
      <Pagination
        currentPage={2}
        totalPages={3}
        totalCount={25}
        startIndex={11}
        endIndex={20}
        onPageChange={noop}
        onPrev={noop}
        onNext={noop}
      />
    );
    expect(screen.getByText("25件中 11-20件")).toBeInTheDocument();
  });

  describe("a11y (FE-RC-044)", () => {
    // nav ランドマーク自体はこのコンポーネントで付けない（OwnerAccountingHistory.tsx など、
    // 既に呼び出し側で `<nav aria-label="ページネーション">` を外付けする規約があるため。
    // ここで同名の nav を追加すると入れ子で aria-label が重複し、呼び出し側の既存テストが
    // 「複数要素が見つかる」で壊れる。詳細は Pagination.tsx のコメント参照）。

    it("現在のページボタンに aria-current='page' が付く（accessible name は数字のまま）", () => {
      render(
        <Pagination
          currentPage={2}
          totalPages={3}
          totalCount={25}
          startIndex={11}
          endIndex={20}
          onPageChange={noop}
          onPrev={noop}
          onNext={noop}
        />
      );
      // aria-label は付けない: 他 feature の既存テストが getByRole("button", { name: "2" })
      // で数字そのものを accessible name として参照しているため、上書きしてはならない。
      const current = screen.getByRole("button", { name: "2" });
      expect(current).toHaveAttribute("aria-current", "page");

      const other = screen.getByRole("button", { name: "1" });
      expect(other).not.toHaveAttribute("aria-current");
    });

    it("省略記号(...)は装飾として aria-hidden を持ち、SRのページ番号読み上げを妨げない", () => {
      render(
        <Pagination
          currentPage={1}
          totalPages={10}
          totalCount={100}
          startIndex={1}
          endIndex={10}
          onPageChange={noop}
          onPrev={noop}
          onNext={noop}
        />
      );
      const ellipses = screen.getAllByText("...");
      for (const el of ellipses) {
        expect(el).toHaveAttribute("aria-hidden", "true");
      }
    });

    it("前後/最初/最後ボタンのアイコンは装飾（aria-hidden）で、意味はボタンの aria-label が伝える", () => {
      render(
        <Pagination
          currentPage={2}
          totalPages={3}
          totalCount={25}
          startIndex={11}
          endIndex={20}
          onPageChange={noop}
          onPrev={noop}
          onNext={noop}
        />
      );
      for (const name of ["最初のページ", "前のページ", "次のページ", "最後のページ"]) {
        const button = screen.getByRole("button", { name });
        const icon = button.querySelector("svg");
        expect(icon).toHaveAttribute("aria-hidden", "true");
      }
    });
  });

  describe("操作", () => {
    it("ページ番号クリックで onPageChange が呼ばれる", () => {
      const onPageChange = vi.fn();
      render(
        <Pagination
          currentPage={1}
          totalPages={3}
          totalCount={25}
          startIndex={1}
          endIndex={10}
          onPageChange={onPageChange}
          onPrev={noop}
          onNext={noop}
        />
      );
      screen.getByRole("button", { name: "2" }).click();
      expect(onPageChange).toHaveBeenCalledWith(2);
    });

    it("先頭ページでは「最初」「前」が disabled", () => {
      render(
        <Pagination
          currentPage={1}
          totalPages={3}
          totalCount={25}
          startIndex={1}
          endIndex={10}
          onPageChange={noop}
          onPrev={noop}
          onNext={noop}
        />
      );
      expect(screen.getByRole("button", { name: "最初のページ" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "前のページ" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "次のページ" })).toBeEnabled();
      expect(screen.getByRole("button", { name: "最後のページ" })).toBeEnabled();
    });

    it("最終ページでは「次」「最後」が disabled", () => {
      render(
        <Pagination
          currentPage={3}
          totalPages={3}
          totalCount={25}
          startIndex={21}
          endIndex={25}
          onPageChange={noop}
          onPrev={noop}
          onNext={noop}
        />
      );
      expect(screen.getByRole("button", { name: "次のページ" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "最後のページ" })).toBeDisabled();
    });
  });
});
