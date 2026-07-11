import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useSearchParams } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";

import { OwnerAccountingHistory } from "./OwnerAccountingHistory";
import type { BackendAccounting } from "../api/types";

/**
 * 飼主詳細から会計履歴を表示し、完了済の会計には「明細兼領収書」リンクが
 * 出ることを保証するテスト。
 *
 * 「再発行」は新 API ではなく、既存の /accounting/:id の AccountingDocument
 * プレビュー＆印刷を再利用する設計のため、本テストでは「リンクが正しい
 * 詳細パスを指していること」までを検証する。
 */

const mockOwnerId = "42";

// BackendAccounting の最小フィクスチャを作る。transformToAccounting が
// 期待するキーだけ埋め、その他の任意フィールドは null/undefined で OK。
function makeBackendAccounting(overrides: Partial<BackendAccounting>): BackendAccounting {
  return {
    id: 0,
    owner_id: Number(mockOwnerId),
    pet_id: 0,
    status: "waiting",
    scheduled_date: "2026-04-29T00:00:00Z",
    subtotal: 0,
    tax_total: 0,
    total_amount: 0,
    items: [],
    payments: [],
    ...overrides,
  } as BackendAccounting;
}

const completedFixture: BackendAccounting = makeBackendAccounting({
  id: 101,
  status: "completed",
  scheduled_date: "2026-04-20T00:00:00Z",
  subtotal: 5000,
  tax_total: 500,
  total_amount: 5500,
  pet: { id: 7, name: "ぽち", animal_species: { id: 1, name: "犬" } },
  payments: [
    {
      id: 201,
      billing_id: 101,
      method: "cash",
      total_amount: 5500,
      billing_amount: 5500,
      received_amount: 6000,
      change_amount: 500,
      insurance_amount: 0,
      discount_amount: 0,
    },
  ],
});

const completedFixture2: BackendAccounting = makeBackendAccounting({
  id: 103,
  status: "completed",
  scheduled_date: "2026-04-22T00:00:00Z",
  subtotal: 3000,
  tax_total: 300,
  total_amount: 3300,
  pet: { id: 9, name: "もも", animal_species: { id: 1, name: "犬" } },
  payments: [
    {
      id: 203,
      billing_id: 103,
      method: "credit_card",
      total_amount: 3300,
      billing_amount: 3300,
      received_amount: 3300,
      change_amount: 0,
      insurance_amount: 0,
      discount_amount: 0,
    },
  ],
});

const waitingFixture: BackendAccounting = makeBackendAccounting({
  id: 102,
  status: "waiting",
  scheduled_date: "2026-04-29T00:00:00Z",
  pet: { id: 8, name: "たま", animal_species: { id: 2, name: "猫" } },
});

const pendingFixture: BackendAccounting = makeBackendAccounting({
  id: 104,
  status: "pending",
  scheduled_date: "2026-04-28T00:00:00Z",
  pet: { id: 10, name: "ちゃこ", animal_species: { id: 2, name: "猫" } },
});

const cancelledFixture: BackendAccounting = makeBackendAccounting({
  id: 105,
  status: "cancelled",
  scheduled_date: "2026-04-15T00:00:00Z",
  pet: { id: 11, name: "ハチ", animal_species: { id: 1, name: "犬" } },
});

const createWrapper = (initialEntries?: string[]) =>
  createTestWrapper({ initialEntries: initialEntries ?? ["/"] });

/** PAGE_SIZE=10 を超えるフィクスチャを生成する。日付は 2026-04-01 から連番。 */
const makePaginationFixtures = (n: number): BackendAccounting[] =>
  Array.from({ length: n }, (_, i) =>
    makeBackendAccounting({
      id: 200 + i,
      status: "completed",
      scheduled_date: `2026-04-${String(i + 1).padStart(2, "0")}T00:00:00Z`,
      total_amount: (i + 1) * 1000,
      payments: [
        {
          id: 300 + i,
          billing_id: 200 + i,
          method: "cash",
          total_amount: (i + 1) * 1000,
          billing_amount: (i + 1) * 1000,
          received_amount: (i + 1) * 1000,
          change_amount: 0,
          insurance_amount: 0,
          discount_amount: 0,
        },
      ],
    }),
  );

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
    const idCells = screen.getAllByRole("cell").filter((c) =>
      c.textContent === "101" || c.textContent === "102",
    );
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
    expect(
      screen.getByRole("link", { name: /受付No 102 の会計詳細を開く/ }),
    ).toHaveAttribute("href", "/accounting/102");
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

    expect(
      await screen.findByText("会計履歴はありません。"),
    ).toBeInTheDocument();
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

    const idCells = screen.getAllByRole("cell").filter((c) =>
      ["101", "102", "103"].includes(c.textContent ?? ""),
    );
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
    const idCells = screen.getAllByRole("cell").filter((c) =>
      ["101", "102", "103"].includes(c.textContent ?? ""),
    );
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
    const idCells = screen.getAllByRole("cell").filter((c) =>
      ["101", "102", "103"].includes(c.textContent ?? ""),
    );
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
    const idCells = screen.getAllByRole("cell").filter((c) =>
      ["101", "102", "104", "105"].includes(c.textContent ?? ""),
    );
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

  describe("ページネーション", () => {
    it("10 件以下ではページネーション UI が非表示", async () => {
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({ data: [completedFixture, waitingFixture], total: 2, page: 1, limit: 50 }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, { wrapper: createWrapper() });
      await screen.findByText("ぽち");
      expect(screen.queryByRole("navigation", { name: "ページネーション" })).not.toBeInTheDocument();
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
      expect(screen.getByText(/1 \/ 2 ページ/)).toBeInTheDocument();
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
      const idCells = screen.getAllByRole("cell").filter((c) =>
        /^2[01]\d$/.test(c.textContent ?? ""),
      );
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
      const idCells = screen.getAllByRole("cell").filter((c) =>
        /^2[01]\d$/.test(c.textContent ?? ""),
      );
      expect(idCells.map((c) => c.textContent)).toContain("200");
      expect(idCells.map((c) => c.textContent)).not.toContain("211");
      expect(idCells).toHaveLength(2);

      expect(screen.getByRole("button", { name: "次のページ" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "前のページ" })).toBeEnabled();
      expect(screen.getByText(/2 \/ 2 ページ/)).toBeInTheDocument();
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

      const idCells = screen.getAllByRole("cell").filter((c) =>
        /^2[01]\d$/.test(c.textContent ?? ""),
      );
      expect(idCells.map((c) => c.textContent)).toContain("211");
      expect(idCells.map((c) => c.textContent)).not.toContain("200");
      expect(idCells).toHaveLength(10);
      expect(screen.getByText(/1 \/ 2 ページ/)).toBeInTheDocument();
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
      expect(screen.getByText(/2 \/ 2 ページ/)).toBeInTheDocument();

      await user.click(screen.getByRole("combobox", { name: "ソート項目" }));
      await user.click(screen.getByRole("option", { name: "金額" }));

      expect(screen.getByText(/1 \/ 2 ページ/)).toBeInTheDocument();
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
      expect(screen.getByText(/2 \/ 2 ページ/)).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /降順 — クリックで昇順に切替/ }));
      expect(screen.getByText(/1 \/ 2 ページ/)).toBeInTheDocument();
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
            total: 3, page: 1, limit: 50,
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
      const idCells = screen.getAllByRole("cell").filter((c) =>
        ["101", "102", "103"].includes(c.textContent ?? ""),
      );
      expect(idCells[0]).toHaveTextContent("101");
    });

    it("ah_order=asc で初期描画すると昇順ボタンが押された状態になる", async () => {
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, completedFixture2, waitingFixture],
            total: 3, page: 1, limit: 50,
          }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
        wrapper: createWrapper(["/?ah_order=asc"]),
      });
      await screen.findByText("ぽち");
      expect(
        screen.getByRole("button", { name: /昇順 — クリックで降順に切替/ }),
      ).toHaveAttribute("aria-pressed", "true");
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
      await screen.findByText(/2 \/ 2 ページ/);
      const idCells = screen.getAllByRole("cell").filter((c) =>
        /^2[01]\d$/.test(c.textContent ?? ""),
      );
      expect(idCells.map((c) => c.textContent)).toContain("200");
      expect(idCells.map((c) => c.textContent)).not.toContain("211");
    });

    it("ah_sort に不正値を渡すとデフォルト「受付日」にフォールバックする", async () => {
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, completedFixture2, waitingFixture],
            total: 3, page: 1, limit: 50,
          }),
        ),
      );
      render(<OwnerAccountingHistory ownerId={mockOwnerId} />, {
        wrapper: createWrapper(["/?ah_sort=invalid_value"]),
      });
      await screen.findByText("ぽち");
      expect(screen.getByRole("combobox", { name: "ソート項目" })).toHaveTextContent("受付日");
      // date desc: 102(04-29) が先頭
      const idCells = screen.getAllByRole("cell").filter((c) =>
        ["101", "102", "103"].includes(c.textContent ?? ""),
      );
      expect(idCells[0]).toHaveTextContent("102");
    });

    it("ソートフィールド変更で URL に ah_sort が反映される", async () => {
      const user = userEvent.setup();
      server.use(
        http.get("/api/v1/accountings", () =>
          HttpResponse.json({
            data: [completedFixture, completedFixture2, waitingFixture],
            total: 3, page: 1, limit: 50,
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
      expect(screen.getByText(/1 \/ 2 ページ/)).toBeInTheDocument();
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
      expect(
        screen.getByRole("button", { name: /先頭の未払い行を確認する/ }),
      ).toBeInTheDocument();
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
      await screen.findByText(/2 \/ 2 ページ/);
      const targetRow = screen.getByTestId("first-unpaid-row");
      expect(window.HTMLElement.prototype.scrollIntoView).toHaveBeenCalledTimes(1);
      expect(targetRow).toHaveFocus();
    });
  });
});
