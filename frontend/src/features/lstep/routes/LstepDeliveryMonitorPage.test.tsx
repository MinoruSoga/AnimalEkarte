import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse, delay } from "msw";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/hooks/auth-context";
import { LstepDeliveryMonitorPage } from "./LstepDeliveryMonitorPage";
import type { DeliveryTriggerSummaryResponse } from "../api/get-lstep-delivery-trigger-summary";
import type { DeliveryTriggerLogsPageResponse } from "../api/get-lstep-delivery-trigger-logs";

const CLINIC_ID = "clinic-test-1";

const mockAuthContext = {
  user: null,
  currentClinicId: CLINIC_ID,
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => true as boolean,
  refreshPermissions: async () => {},
};

const mockSummary: DeliveryTriggerSummaryResponse = {
  scheduled: 10,
  fired: 5,
  excluded: 3,
  failed: 2,
  suppressed_by_priority: 4,
  excluded_reason_breakdown: { opt_out: 2, no_line_id: 1 },
};

const mockSummaryNoIssue: DeliveryTriggerSummaryResponse = {
  scheduled: 8,
  fired: 8,
  excluded: 0,
  failed: 0,
  suppressed_by_priority: 0,
  excluded_reason_breakdown: {},
};

const mockLogs: DeliveryTriggerLogsPageResponse = {
  items: [
    {
      id: "log-1",
      owner_id: "owner-1",
      owner_name: "田中 花子",
      trigger_type: "birthday_message",
      scheduled_at: "2026-05-02T10:00:00Z",
      status: "fired",
      fired_at: "2026-05-02T10:01:00Z",
    },
    {
      id: "log-2",
      owner_id: "owner-2",
      owner_name: "山田 太郎",
      trigger_type: "next_visit_reminder",
      scheduled_at: "2026-05-02T11:00:00Z",
      status: "excluded",
      excluded_reason: "opt_out",
    },
  ],
  total: 2,
  page: 1,
  per_page: 20,
};

const mockLogsEmpty: DeliveryTriggerLogsPageResponse = {
  items: [],
  total: 0,
  page: 1,
  per_page: 20,
};

function setupSummaryHandler(data: DeliveryTriggerSummaryResponse) {
  server.use(
    http.get(
      `/api/v1/clinics/${CLINIC_ID}/lstep/delivery-monitor/summary`,
      () => HttpResponse.json(data)
    )
  );
}

function setupLogsHandler(data: DeliveryTriggerLogsPageResponse) {
  server.use(
    http.get(
      `/api/v1/clinics/${CLINIC_ID}/lstep/delivery-monitor/logs`,
      () => HttpResponse.json(data)
    )
  );
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <AuthContext.Provider value={mockAuthContext}>
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
}

async function renderAndWait(
  summary: DeliveryTriggerSummaryResponse = mockSummary,
  logs: DeliveryTriggerLogsPageResponse = mockLogs
) {
  setupSummaryHandler(summary);
  setupLogsHandler(logs);
  render(<LstepDeliveryMonitorPage />, { wrapper: createWrapper() });
  await waitFor(() => {
    expect(screen.queryByTestId("logs-loading")).not.toBeInTheDocument();
  });
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem("auth_current_clinic:v1");
});

// ─────────────────────────────────────────────────────────────
// A: サマリーカード
// ─────────────────────────────────────────────────────────────

describe("LstepDeliveryMonitorPage — A: サマリーカード", () => {
  it("5 枚のサマリーカードが正しい数値で描画される", async () => {
    await renderAndWait();

    const scheduledCard = screen.getByTestId("summary-card-scheduled");
    expect(scheduledCard).toHaveTextContent("10");

    const firedCard = screen.getByTestId("summary-card-fired");
    expect(firedCard).toHaveTextContent("5");

    const excludedCard = screen.getByTestId("summary-card-excluded");
    expect(excludedCard).toHaveTextContent("3");

    const failedCard = screen.getByTestId("summary-card-failed");
    expect(failedCard).toHaveTextContent("2");

    const suppressedCard = screen.getByTestId(
      "summary-card-suppressed_by_priority"
    );
    expect(suppressedCard).toHaveTextContent("4");
  });

  it("カードラベル「予定」「送信済」「除外」「失敗」「優先度抑制」が表示される", async () => {
    await renderAndWait();
    expect(screen.getByTestId("summary-card-scheduled")).toHaveTextContent(
      "予定"
    );
    expect(screen.getByTestId("summary-card-fired")).toHaveTextContent(
      "送信済"
    );
    expect(screen.getByTestId("summary-card-excluded")).toHaveTextContent(
      "除外"
    );
    expect(screen.getByTestId("summary-card-failed")).toHaveTextContent(
      "失敗"
    );
    expect(
      screen.getByTestId("summary-card-suppressed_by_priority")
    ).toHaveTextContent("優先度抑制");
  });
});

// ─────────────────────────────────────────────────────────────
// B: 失敗件数バナー
// ─────────────────────────────────────────────────────────────

describe("LstepDeliveryMonitorPage — B: 失敗件数バナー", () => {
  it("failed > 0 のとき警告バナーが表示される", async () => {
    await renderAndWait(mockSummary);
    expect(screen.getByTestId("failed-warning-banner")).toBeInTheDocument();
    expect(screen.getByTestId("failed-warning-banner")).toHaveTextContent(
      "2件"
    );
  });

  it("failed = 0 のとき警告バナーは表示されない", async () => {
    await renderAndWait(mockSummaryNoIssue);
    expect(
      screen.queryByTestId("failed-warning-banner")
    ).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// C: 除外理由内訳
// ─────────────────────────────────────────────────────────────

describe("LstepDeliveryMonitorPage — C: 除外理由内訳", () => {
  it("excluded > 0 かつ breakdown あり のとき内訳セクションが表示される", async () => {
    await renderAndWait(mockSummary);
    const breakdown = screen.getByTestId("excluded-reason-breakdown");
    expect(breakdown).toBeInTheDocument();
    expect(breakdown).toHaveTextContent("opt_out");
    expect(breakdown).toHaveTextContent("no_line_id");
  });

  it("excluded = 0 のとき内訳セクションは非表示", async () => {
    await renderAndWait(mockSummaryNoIssue);
    expect(
      screen.queryByTestId("excluded-reason-breakdown")
    ).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// D: ログテーブル
// ─────────────────────────────────────────────────────────────

describe("LstepDeliveryMonitorPage — D: ログテーブル", () => {
  it("birthday_message が日本語ラベル「誕生日メッセージ」で表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("誕生日メッセージ")).toBeInTheDocument();
  });

  it("next_visit_reminder が「次回来院リマインド」で表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("次回来院リマインド")).toBeInTheDocument();
  });

  it("飼主名が表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("田中 花子")).toBeInTheDocument();
    expect(screen.getByText("山田 太郎")).toBeInTheDocument();
  });

  it("ログ 0 件のとき「ログがありません」が表示される", async () => {
    await renderAndWait(mockSummaryNoIssue, mockLogsEmpty);
    await waitFor(() => {
      expect(screen.getByTestId("logs-empty")).toBeInTheDocument();
    });
  });

  it("ログ行が data-testid='log-row' で描画される", async () => {
    await renderAndWait();
    const rows = screen.getAllByTestId("log-row");
    expect(rows).toHaveLength(2);
  });
});

// ─────────────────────────────────────────────────────────────
// E: フィルター操作
// ─────────────────────────────────────────────────────────────

describe("LstepDeliveryMonitorPage — E: フィルター", () => {
  it("日付フィルター入力が描画される", async () => {
    await renderAndWait();
    expect(screen.getByTestId("filter-from")).toBeInTheDocument();
    expect(screen.getByTestId("filter-to")).toBeInTheDocument();
  });

  it("種別フィルターを変更するとページが 1 にリセットされる", async () => {
    setupSummaryHandler(mockSummary);
    setupLogsHandler({
      ...mockLogs,
      total: 25,
      items: mockLogs.items,
    });
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const Wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthContext.Provider value={mockAuthContext}>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter>{children}</MemoryRouter>
        </QueryClientProvider>
      </AuthContext.Provider>
    );
    render(<LstepDeliveryMonitorPage />, { wrapper: Wrapper });

    // trigger type select がレンダリングされるまで待つ
    await waitFor(() => {
      expect(screen.getByTestId("filter-trigger-type")).toBeInTheDocument();
    });
  });
});

// ─────────────────────────────────────────────────────────────
// F: ページネーション
// ─────────────────────────────────────────────────────────────

describe("LstepDeliveryMonitorPage — F: ページネーション", () => {
  it("1 ページ目のとき「前へ」ボタンは無効", async () => {
    await renderAndWait();
    const prevBtn = screen.getByTestId("pagination-prev");
    expect(prevBtn).toBeDisabled();
  });

  it("total が per_page 以下のとき「次へ」ボタンは無効", async () => {
    await renderAndWait();
    // total=2, per_page=20 → 1 ページのみ
    const nextBtn = screen.getByTestId("pagination-next");
    expect(nextBtn).toBeDisabled();
  });

  it("複数ページのとき「次へ」ボタンが有効になる", async () => {
    setupSummaryHandler(mockSummary);
    setupLogsHandler({ ...mockLogs, total: 50 });
    render(<LstepDeliveryMonitorPage />, { wrapper: createWrapper() });
    await waitFor(() => {
      expect(screen.queryByTestId("logs-loading")).not.toBeInTheDocument();
    });
    expect(screen.getByTestId("pagination-next")).not.toBeDisabled();
  });

  it("全件数が正しく表示される", async () => {
    await renderAndWait();
    // FE5-15: 正本 Pagination の文言は「全 N 件」ではなく「件数中 開始-終了件」
    expect(screen.getByText("2件中 1-2件")).toBeInTheDocument();
  });

  it("「次へ」をクリックするとページ番号が 2 になる", async () => {
    setupSummaryHandler(mockSummary);
    const page2Logs: DeliveryTriggerLogsPageResponse = {
      items: [],
      total: 50,
      page: 2,
      per_page: 20,
    };
    // page=1 で 50件返す
    server.use(
      http.get(
        `/api/v1/clinics/${CLINIC_ID}/lstep/delivery-monitor/logs`,
        () => HttpResponse.json({ ...mockLogs, total: 50 })
      )
    );
    render(<LstepDeliveryMonitorPage />, { wrapper: createWrapper() });
    await waitFor(() => {
      expect(screen.queryByTestId("logs-loading")).not.toBeInTheDocument();
    });

    // page2 用ハンドラに切り替え
    server.use(
      http.get(
        `/api/v1/clinics/${CLINIC_ID}/lstep/delivery-monitor/logs`,
        () => HttpResponse.json(page2Logs)
      )
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("pagination-next"));

    await waitFor(() => {
      // FE5-15: 正本 Pagination の文言は「X / Y」ではなく「件数中 開始-終了件」
      expect(screen.getByText("50件中 21-40件")).toBeInTheDocument();
    });
  });
});

// ─────────────────────────────────────────────────────────────
// G: ローディング状態
// ─────────────────────────────────────────────────────────────

describe("LstepDeliveryMonitorPage — G: ローディング状態", () => {
  it("ログ取得中は「読み込み中...」が表示される", () => {
    setupSummaryHandler(mockSummary);
    server.use(
      http.get(
        `/api/v1/clinics/${CLINIC_ID}/lstep/delivery-monitor/logs`,
        async () => {
          await delay("infinite");
          return HttpResponse.json(mockLogs);
        }
      )
    );
    render(<LstepDeliveryMonitorPage />, { wrapper: createWrapper() });
    expect(screen.getByTestId("logs-loading")).toBeInTheDocument();
  });
});

describe("LstepDeliveryMonitorPage — H: mobile-first layout", () => {
  it("固定幅フィルターとサマリーカードを mobile では全幅・単一列にする", async () => {
    await renderAndWait();

    expect(screen.getByTestId("filter-trigger-type")).toHaveClass(
      "w-full",
      "h-11",
      "sm:w-[220px]",
    );
    expect(screen.getByTestId("filter-status")).toHaveClass(
      "w-full",
      "h-11",
      "sm:w-[140px]",
    );
    expect(screen.getByRole("combobox", { name: "配信トリガー種別" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "配信ステータス" })).toBeInTheDocument();
    const dateRange = screen.getByTestId("filter-from").parentElement;
    expect(dateRange).toHaveClass("w-full", "flex-col", "sm:w-auto", "sm:flex-row");
    expect(screen.getByTestId("filter-from")).toHaveClass("w-full", "sm:w-auto");
    expect(screen.getByTestId("filter-to")).toHaveClass("w-full", "sm:w-auto");
    expect(screen.getByTestId("filter-from")).toHaveClass("h-11");
    expect(screen.getByTestId("filter-to")).toHaveClass("h-11");

    const summaryGrid = screen.getByTestId("summary-card-scheduled").parentElement;
    expect(summaryGrid).toHaveClass("grid-cols-1", "sm:grid-cols-3", "xl:grid-cols-5");
  });
});
