import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { AggregationDashboardPage } from "../routes/AggregationDashboardPage";

// EMR-334: 一覧レスポンスに同梱される CPM 区分人数 (cpm_stage_counts)。
// BE は cpm_stage 絞り込み・ページネーション前の母集団を 7 区分すべて (0件含む) 返す。
// per_page=1 で返却行を1件に絞っても、counts は母集団全体 (=4) を表す。
const sharedResponse = {
  owners: [
    {
      owner_id: "owner1",
      owner_name: "田中太郎",
      total_fee: 500000,
      total_visit_count: 20,
      annual_visit_count: 12,
      last_visit_date: "2026-04-20",
      first_visit_date: "2024-01-15",
      annual_amount: 150000,
      billing_count: 8,
      period_visit_count: 5,
      days_since_last_visit: 7,
      last_visit_bucket: "within_3m",
      cpm_stage: "cpm_spot",
    },
  ],
  total: 4,
  page: 1,
  per_page: 1,
  cpm_stage_counts: {
    cpm_encounter: 1,
    cpm_growing: 0,
    cpm_core: 0,
    cpm_spot: 2,
    cpm_noah: 0,
    cpm_dormant: 0,
    cpm_unclassified: 1,
  },
};

describe("aggregation single-request CPM counts (EMR-334)", () => {
  beforeEach(() => {
    localStorage.setItem("auth_current_clinic:v1", "1");
  });

  afterEach(() => {
    localStorage.removeItem("auth_current_clinic:v1");
  });

  it("drives the owner list and CPM chips from one aggregations request", async () => {
    const requestUrls: string[] = [];
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", ({ request }) => {
        requestUrls.push(request.url);
        return HttpResponse.json(sharedResponse);
      }),
    );

    render(<AggregationDashboardPage />, {
      wrapper: createTestWrapper({ initialEntries: ["/aggregation"] }),
    });

    await waitFor(() => {
      expect(screen.getByText("田中太郎")).toBeInTheDocument();
    });

    // 一覧はページネーション済み (per_page=1 → 1行) だが件数バッジは母集団 total。
    expect(screen.getByText("4 件")).toBeInTheDocument();

    // 契約: 一覧 + CPM 人数は同一レスポンス。区分ごとの追加 GET は発行しない。
    // 旧実装は一覧 1 + 各区分 7 = 8 リクエストを並列発行するため RED。
    expect(requestUrls).toHaveLength(1);
    expect(requestUrls[0]).not.toContain("cpm_stage=");

    // 各チップは同一レスポンスの cpm_stage_counts を表示する (unclassified 含む 7 区分)。
    // 旧実装では各区分リクエストの data.total (=4) が全チップに出るため RED。
    expect(screen.getByRole("button", { name: /すべて/ })).toHaveTextContent("4");
    expect(screen.getByRole("button", { name: /Encounter/ })).toHaveTextContent("1");
    expect(screen.getByRole("button", { name: /Spot/ })).toHaveTextContent("2");
    expect(screen.getByRole("button", { name: /Core/ })).toHaveTextContent("0");
    expect(screen.getByRole("button", { name: /Unclassified/ })).toHaveTextContent("1");
  });

  it("renders zero counts for all stages when cpm_stage_counts is empty", async () => {
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", () =>
        HttpResponse.json({ ...sharedResponse, cpm_stage_counts: {} }),
      ),
    );

    render(<AggregationDashboardPage />, {
      wrapper: createTestWrapper({ initialEntries: ["/aggregation"] }),
    });

    await waitFor(() => {
      expect(screen.getByText("田中太郎")).toBeInTheDocument();
    });

    // counts が空でも全チップが 0 表示で壊れない (旧実装は各区分 total=4 を表示するため RED)。
    expect(screen.getByRole("button", { name: /すべて/ })).toHaveTextContent("0");
    expect(screen.getByRole("button", { name: /Spot/ })).toHaveTextContent("0");
    expect(screen.getByRole("button", { name: /Unclassified/ })).toHaveTextContent("0");
  });

  it("shows the dedicated retry message once on 503 aggregation_timeout, without extra requests", async () => {
    const requestUrls: string[] = [];
    server.use(
      http.get("/api/v1/clinics/:clinic_id/owners/aggregations", ({ request }) => {
        requestUrls.push(request.url);
        return HttpResponse.json(
          { error: "aggregation timed out", code: "aggregation_timeout" },
          { status: 503 },
        );
      }),
    );

    render(<AggregationDashboardPage />, {
      wrapper: createTestWrapper({ initialEntries: ["/aggregation"] }),
    });

    await waitFor(() => {
      expect(
        screen.getByText("集計に時間がかかっています。しばらくして再試行してください"),
      ).toBeInTheDocument();
    });

    // 専用メッセージは1回だけ。汎用の「サーバーエラーが発生しました」は出さない。
    expect(
      screen.getAllByText("集計に時間がかかっています。しばらくして再試行してください"),
    ).toHaveLength(1);
    expect(screen.queryByText(/サーバーエラーが発生しました/)).not.toBeInTheDocument();

    // 自動リトライも区分ごとの追加 GET も発行しない。
    expect(requestUrls).toHaveLength(1);
  });
});
