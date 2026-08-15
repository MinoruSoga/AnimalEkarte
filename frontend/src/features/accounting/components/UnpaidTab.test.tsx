import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { clinics: [{ clinicId: "1", clinicName: "ノア動物病院", isMain: true }] },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";

import { UnpaidTab } from "./UnpaidTab";

function renderTab(initialSearch = "") {
  return render(<UnpaidTab />, {
    wrapper: createTestWrapper({ initialEntries: [`/?${initialSearch}`] }),
  });
}

const MONTHLY_RESPONSE = {
  data: [
    {
      owner_id: 1,
      owner_name: "山田花子",
      pet_id: 10,
      pet_name: "チョコ",
      prev_month_carryover: 3000,
      current_month_unpaid: 5000,
      next_month_carryover: 8000,
    },
    {
      owner_id: 2,
      owner_name: "鈴木一郎",
      pet_id: null,
      pet_name: "",
      prev_month_carryover: 0,
      current_month_unpaid: 2000,
      next_month_carryover: 2000,
    },
  ],
  total: 2,
  page: 1,
  limit: 20,
  summary: {
    prev_month_carryover: 3000,
    current_month_unpaid: 7000,
    next_month_carryover: 10000,
  },
};

describe("UnpaidTab — 月次繰越モード", () => {
  beforeEach(() => {
    server.use(
      http.get("/api/v1/accountings/unpaid-monthly", () =>
        HttpResponse.json(MONTHLY_RESPONSE),
      ),
    );
  });

  it("月次繰越ボタンをクリックすると月次モードに切り替わる", async () => {
    const user = userEvent.setup();
    renderTab();

    await user.click(screen.getByRole("button", { name: "月次繰越" }));

    expect(screen.getByLabelText("対象月")).toBeInTheDocument();
    expect(screen.queryByLabelText("開始日")).not.toBeInTheDocument();
  });

  it("月次サマリーカードに前月繰越・当月未払い・次月繰越が表示される", async () => {
    renderTab("group_by=monthly&month=2026-06");

    // サマリーカードとテーブルヘッダーに同一テキストが存在するため getAllByText を使用
    await waitFor(() => {
      expect(screen.getAllByText("前月繰越").length).toBeGreaterThanOrEqual(1);
    });

    expect(screen.getAllByText("当月未払い").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("次月繰越").length).toBeGreaterThanOrEqual(1);
    // サマリーのみに存在する金額で検証
    expect(screen.getByText("¥7,000")).toBeInTheDocument();
    expect(screen.getByText("¥10,000")).toBeInTheDocument();
  });

  it("月次テーブルに飼主名・ペット名・3金額列が表示される", async () => {
    renderTab("group_by=monthly&month=2026-06");

    await waitFor(() => {
      expect(screen.getByText("山田花子")).toBeInTheDocument();
    });

    expect(screen.getByText("チョコ")).toBeInTheDocument();
    expect(screen.getByText("鈴木一郎")).toBeInTheDocument();

    const headers = screen.getAllByRole("columnheader");
    const headerTexts = headers.map((h) => h.textContent);
    expect(headerTexts).toContain("飼主名");
    expect(headerTexts).toContain("ペット名");
    expect(headerTexts).toContain("前月繰越");
    expect(headerTexts).toContain("当月未払い");
    expect(headerTexts).toContain("次月繰越");
  });

  it("月次行は非interactiveで、飼主詳細への固有名44px native linkを使う", async () => {
    renderTab("group_by=monthly&month=2026-06");

    const link = await screen.findByRole("link", {
      name: "飼主詳細: 山田花子 (ID 1)",
    });
    expect(link.tagName).toBe("A");
    expect(link).toHaveClass("min-h-11", "min-w-11");
    expect(link).toHaveAttribute("href", "/owners/1");
    expect(link.closest("tr")).not.toHaveAttribute("role", "link");
  });

  it("ペット名が空の行は - を表示する", async () => {
    renderTab("group_by=monthly&month=2026-06");

    await waitFor(() => {
      expect(screen.getByText("鈴木一郎")).toBeInTheDocument();
    });

    const rows = screen.getAllByRole("row");
    const suzukiRow = rows.find((r) => r.textContent?.includes("鈴木一郎"));
    expect(suzukiRow?.textContent).toContain("-");
  });

  it("データなしのとき空メッセージを表示する", async () => {
    server.use(
      http.get("/api/v1/accountings/unpaid-monthly", () =>
        HttpResponse.json({ data: [], total: 0, page: 1, limit: 20, summary: { prev_month_carryover: 0, current_month_unpaid: 0, next_month_carryover: 0 } }),
      ),
    );

    renderTab("group_by=monthly&month=2026-06");

    await waitFor(() => {
      expect(screen.getByText("対象月の未納データがありません")).toBeInTheDocument();
    });
  });

  it("月次モードでは飼主単位クエリは発火しない", async () => {
    let ownerHit = false;
    server.use(
      http.get("/api/v1/accountings/unpaid", () => {
        ownerHit = true;
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 20, summary: { total_amount: 0, billing_count: 0, owner_count: 0 } });
      }),
    );

    renderTab("group_by=monthly&month=2026-06");

    await screen.findByText("山田花子");
    expect(ownerHit).toBe(false);
  });
});

describe("UnpaidTab — 明細遷移", () => {
  beforeEach(() => {
    server.use(
      http.get("/api/v1/accountings/unpaid-monthly", () =>
        HttpResponse.json({
          data: [],
          total: 0,
          page: 1,
          limit: 20,
          summary: {
            prev_month_carryover: 0,
            current_month_unpaid: 0,
            next_month_carryover: 0,
          },
        }),
      ),
    );
  });

  it("飼主単位は行ではなく固有名の44px native linkで飼主詳細へ遷移する", async () => {
    server.use(
      http.get("/api/v1/accountings/unpaid", () =>
        HttpResponse.json({
          data: [
            {
              owner_id: 31,
              owner_name: "山田花子",
              count: 2,
              total_amount: 3000,
              oldest_scheduled: "2026-06-01",
              latest_scheduled: "2026-06-20",
            },
          ],
          total: 1,
          page: 1,
          limit: 20,
          summary: { total_amount: 3000, billing_count: 2, owner_count: 1 },
        }),
      ),
    );
    renderTab("group_by=owner&start_date=2026-06-01&end_date=2026-06-30");

    const link = await screen.findByRole("link", {
      name: "飼主詳細: 山田花子 (ID 31)",
    });
    expect(link.tagName).toBe("A");
    expect(link).toHaveClass("min-h-11", "min-w-11");
    expect(link).toHaveAttribute("href", "/owners/31");
  });

  it("会計単位は行ではなく固有名の44px native linkで会計詳細へ遷移する", async () => {
    server.use(
      http.get("/api/v1/accountings/unpaid", () =>
        HttpResponse.json({
          data: [
            {
              id: 77,
              clinic_id: 1,
              owner_id: 31,
              pet_id: 41,
              status: "pending",
              scheduled_date: "2026-06-10T00:00:00+09:00",
              owner: { name: "山田花子" },
              pet: { name: "チョコ" },
              items: [
                {
                  id: 1,
                  category: "other",
                  name: "物販",
                  unit_price: 1000,
                  quantity: 1,
                  discount_rate: 0,
                  discount_amount: 0,
                  tax_type: "excluded",
                  tax_rate: 0.1,
                  is_insurance_applicable: false,
                  source: "manual",
                },
              ],
              total_refunded_amount: 0,
            },
          ],
          total: 1,
          page: 1,
          limit: 20,
        }),
      ),
    );
    renderTab("group_by=billing&start_date=2026-06-01&end_date=2026-06-30");

    const link = await screen.findByRole("link", {
      name: "会計詳細: 山田花子 / チョコ (ID 77)",
    });
    expect(link.tagName).toBe("A");
    expect(link).toHaveClass("min-h-11", "min-w-11");
    expect(link).toHaveAttribute("href", "/accounting/77");
  });
});
