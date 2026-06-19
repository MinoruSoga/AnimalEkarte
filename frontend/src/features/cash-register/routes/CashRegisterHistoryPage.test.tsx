import { describe, it, expect, afterEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import type { CashRegisterClose } from "@/types/generated/models";
import { CashRegisterHistoryPage } from "./CashRegisterHistoryPage";

const makeClose = (id: number, closeDate: string, period: string): CashRegisterClose => ({
  id,
  clinic_id: 1,
  close_date: closeDate,
  period,
  theoretical_cash: 10000,
  actual_cash: 10000,
  cash_difference: 0,
  category_breakdown: null,
  memo: "",
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
    server.use(
      http.get("*/v1/cash-register/closes", () =>
        HttpResponse.json({ data: mockCloses, total: mockCloses.length }),
      ),
    );

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
