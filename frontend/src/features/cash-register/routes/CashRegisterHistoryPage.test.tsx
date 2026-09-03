import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import type { CashRegisterClose } from "@/types/generated/models";
import { CashRegisterHistoryPage } from "./CashRegisterHistoryPage";

// PageLayout の resource prop 経由で PermissionBadges → usePermission → useAuth が呼ばれるため、
// AuthProvider 非依存でレンダーできるよう最小限のモックを用意する（AccountingReportsPage.test.tsx と同パターン）。
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

interface CloseOverrides {
  memo?: string;
  category_breakdown?: unknown;
}

const makeClose = (
  id: number,
  closeDate: string,
  period: string,
  overrides: CloseOverrides = {},
): CashRegisterClose => ({
  id,
  clinic_id: 1,
  close_date: closeDate,
  period,
  theoretical_cash: 10000,
  actual_cash: 10000,
  cash_difference: 0,
  category_breakdown: overrides.category_breakdown ?? null,
  memo: overrides.memo ?? "",
  closed_by: 1,
  closed_at: "2026-06-15T18:00:00Z",
  created_at: "2026-06-15T18:00:00Z",
  updated_at: "2026-06-15T18:00:00Z",
});

const mockCloses = [makeClose(1, "2026-06-15", "am"), makeClose(2, "2026-06-20", "pm")];

const renderPage = (initialEntry: string) =>
  render(<CashRegisterHistoryPage />, {
    wrapper: createTestWrapper({ initialEntries: [initialEntry] }),
  });

const stubCloses = (closes: CashRegisterClose[]) => {
  server.use(
    http.get("*/v1/cash-register/closes", () =>
      HttpResponse.json({ data: closes, total: closes.length }),
    ),
  );
};

describe("CashRegisterHistoryPage drill-down deep-link", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("scopes the query to the single ?date day and highlights the matching close", async () => {
    let capturedParams: URLSearchParams | null = null;
    server.use(
      http.get("*/v1/cash-register/closes", ({ request }) => {
        capturedParams = new URL(request.url).searchParams;
        return HttpResponse.json({ data: mockCloses, total: mockCloses.length });
      }),
    );

    renderPage("/accounting/close/history?date=2026-06-15");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    // The deep-link scopes the query to that single day (start_date = end_date = date).
    expect(capturedParams?.get("start_date")).toBe("2026-06-15");
    expect(capturedParams?.get("end_date")).toBe("2026-06-15");

    // A contextual note explains the highlight.
    expect(screen.getByText(/ハイライト表示しています/)).toBeInTheDocument();

    // The matching row is highlighted, the other is not.
    const table = screen.getByRole("table");
    const targetRow = within(table).getByText("2026-06-15").closest("tr");
    const otherRow = within(table).getByText("2026-06-20").closest("tr");
    expect(targetRow).toHaveAttribute("data-highlighted", "true");
    expect(otherRow).not.toHaveAttribute("data-highlighted");
  });

  it("renders no highlight when no date param is provided", async () => {
    stubCloses(mockCloses);

    renderPage("/accounting/close/history");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.queryByText(/ハイライト表示しています/)).not.toBeInTheDocument();
    const table = screen.getByRole("table");
    const row = within(table).getByText("2026-06-15").closest("tr");
    expect(row).not.toHaveAttribute("data-highlighted");
  });
});

describe("CashRegisterHistoryPage period filter", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("narrows the visible rows to the selected period (AM/PM/EMG)", async () => {
    const user = userEvent.setup();
    stubCloses([
      makeClose(1, "2026-06-15", "am"),
      makeClose(2, "2026-06-20", "pm"),
      makeClose(3, "2026-06-25", "emg"),
    ]);

    renderPage("/accounting/close/history");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    // All three periods are visible by default.
    expect(screen.getByText("2026-06-15")).toBeInTheDocument();
    expect(screen.getByText("2026-06-20")).toBeInTheDocument();
    expect(screen.getByText("2026-06-25")).toBeInTheDocument();

    await user.selectOptions(screen.getByLabelText("区分"), "emg");

    // Only the EMG close remains.
    expect(screen.queryByText("2026-06-15")).not.toBeInTheDocument();
    expect(screen.queryByText("2026-06-20")).not.toBeInTheDocument();
    expect(screen.getByText("2026-06-25")).toBeInTheDocument();
  });

  it("shows the empty state when the period filter matches no rows", async () => {
    const user = userEvent.setup();
    stubCloses([makeClose(1, "2026-06-15", "am")]);

    renderPage("/accounting/close/history");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    await user.selectOptions(screen.getByLabelText("区分"), "pm");

    expect(screen.queryByText("2026-06-15")).not.toBeInTheDocument();
    expect(screen.getByText("締め履歴がありません")).toBeInTheDocument();
  });
});

describe("CashRegisterHistoryPage month-range query", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("derives start_date/end_date from the selected month and resets to page 1", async () => {
    const user = userEvent.setup();
    let latestParams: URLSearchParams | null = null;
    server.use(
      http.get("*/v1/cash-register/closes", ({ request }) => {
        latestParams = new URL(request.url).searchParams;
        return HttpResponse.json({ data: mockCloses, total: mockCloses.length });
      }),
    );

    renderPage("/accounting/close/history");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    // Read the component's own default (JST) year to avoid timezone coupling in CI.
    const currentYear = (screen.getByLabelText("年") as HTMLSelectElement).value;
    // August always has 31 days, so the expected range is stable regardless of the run year.
    await user.selectOptions(screen.getByLabelText("月"), "8");

    await waitFor(() => {
      expect(latestParams?.get("start_date")).toBe(`${currentYear}-08-01`);
    });
    expect(latestParams?.get("end_date")).toBe(`${currentYear}-08-31`);
    expect(latestParams?.get("page")).toBe("1");
    // The legacy year/month params must no longer be sent.
    expect(latestParams?.get("year")).toBeNull();
    expect(latestParams?.get("month")).toBeNull();
  });
});

describe("CashRegisterHistoryPage pagination", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("shows the pager only when the total exceeds one page and sends the requested page", async () => {
    const user = userEvent.setup();
    let latestParams: URLSearchParams | null = null;
    server.use(
      http.get("*/v1/cash-register/closes", ({ request }) => {
        latestParams = new URL(request.url).searchParams;
        return HttpResponse.json({ data: mockCloses, total: 45 });
      }),
    );

    renderPage("/accounting/close/history");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    // 45 closes / 20 per page = 3 pages, so the page-2 control is present.
    await user.click(screen.getByRole("button", { name: "2" }));

    await waitFor(() => {
      expect(latestParams?.get("page")).toBe("2");
    });
    expect(latestParams?.get("limit")).toBe("20");
  });

  it("hides the pager when the total fits on a single page", async () => {
    stubCloses(mockCloses);

    renderPage("/accounting/close/history");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    expect(screen.queryByText(/件中/)).not.toBeInTheDocument();
  });
});

describe("CashRegisterHistoryPage detail dialog", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("opens a detail dialog only from the named 44px native button", async () => {
    const user = userEvent.setup();
    stubCloses([
      makeClose(1, "2026-06-15", "am", {
        memo: "現金過不足あり",
        category_breakdown: {
          categories: {
            examination: { 現金: 5000 },
            food: { クレジットカード: 1500 },
          },
        },
      }),
    ]);

    renderPage("/accounting/close/history");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    const detailButton = screen.getByRole("button", {
      name: "締め詳細: 2026-06-15 午前 (ID 1)",
    });
    expect(detailButton.tagName).toBe("BUTTON");
    expect(detailButton).toHaveClass("min-h-11", "min-w-11");
    expect(detailButton.closest("tr")).not.toHaveAttribute("role");
    expect(detailButton.closest("tr")).not.toHaveAttribute("tabindex");

    await user.click(detailButton);

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("2026-06-15 午前 の締め詳細")).toBeInTheDocument();
    expect(within(dialog).getByText("現金過不足あり")).toBeInTheDocument();
    // Category subtotals are grouped and labelled (examination → 診療).
    expect(within(dialog).getByText("診療")).toBeInTheDocument();
    expect(within(dialog).getByText("¥5,000")).toBeInTheDocument();
    expect(within(dialog).getByText("フード")).toBeInTheDocument();
    expect(within(dialog).getByText("¥1,500")).toBeInTheDocument();
  });

  it("falls back to a no-data notice when the record has no breakdown", async () => {
    const user = userEvent.setup();
    stubCloses([makeClose(1, "2026-06-15", "am")]);

    renderPage("/accounting/close/history");

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });

    await user.click(
      screen.getByRole("button", { name: "締め詳細: 2026-06-15 午前 (ID 1)" }),
    );

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("内訳データなし")).toBeInTheDocument();
  });

  it("DEC-40: 新 snapshot は未分類・要確認件数を表示する", async () => {
    const user = userEvent.setup();
    stubCloses([
      makeClose(1, "2026-06-15", "am", {
        category_breakdown: {
          categories: { other: { 現金: 800 } },
          unclassified_other_count: 2,
        },
      }),
    ]);

    renderPage("/accounting/close/history");
    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });
    await user.click(
      screen.getByRole("button", { name: "締め詳細: 2026-06-15 午前 (ID 1)" }),
    );

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("未分類・要確認")).toBeInTheDocument();
    expect(within(dialog).getByText("2件")).toBeInTheDocument();
    expect(within(dialog).queryByText("記録なし")).not.toBeInTheDocument();
  });

  it("DEC-40: 旧 snapshot は未分類・要確認件数を「記録なし」と表示する（0 や live 再計算にしない）", async () => {
    const user = userEvent.setup();
    stubCloses([
      makeClose(1, "2026-06-15", "am", {
        category_breakdown: {
          categories: { other: { 現金: 500 } },
          // unclassified_other_count 欠落
        },
      }),
    ]);

    renderPage("/accounting/close/history");
    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });
    await user.click(
      screen.getByRole("button", { name: "締め詳細: 2026-06-15 午前 (ID 1)" }),
    );

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("未分類・要確認")).toBeInTheDocument();
    expect(within(dialog).getByText("記録なし")).toBeInTheDocument();
    expect(within(dialog).queryByText("0件")).not.toBeInTheDocument();
  });
});

describe("CashRegisterHistoryPage DESIGN.md table contract", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("header/body の typography と cell padding が DESIGN.md に一致する", async () => {
    stubCloses([makeClose(1, "2026-06-15", "am")]);
    renderPage("/accounting/close/history");

    const table = await screen.findByRole("table");
    const headers = within(table).getAllByRole("columnheader");
    const cells = within(table).getAllByRole("cell");

    for (const header of headers) {
      expect(header).toHaveClass("px-4", "py-3", "text-2xs", "font-semibold");
    }
    expect(table).toHaveClass("text-sm");
    for (const cell of cells) {
      expect(cell).toHaveClass("px-4", "py-3");
    }
  });
});
