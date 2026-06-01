import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { RefundSection } from "../RefundSection";

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

const handlers = [http.get("/api/v1/accountings/:id/refunds", () => HttpResponse.json([]))];

// 会計の支払方法内訳（混在: カード + 現金）
const splits = [
  { id: 1, clinicId: "1", billingId: "1", method: "credit_card" as const, amount: 3000, receivedAmount: 3000, changeAmount: 0 },
  { id: 2, clinicId: "1", billingId: "1", method: "cash" as const, amount: 2000, receivedAmount: 2000, changeAmount: 0 },
];

afterEach(() => server.resetHandlers());

describe("RefundSection — 支払方法別返金 (#60)", () => {
  it("支払方法未指定で返金登録すると onRefund の第3引数が undefined", async () => {
    server.use(...handlers);
    const onRefund = vi.fn();
    const user = userEvent.setup({ delay: null });

    render(
      <RefundSection
        accountingId="1"
        totalAmount={10000}
        paymentSplits={splits}
        isRefunding={false}
        onRefund={onRefund}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await user.click(screen.getByRole("button", { name: /返金を登録/ }));

    const amountInput = await screen.findByPlaceholderText("0");
    fireEvent.change(amountInput, { target: { value: "3000" } });

    await user.click(screen.getByRole("button", { name: "登録する" }));

    await waitFor(() => {
      expect(onRefund).toHaveBeenCalledOnce();
    });
    expect(onRefund).toHaveBeenCalledWith(3000, "", undefined);
  }, 15000);

  it("会計で使われた支払方法が選択肢として表示される", async () => {
    server.use(...handlers);
    const user = userEvent.setup({ delay: null });

    render(
      <RefundSection
        accountingId="1"
        totalAmount={10000}
        paymentSplits={splits}
        isRefunding={false}
        onRefund={vi.fn()}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await user.click(screen.getByRole("button", { name: /返金を登録/ }));
    expect(await screen.findByTestId("refund-payment-method-trigger")).toBeInTheDocument();
  }, 15000);

  it("単一支払い（splits 空）では支払方法セレクタを表示しない", async () => {
    server.use(...handlers);
    const user = userEvent.setup({ delay: null });

    render(
      <RefundSection
        accountingId="1"
        totalAmount={10000}
        paymentSplits={[]}
        isRefunding={false}
        onRefund={vi.fn()}
        canEdit={true}
      />,
      { wrapper: createWrapper() },
    );

    await user.click(screen.getByRole("button", { name: /返金を登録/ }));
    await screen.findByPlaceholderText("0");
    expect(screen.queryByTestId("refund-payment-method-trigger")).not.toBeInTheDocument();
  }, 15000);
});
