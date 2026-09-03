import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { OwnerAccountingHistory } from "./OwnerAccountingHistory";
import {
  mockOwnerId,
  completedFixture,
  completedFixture2,
  waitingFixture,
  pendingFixture,
  cancelledFixture,
  createWrapper,
} from "../lib/owner-accounting-history.test-fixtures";

/**
 * 飼主詳細から会計履歴を表示し、完了済の会計には「明細兼領収書」リンクが
 * 出ることを保証するテスト。
 *
 * 「再発行」は新 API ではなく、既存の /accounting/:id の AccountingDocument
 * プレビュー＆印刷を再利用する設計のため、本テストでは「リンクが正しい
 * 詳細パスを指していること」までを検証する。
 *
 * FE4-18: 867 行(分割前)だったため describe 境界で分割。ページネーション/URL
 * クエリ同期/未払い警告バナーの回帰は OwnerAccountingHistory.pagination.test.tsx
 * を参照。fixture/render ヘルパーは owner-accounting-history.test-fixtures.ts に
 * 共有化した（値は逐語移動・1 文字も変えていない）。
 */

describe("OwnerAccountingHistory", () => {
  beforeEach(() => {
    server.use(
      http.get("/api/v1/accountings", ({ request }) => {
        const url = new URL(request.url);
        // owner_id クエリが正しく送られているか確認した上で結果を返す
        if (url.searchParams.get("owner_id") !== mockOwnerId) {
          return HttpResponse.json({ data: [], total: 0, page: 1, limit: 50 });
        }
        return HttpResponse.json({
          data: [completedFixture, waitingFixture],
          total: 2,
          page: 1,
          limit: 50,
        });
      }),
    );
  });

  it("会計履歴を受付日降順で表示する", async () => {
    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText("ぽち")).toBeInTheDocument();
    });
    expect(screen.getByText("たま")).toBeInTheDocument();

    // 降順: 102 (2026-04-29) → 101 (2026-04-20)
    const idCells = screen
      .getAllByRole("cell")
      .filter((c) => c.textContent === "101" || c.textContent === "102");
    expect(idCells[0]).toHaveTextContent("102");
    expect(idCells[1]).toHaveTextContent("101");
  });

  it("完了済会計には明細兼領収書リンクが表示され、詳細ページへ遷移する", async () => {
    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    const link = await screen.findByRole("link", {
      name: /受付No 101 の明細兼領収書を表示/,
    });
    expect(link).toHaveAttribute("href", "/accounting/101");
  });

  it("未完了会計には明細兼領収書リンクが表示されない", async () => {
    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    await screen.findByText("たま");

    expect(
      screen.queryByRole("link", {
        name: /受付No 102 の明細兼領収書を表示/,
      }),
    ).not.toBeInTheDocument();
    // 詳細リンクは出る
    expect(screen.getByRole("link", { name: /受付No 102 の会計詳細を開く/ })).toHaveAttribute(
      "href",
      "/accounting/102",
    );
  });

  it("会計が0件のとき空状態メッセージを表示する", async () => {
    server.use(
      http.get("/api/v1/accountings", () =>
        HttpResponse.json({ data: [], total: 0, page: 1, limit: 50 }),
      ),
    );

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    expect(await screen.findByText("会計履歴はありません。")).toBeInTheDocument();
  });

  it("累計支払い金額は completed の billingAmount を合計して表示する", async () => {
    server.use(
      http.get("/api/v1/accountings", () =>
        HttpResponse.json({
          data: [completedFixture, completedFixture2, waitingFixture, cancelledFixture],
          total: 4,
          page: 1,
          limit: 50,
        }),
      ),
    );

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    const summary = await screen.findByTestId("accounting-history-summary");
    // completed 2 件 (5500 + 3300) を合計、waiting/cancelled は集計対象外
    expect(summary).toHaveTextContent("精算済 2 件");
    expect(summary).toHaveTextContent("¥8,800");
    // cancelled は accounted から除外されているため文言に「3」は出ない
    expect(summary).not.toHaveTextContent("精算済 3 件");
  });

  it("completed が 0 件のときは累計支払い金額の表示自体を出さない", async () => {
    server.use(
      http.get("/api/v1/accountings", () =>
        HttpResponse.json({
          data: [waitingFixture, cancelledFixture],
          total: 2,
          page: 1,
          limit: 50,
        }),
      ),
    );

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    // 履歴自体は描画される（waiting 行が描画されるのを待つ）
    await screen.findByText("たま");
    expect(screen.queryByTestId("accounting-history-summary")).not.toBeInTheDocument();
  });

  it("waiting / pending が 1 件以上あれば未払い警告を件数付きで表示する", async () => {
    server.use(
      http.get("/api/v1/accountings", () =>
        HttpResponse.json({
          data: [completedFixture, waitingFixture, pendingFixture, cancelledFixture],
          total: 4,
          page: 1,
          limit: 50,
        }),
      ),
    );

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    // role=alert は既存のエラー文言と被るが、ここは isError=false なので
    // 未払い警告だけが alert として描画される
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("未払いの会計が");
    expect(alert).toHaveTextContent("2");
    expect(alert).toHaveTextContent("件");
    expect(within(alert).getByRole("button")).toHaveClass("min-h-11");
  });

  it("会計詳細リンクの操作領域を44px以上に保つ", async () => {
    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    const detailLink = await screen.findByRole("link", {
      name: /受付No 102 の会計詳細を開く/,
    });
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it("cancelled のみのときは未払い警告を出さない", async () => {
    server.use(
      http.get("/api/v1/accountings", () =>
        HttpResponse.json({
          data: [completedFixture, cancelledFixture],
          total: 2,
          page: 1,
          limit: 50,
        }),
      ),
    );

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    // 完了済の summary が出るまで待つ
    await screen.findByTestId("accounting-history-summary");

    // 警告は描画されない（role=alert を持つ要素が無い）
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("初期状態は受付日 × 降順で並ぶ（既存挙動を維持）", async () => {
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
      wrapper: createWrapper(),
    });

    await screen.findByText("ぽち");

    const idCells = screen
      .getAllByRole("cell")
      .filter((c) => ["101", "102", "103"].includes(c.textContent ?? ""));
    // 102 (04-29) → 103 (04-22) → 101 (04-20)
    expect(idCells.map((c) => c.textContent)).toEqual(["102", "103", "101"]);
  });

  it("ソート方向の Toggle ボタンが昇順 / 降順を切り替える", async () => {
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

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    await screen.findByText("ぽち");

    // 初期状態: 降順 → 「降順 — クリックで昇順に切替」ボタン
    const toggle = screen.getByRole("button", {
      name: /降順 — クリックで昇順に切替/,
    });
    expect(toggle).toHaveAttribute("aria-pressed", "false");

    await user.click(toggle);
    // 昇順に切り替わると aria-label / aria-pressed が更新される
    const ascToggle = screen.getByRole("button", {
      name: /昇順 — クリックで降順に切替/,
    });
    expect(ascToggle).toHaveAttribute("aria-pressed", "true");

    // 受付日昇順 → 101 (04-20) が先頭
    const idCells = screen
      .getAllByRole("cell")
      .filter((c) => ["101", "102", "103"].includes(c.textContent ?? ""));
    expect(idCells[0]).toHaveTextContent("101");
    expect(idCells[2]).toHaveTextContent("102");
  });

  it("Select で「金額」を選ぶと completed の金額順、未完了は 0 として末尾扱い", async () => {
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

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    await screen.findByText("ぽち");

    // Select を「金額」に切替（初期は降順）
    await user.click(screen.getByRole("combobox", { name: "ソート項目" }));
    await user.click(screen.getByRole("option", { name: "金額" }));

    // 金額降順: 101 (5500) → 103 (3300) → 102 (waiting=0)
    const idCells = screen
      .getAllByRole("cell")
      .filter((c) => ["101", "102", "103"].includes(c.textContent ?? ""));
    expect(idCells[0]).toHaveTextContent("101");
    expect(idCells[1]).toHaveTextContent("103");
    expect(idCells[2]).toHaveTextContent("102");
  });

  it("Select で「ステータス」昇順を選ぶと waiting → pending → completed → cancelled の順に並ぶ", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("/api/v1/accountings", () =>
        HttpResponse.json({
          data: [completedFixture, waitingFixture, pendingFixture, cancelledFixture],
          total: 4,
          page: 1,
          limit: 50,
        }),
      ),
    );

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    await screen.findByText("ハチ");

    await user.click(screen.getByRole("combobox", { name: "ソート項目" }));
    await user.click(screen.getByRole("option", { name: "ステータス" }));
    await user.click(screen.getByRole("button", { name: /降順 — クリックで昇順に切替/ }));

    // ID 順: 101=completed, 102=waiting, 104=pending, 105=cancelled
    const idCells = screen
      .getAllByRole("cell")
      .filter((c) => ["101", "102", "104", "105"].includes(c.textContent ?? ""));
    // 昇順並び: waiting(102) → pending(104) → completed(101) → cancelled(105)
    expect(idCells.map((c) => c.textContent)).toEqual(["102", "104", "101", "105"]);
  });

  it("ソート変更後も累計支払い金額の集計値はソートと独立して保たれる", async () => {
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

    render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
      wrapper: createWrapper(),
    });

    const summaryBefore = await screen.findByTestId("accounting-history-summary");
    expect(within(summaryBefore).getByText(/¥8,800/)).toBeInTheDocument();

    // ソートを切替
    await user.click(screen.getByRole("combobox", { name: "ソート項目" }));
    await user.click(screen.getByRole("option", { name: "金額" }));

    const summaryAfter = screen.getByTestId("accounting-history-summary");
    // 累計値は変わらない
    expect(within(summaryAfter).getByText(/¥8,800/)).toBeInTheDocument();
    expect(within(summaryAfter).getByText(/精算済 2 件/)).toBeInTheDocument();
  });

  it("ownerId が空文字なら API を叩かず空状態にもならない（不正状態を抑止）", async () => {
    let called = false;
    server.use(
      http.get("/api/v1/accountings", () => {
        called = true;
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 50 });
      }),
    );

    render(<OwnerAccountingHistory ownerId="" />, {
      wrapper: createWrapper(),
    });

    // queryFn が enabled=false で実行されないことを確認するため少し待つ
    await new Promise((r) => setTimeout(r, 50));
    expect(called).toBe(false);
  });
});
