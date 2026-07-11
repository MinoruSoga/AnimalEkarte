import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/utils";

import { OwnerUnpaidBalanceCard } from "./OwnerUnpaidBalanceCard";

function renderCard(ownerId: string) {
  return render(<OwnerUnpaidBalanceCard ownerId={ownerId} />, { wrapper: createTestWrapper() });
}

describe("OwnerUnpaidBalanceCard 未納残高表示 (#182)", () => {
  it("未納がある場合は残高と件数を表示する", async () => {
    server.use(
      http.get("*/v1/accountings/unpaid-balance", () =>
        HttpResponse.json({ unpaid_total: 12345, unpaid_count: 3 }),
      ),
    );
    renderCard("42");
    const card = await screen.findByTestId("owner-unpaid-balance");
    expect(card.textContent).toContain("未納残高");
    expect(card.textContent).toContain("¥12,345");
    expect(card.textContent).toContain("3 件");
  });

  it("未納が0件のときは何も表示しない（空値ケース）", async () => {
    server.use(
      http.get("*/v1/accountings/unpaid-balance", () =>
        HttpResponse.json({ unpaid_total: 0, unpaid_count: 0 }),
      ),
    );
    renderCard("42");
    // データ取得完了後も残高カードは描画されない
    await waitFor(() => {
      expect(screen.queryByTestId("owner-unpaid-balance")).not.toBeInTheDocument();
    });
  });

  it("ownerId が空のときはクエリを実行せず何も表示しない", () => {
    renderCard("");
    expect(screen.queryByTestId("owner-unpaid-balance")).not.toBeInTheDocument();
  });
});
