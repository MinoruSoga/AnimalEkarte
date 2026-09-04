import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useSearchParams } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { OwnerAccountingHistory } from "./OwnerAccountingHistory";
import {
  mockOwnerId,
  completedFixture,
  completedFixture2,
  waitingFixture,
  makeBackendAccounting,
  makePaginationFixtures,
  createWrapper,
} from "../lib/owner-accounting-history.test-fixtures";

/**
 * FE4-18: OwnerAccountingHistory.test.tsx（867 行）から describe 境界で分割。
 * ページネーション / URL クエリ同期 / 未払い警告バナー（スクロール・フォーカス）の
 * 回帰をカバーする。fixture/render ヘルパーは owner-accounting-history.test-fixtures.ts
 * を共有（値は逐語移動・1 文字も変えていない）。基本表示・ソート等の回帰は
 * OwnerAccountingHistory.test.tsx を参照。
 */

describe("OwnerAccountingHistory", () => {
  describe("ページネーション", () => {
    it("10 件以下ではページネーション UI が非表示", async () => {
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, waitingFixture],
            total: 2,
            page: 1,
            limit: 50,
          }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByText("ぽち");
      expect(
        screen.queryByRole("navigation", { name: "ページネーション" }),
      ).not.toBeInTheDocument();
    });

    it("11 件以上ではページネーション UI が表示され総ページ数が出る", async () => {
      const fixtures = makePaginationFixtures(12);
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByRole("navigation", { name: "ページネーション" });
      // FE5-15: 正本 Pagination の文言は「X / Y ページ」ではなく「件数中 開始-終了件」
      expect(screen.getByText("12件中 1-10件")).toBeInTheDocument();
    });

    it("ページ 1 では前のページボタンが無効で最新 10 件が表示される", async () => {
      const fixtures = makePaginationFixtures(12); // ids 200-211
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByRole("navigation", { name: "ページネーション" });

      expect(screen.getByRole("button", { name: "前のページ" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "次のページ" })).toBeEnabled();

      // date desc: 211 (04-12) が先頭、200 (04-01) はページ 2
      const idCells = screen
        .getAllByRole("cell")
        .filter((c) => /^2[01]\d$/.test(c.textContent ?? ""));
      expect(idCells.map((c) => c.textContent)).toContain("211");
      expect(idCells.map((c) => c.textContent)).not.toContain("200");
      expect(idCells).toHaveLength(10);
    });

    it("次のページへ移動するとページ 2 の内容が表示される", async () => {
      const user = userEvent.setup();
      const fixtures = makePaginationFixtures(12);
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByRole("navigation", { name: "ページネーション" });

      await user.click(screen.getByRole("button", { name: "次のページ" }));

      // page 2: ids 201, 200 (oldest two)
      const idCells = screen
        .getAllByRole("cell")
        .filter((c) => /^2[01]\d$/.test(c.textContent ?? ""));
      expect(idCells.map((c) => c.textContent)).toContain("200");
      expect(idCells.map((c) => c.textContent)).not.toContain("211");
      expect(idCells).toHaveLength(2);

      expect(screen.getByRole("button", { name: "次のページ" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "前のページ" })).toBeEnabled();
      expect(screen.getByText("12件中 11-12件")).toBeInTheDocument();
    });

    it("前のページへ戻るとページ 1 の内容が再表示される", async () => {
      const user = userEvent.setup();
      const fixtures = makePaginationFixtures(12);
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByRole("navigation", { name: "ページネーション" });

      await user.click(screen.getByRole("button", { name: "次のページ" }));
      await user.click(screen.getByRole("button", { name: "前のページ" }));

      const idCells = screen
        .getAllByRole("cell")
        .filter((c) => /^2[01]\d$/.test(c.textContent ?? ""));
      expect(idCells.map((c) => c.textContent)).toContain("211");
      expect(idCells.map((c) => c.textContent)).not.toContain("200");
      expect(idCells).toHaveLength(10);
      expect(screen.getByText("12件中 1-10件")).toBeInTheDocument();
    });

    it("ソートフィールド変更でページ 1 にリセットされる", async () => {
      const user = userEvent.setup();
      const fixtures = makePaginationFixtures(12);
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByRole("navigation", { name: "ページネーション" });

      await user.click(screen.getByRole("button", { name: "次のページ" }));
      expect(screen.getByText("12件中 11-12件")).toBeInTheDocument();

      await user.click(screen.getByRole("combobox", { name: "ソート項目" }));
      await user.click(screen.getByRole("option", { name: "金額" }));

      expect(screen.getByText("12件中 1-10件")).toBeInTheDocument();
    });

    it("ソート方向切替でページ 1 にリセットされる", async () => {
      const user = userEvent.setup();
      const fixtures = makePaginationFixtures(12);
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByRole("navigation", { name: "ページネーション" });

      await user.click(screen.getByRole("button", { name: "次のページ" }));
      expect(screen.getByText("12件中 11-12件")).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /降順 — クリックで昇順に切替/ }));
      expect(screen.getByText("12件中 1-10件")).toBeInTheDocument();
    });
  });

  describe("URL クエリ同期", () => {
    function SearchParamsSpy() {
      const [searchParams] = useSearchParams();
      return <output data-testid="search-params">{searchParams.toString()}</output>;
    }

    it("ah_sort=amount で初期描画すると「金額」ソートが選択された状態になる", async () => {
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, completedFixture2, waitingFixture],
            total: 3,
            page: 1,
            limit: 50,
          }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
        wrapper: createWrapper(["/?ah_sort=amount"]),
      });
      await screen.findByText("ぽち");
      // Select が「金額」を表示している
      expect(screen.getByRole("combobox", { name: "ソート項目" })).toHaveTextContent("金額");
      // 金額降順: 101(5500) → 103(3300) → 102(waiting=0)
      const idCells = screen
        .getAllByRole("cell")
        .filter((c) => ["101", "102", "103"].includes(c.textContent ?? ""));
      expect(idCells[0]).toHaveTextContent("101");
    });

    it("ah_order=asc で初期描画すると昇順ボタンが押された状態になる", async () => {
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, completedFixture2, waitingFixture],
            total: 3,
            page: 1,
            limit: 50,
          }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
        wrapper: createWrapper(["/?ah_order=asc"]),
      });
      await screen.findByText("ぽち");
      expect(screen.getByRole("button", { name: /昇順 — クリックで降順に切替/ })).toHaveAttribute(
        "aria-pressed",
        "true",
      );
    });

    it("ah_page=2 で初期描画するとページ 2 が表示される", async () => {
      const fixtures = makePaginationFixtures(12);
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
        wrapper: createWrapper(["/?ah_page=2"]),
      });
      await screen.findByText("12件中 11-12件");
      const idCells = screen
        .getAllByRole("cell")
        .filter((c) => /^2[01]\d$/.test(c.textContent ?? ""));
      expect(idCells.map((c) => c.textContent)).toContain("200");
      expect(idCells.map((c) => c.textContent)).not.toContain("211");
    });

    it("ah_sort に不正値を渡すとデフォルト「受付日」にフォールバックする", async () => {
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, completedFixture2, waitingFixture],
            total: 3,
            page: 1,
            limit: 50,
          }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
        wrapper: createWrapper(["/?ah_sort=invalid_value"]),
      });
      await screen.findByText("ぽち");
      expect(screen.getByRole("combobox", { name: "ソート項目" })).toHaveTextContent("受付日");
      // date desc: 102(04-29) が先頭
      const idCells = screen
        .getAllByRole("cell")
        .filter((c) => ["101", "102", "103"].includes(c.textContent ?? ""));
      expect(idCells[0]).toHaveTextContent("102");
    });

    it("ソートフィールド変更で URL に ah_sort が反映される", async () => {
      const user = userEvent.setup();
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, completedFixture2, waitingFixture],
            total: 3,
            page: 1,
            limit: 50,
          }),
        ),
      );
      render(
        <>
          <OwnerAccountingHistory ownerId={mockOwnerId} />
          <SearchParamsSpy />
        </>,
        { wrapper: createWrapper() },
      );
      await screen.findByText("ぽち");
      await user.click(screen.getByRole("combobox", { name: "ソート項目" }));
      await user.click(screen.getByRole("option", { name: "金額" }));
      expect(screen.getByTestId("search-params")).toHaveTextContent("ah_sort=amount");
    });

    it("ページ移動で URL に ah_page が反映される", async () => {
      const user = userEvent.setup();
      const fixtures = makePaginationFixtures(12);
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(
        <>
          <OwnerAccountingHistory ownerId={mockOwnerId} />
          <SearchParamsSpy />
        </>,
        { wrapper: createWrapper() },
      );
      await screen.findByRole("navigation", { name: "ページネーション" });
      await user.click(screen.getByRole("button", { name: "次のページ" }));
      expect(screen.getByTestId("search-params")).toHaveTextContent("ah_page=2");
    });

    it("ページ 2 でソートを変更すると ah_page が URL から削除される", async () => {
      const user = userEvent.setup();
      const fixtures = makePaginationFixtures(12);
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: fixtures, total: 12, page: 1, limit: 50 }),
        ),
      );
      render(
        <>
          <OwnerAccountingHistory ownerId={mockOwnerId} />
          <SearchParamsSpy />
        </>,
        { wrapper: createWrapper() },
      );
      await screen.findByRole("navigation", { name: "ページネーション" });
      await user.click(screen.getByRole("button", { name: "次のページ" }));
      expect(screen.getByTestId("search-params")).toHaveTextContent("ah_page=2");

      await user.click(screen.getByRole("combobox", { name: "ソート項目" }));
      await user.click(screen.getByRole("option", { name: "金額" }));

      expect(screen.getByTestId("search-params")).not.toHaveTextContent("ah_page");
      expect(screen.getByText("12件中 1-10件")).toBeInTheDocument();
    });
  });

  describe("未払い警告バナー — スクロール・フォーカス", () => {
    beforeEach(() => {
      window.HTMLElement.prototype.scrollIntoView = vi.fn();
    });

    it("バナーにボタンロールが含まれる", async () => {
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, waitingFixture],
            total: 2,
            page: 1,
            limit: 50,
          }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByText(/未払いの会計が/);
      expect(screen.getByRole("button", { name: /先頭の未払い行を確認する/ })).toBeInTheDocument();
    });

    it("バナークリックで先頭未払い行にスクロール＆フォーカスが移る（同一ページ内）", async () => {
      const user = userEvent.setup();
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, waitingFixture],
            total: 2,
            page: 1,
            limit: 50,
          }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByText(/未払いの会計が/);

      await user.click(screen.getByRole("button", { name: /先頭の未払い行を確認する/ }));

      const targetRow = screen.getByTestId("first-unpaid-row");
      expect(window.HTMLElement.prototype.scrollIntoView).toHaveBeenCalledTimes(1);
      expect(targetRow).toHaveFocus();
    });

    it("先頭未払い行がページ 2 にある場合、クリックでページ切替後にフォーカスが移る", async () => {
      const user = userEvent.setup();

      // completed 10件 (日付降順で indices 0-9 → page 1) + waiting 1件 (最古 → index 10, page 2)
      const completedEntries = Array.from({ length: 10 }, (_, i) =>
        makeBackendAccounting({
          id: 300 + i,
          status: "completed",
          scheduled_date: `2026-04-${String(29 - i).padStart(2, "0")}T00:00:00Z`,
        }),
      );
      const unpaidEntry = makeBackendAccounting({
        id: 400,
        status: "waiting",
        scheduled_date: "2026-04-10T00:00:00Z",
      });
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [...completedEntries, unpaidEntry],
            total: 11,
            page: 1,
            limit: 50,
          }),
        ),
      );

      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByText(/未払いの会計が/);

      // page 1 にいる状態でバナークリック
      await user.click(screen.getByRole("button", { name: /先頭の未払い行を確認する/ }));

      // ページ 2 に切り替わり、未払い行にフォーカスが移る
      await screen.findByText("11件中 11-11件");
      const targetRow = screen.getByTestId("first-unpaid-row");
      expect(window.HTMLElement.prototype.scrollIntoView).toHaveBeenCalledTimes(1);
      expect(targetRow).toHaveFocus();
    });
  });
});
