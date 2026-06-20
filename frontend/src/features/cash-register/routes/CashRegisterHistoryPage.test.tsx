import { describe, it, expect, afterEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import type { CashRegisterClose } from "@/types/generated/models";
import { CashRegisterHistoryPage } from "./CashRegisterHistoryPage";

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

const renderPage = (initialEntry: string) => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialEntry]}>
        <CashRegisterHistoryPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
};

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

  it("scopes the month query and highlights the close matching the ?date param", async () => {
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

    // The deep-link initialized the query to the target month.
    expect(capturedParams?.get("year")).toBe("2026");
    expect(capturedParams?.get("month")).toBe("6");

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

describe("CashRegisterHistoryPage detail dialog", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("opens a detail dialog with memo and category subtotals when a row is clicked", async () => {
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

    await user.click(
      screen.getByRole("button", { name: "2026-06-15 午前 の締め詳細を表示" }),
    );

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
      screen.getByRole("button", { name: "2026-06-15 午前 の締め詳細を表示" }),
    );

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("内訳データなし")).toBeInTheDocument();
  });
});
