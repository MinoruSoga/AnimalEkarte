import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/hooks/auth-context";
import { axios } from "@/lib/axios";
import { LstepAnalyticsPage } from "./LstepAnalyticsPage";
import type { MonthlyDeliveryStatsResponse } from "../api/get-lstep-delivery-stats";
import type { LstepCsvImportItem } from "../api/get-lstep-csv-imports";
import type { VisitConversionSummaryResponse } from "../api/get-lstep-visit-conversion";

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

const mockStats: MonthlyDeliveryStatsResponse = {
  year_month: "2026-05",
  rows: [
    { trigger_type: "birthday_message", status: "fired", count: 10 },
    { trigger_type: "birthday_message", status: "excluded", count: 2 },
    { trigger_type: "birthday_message", status: "failed", count: 1 },
    { trigger_type: "birthday_message", status: "scheduled", count: 0 },
    { trigger_type: "next_visit_reminder", status: "fired", count: 5 },
    { trigger_type: "next_visit_reminder", status: "excluded", count: 0 },
    { trigger_type: "next_visit_reminder", status: "failed", count: 0 },
    { trigger_type: "next_visit_reminder", status: "scheduled", count: 3 },
  ],
};

const mockStatsEmpty: MonthlyDeliveryStatsResponse = {
  year_month: "2026-04",
  rows: [],
};

const mockCsvImports: LstepCsvImportItem[] = [
  {
    id: "import-1",
    clinic_id: CLINIC_ID,
    csv_type: "friend_attributes",
    file_name: "friends_2026_05.csv",
    row_count: 100,
    success_count: 98,
    error_count: 2,
    status: "completed",
    created_at: "2026-05-10T09:00:00Z",
  },
  {
    id: "import-2",
    clinic_id: CLINIC_ID,
    csv_type: "friend_attributes",
    file_name: "friends_2026_04.csv",
    row_count: 80,
    success_count: 75,
    error_count: 5,
    status: "failed",
    created_at: "2026-04-10T09:00:00Z",
  },
];

const mockVisitConversion: VisitConversionSummaryResponse = {
  year_month: "2026-05",
  days: 30,
  delivered_count: 15,
  visited_count: 5,
  visit_rate: 5 / 15,
  rows: [
    {
      trigger_type: "birthday_message",
      delivered_count: 10,
      visited_count: 4,
      visit_rate: 0.4,
    },
    {
      trigger_type: "next_visit_reminder",
      delivered_count: 5,
      visited_count: 1,
      visit_rate: 0.2,
    },
  ],
};

const mockVisitConversionEmpty: VisitConversionSummaryResponse = {
  year_month: "2026-04",
  days: 30,
  delivered_count: 0,
  visited_count: 0,
  visit_rate: 0,
  rows: [],
};

function setupStatsHandler(data: MonthlyDeliveryStatsResponse) {
  server.use(
    http.get(
      `/api/v1/clinics/${CLINIC_ID}/lstep/analytics/delivery-stats`,
      () => HttpResponse.json(data)
    )
  );
}

function setupCsvImportsHandler(data: LstepCsvImportItem[]) {
  server.use(
    http.get(
      `/api/v1/clinics/${CLINIC_ID}/lstep/csv-imports`,
      () => HttpResponse.json(data)
    )
  );
}

function setupVisitConversionHandler(data: VisitConversionSummaryResponse) {
  server.use(
    http.get(
      `/api/v1/clinics/${CLINIC_ID}/lstep/analytics/visit-conversion`,
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
  stats: MonthlyDeliveryStatsResponse = mockStats,
  csvImports: LstepCsvImportItem[] = mockCsvImports,
  visitConversion: VisitConversionSummaryResponse = mockVisitConversion
) {
  setupStatsHandler(stats);
  setupCsvImportsHandler(csvImports);
  setupVisitConversionHandler(visitConversion);
  render(<LstepAnalyticsPage />, { wrapper: createWrapper() });
  await waitFor(() => {
    expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
  });
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem("auth_current_clinic:v1");
});

// ─────────────────────────────────────────────────────────────
// A: セクション描画
// ─────────────────────────────────────────────────────────────

describe("LstepAnalyticsPage — A: セクション描画", () => {
  it("「月次配信統計」セクション見出しが表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("月次配信統計")).toBeInTheDocument();
  });

  it("「友だち属性 CSV インポート」セクション見出しが表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("友だち属性 CSV インポート")).toBeInTheDocument();
  });

  it("年月セレクトが描画される", async () => {
    await renderAndWait();
    expect(screen.getByLabelText("集計対象年月")).toHaveClass("min-h-11");
  });

  it("CSV選択を44px以上のsemantic label proxyにする", async () => {
    await renderAndWait();
    const input = screen.getByLabelText("友だち属性CSVファイルを選択");
    const uploadRow = screen.getByTestId("csv-upload-row");
    const uploadControls = screen.getByTestId("csv-upload-controls");
    const fileName = screen.getByTestId("csv-selected-file-name");

    expect(input).toHaveClass("absolute", "inset-0", "size-full");
    expect(uploadRow).toHaveClass("flex-col", "sm:flex-row", "sm:items-center");
    expect(uploadControls).toHaveClass("w-full", "min-w-0", "sm:w-auto");
    expect(fileName).toHaveClass("min-w-0", "flex-1", "truncate", "sm:max-w-40");
    expect(input.closest("label")).toHaveClass(
      "min-h-11",
      "min-w-11",
      "focus-within:ring-ring",
    );
  });
});

// ─────────────────────────────────────────────────────────────
// B: 配信統計テーブル
// ─────────────────────────────────────────────────────────────

describe("LstepAnalyticsPage — B: 配信統計テーブル", () => {
  it("トリガー別の行が描画される (誕生日メッセージ・次回来院リマインド)", async () => {
    await renderAndWait();
    const statsSection = screen.getByRole("region", { name: "月次配信統計" });
    expect(within(statsSection).getByText("誕生日メッセージ")).toBeInTheDocument();
    expect(within(statsSection).getByText("次回来院リマインド")).toBeInTheDocument();
  });

  it("ステータスヘッダー「送信済」「除外」「失敗」「予定」が表示される", async () => {
    await renderAndWait();
    // 「失敗」は CSV 履歴バッジにも出現するため stats セクション内に絞る
    const statsSection = screen.getByRole("region", { name: "月次配信統計" });
    expect(within(statsSection).getByText("送信済")).toBeInTheDocument();
    expect(within(statsSection).getByText("除外")).toBeInTheDocument();
    expect(within(statsSection).getByText("失敗")).toBeInTheDocument();
    expect(within(statsSection).getByText("予定")).toBeInTheDocument();
  });

  it("データなしのとき「この月のデータはありません」が表示される", async () => {
    await renderAndWait(mockStatsEmpty);
    expect(screen.getByText("この月のデータはありません")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// C: 年月フィルター
// ─────────────────────────────────────────────────────────────

describe("LstepAnalyticsPage — C: 年月フィルター", () => {
  it("年月を変更すると API が再リクエストされる", async () => {
    let requestedYearMonth = "";
    server.use(
      http.get(
        `/api/v1/clinics/${CLINIC_ID}/lstep/analytics/delivery-stats`,
        ({ request }) => {
          const url = new URL(request.url);
          requestedYearMonth = url.searchParams.get("year_month") ?? "";
          return HttpResponse.json(mockStatsEmpty);
        }
      )
    );
    setupCsvImportsHandler(mockCsvImports);
    setupVisitConversionHandler(mockVisitConversionEmpty);

    render(<LstepAnalyticsPage />, { wrapper: createWrapper() });
    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    const select = screen.getByLabelText("集計対象年月");
    const options = Array.from(select.querySelectorAll("option"));
    // 1 つ前の月を選択
    const prevMonthOption = options[1];
    const user = userEvent.setup();
    await user.selectOptions(select, prevMonthOption.value);

    await waitFor(() => {
      expect(requestedYearMonth).toBe(prevMonthOption.value);
    });
  });
});

// ─────────────────────────────────────────────────────────────
// F: 配信後来院率
// ─────────────────────────────────────────────────────────────

describe("LstepAnalyticsPage — F: 配信後来院率", () => {
  it("来院率セクション見出しと全体カードが表示される", async () => {
    await renderAndWait();
    const conversionSection = screen.getByRole("region", { name: "配信後来院率" });
    expect(within(conversionSection).getByText("配信後来院率")).toBeInTheDocument();
    const [deliveredLabel] = within(conversionSection).getAllByText("送信件数");
    const [visitedLabel] = within(conversionSection).getAllByText("来院件数");
    const [rateLabel] = within(conversionSection).getAllByText("来院率");

    const deliveredCard = deliveredLabel.closest("div");
    const visitedCard = visitedLabel.closest("div");
    const rateCard = rateLabel.closest("div");

    expect(deliveredCard).not.toBeNull();
    expect(visitedCard).not.toBeNull();
    expect(rateCard).not.toBeNull();

    expect(within(deliveredCard as HTMLElement).getByText("15")).toBeInTheDocument();
    expect(within(visitedCard as HTMLElement).getByText("5")).toBeInTheDocument();
    expect(within(rateCard as HTMLElement).getByText("33.3%")).toBeInTheDocument();
  });

  it("トリガー別の来院率行が表示される", async () => {
    await renderAndWait();
    const conversionSection = screen.getByRole("region", { name: "配信後来院率" });
    expect(within(conversionSection).getByText("誕生日メッセージ")).toBeInTheDocument();
    expect(within(conversionSection).getByText("次回来院リマインド")).toBeInTheDocument();
    expect(within(conversionSection).getByText("40.0%")).toBeInTheDocument();
    expect(within(conversionSection).getByText("20.0%")).toBeInTheDocument();
  });

  it("データなしのとき空状態が表示される", async () => {
    await renderAndWait(mockStats, mockCsvImports, mockVisitConversionEmpty);
    expect(screen.getByText("この月の来院率データはありません")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// D: CSV インポート履歴テーブル
// ─────────────────────────────────────────────────────────────

describe("LstepAnalyticsPage — D: CSV インポート履歴", () => {
  it("ファイル名が表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("friends_2026_05.csv")).toBeInTheDocument();
    expect(screen.getByText("friends_2026_04.csv")).toBeInTheDocument();
  });

  it("ステータス「完了」「失敗」が日本語で表示される", async () => {
    await renderAndWait();
    // 「失敗」は CSV テーブルヘッダーにも出現するため、履歴行内のステータスに絞る
    const csvSection = screen.getByRole("region", { name: "友だち属性 CSV インポート" });
    const rows = within(csvSection).getAllByRole("row");
    expect(within(rows[1]).getByText("完了")).toBeInTheDocument();
    expect(within(rows[2]).getByText("失敗")).toBeInTheDocument();
  });

  it("インポート履歴が 0 件のとき「インポート履歴はありません」が表示される", async () => {
    await renderAndWait(mockStats, []);
    expect(screen.getByText("インポート履歴はありません")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// E: CSV アップロード
// ─────────────────────────────────────────────────────────────

describe("LstepAnalyticsPage — E: CSV アップロード", () => {
  it("ファイルなしで送信すると「CSVファイルを選択してください」エラーが表示される", async () => {
    await renderAndWait();
    const user = userEvent.setup();
    const submitBtn = screen.getByRole("button", { name: "アップロード" });
    await user.click(submitBtn);
    await waitFor(() => {
      expect(
        screen.getByText("CSVファイルを選択してください")
      ).toBeInTheDocument();
    });
  });

  it("アップロード成功時に「アップロードが完了しました」が表示される", async () => {
    setupStatsHandler(mockStats);
    setupCsvImportsHandler(mockCsvImports);
    setupVisitConversionHandler(mockVisitConversion);

    // axios.post を直接モック: sanitizeNullBytes が FormData を破壊する問題を回避
    const postSpy = vi
      .spyOn(axios, "post")
      .mockResolvedValueOnce({ data: { id: "new-import" }, status: 201 } as never);

    render(<LstepAnalyticsPage />, { wrapper: createWrapper() });
    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    const user = userEvent.setup();
    const file = new File(["line_id,tag\n12345,VIP"], "test.csv", {
      type: "text/csv",
    });
    const fileInput = screen.getByLabelText("友だち属性CSVファイルを選択");
    await user.upload(fileInput, file);

    const submitBtn = screen.getByRole("button", { name: "アップロード" });
    await user.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText("アップロードが完了しました")).toBeInTheDocument();
    }, { timeout: 3000 });

    postSpy.mockRestore();
  });
});
