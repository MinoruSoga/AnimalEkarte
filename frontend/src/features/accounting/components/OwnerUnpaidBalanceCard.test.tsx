import { describe, it, expect, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";

import { OwnerUnpaidBalanceCard } from "./OwnerUnpaidBalanceCard";

function renderCard(ownerId: string, clinicId = "1") {
  return render(<OwnerUnpaidBalanceCard ownerId={ownerId} clinicId={clinicId} />, {
    wrapper: createTestWrapper(),
  });
}

afterEach(() => {
  localStorage.clear();
});

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

describe("OwnerUnpaidBalanceCard クリニックスコープ (#186 review P2-11)", () => {
  it("グローバル選択クリニックと異なる clinicId prop を渡した場合、X-Clinic-ID は clinicId prop の値を送信する", async () => {
    // グローバル選択クリニックは "1"
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.get("*/v1/accountings/unpaid-balance", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json({ unpaid_total: 500, unpaid_count: 1 });
      }),
    );

    // 対象会計のクリニックは "2"（グローバル選択とは異なる拠点横断ケース）
    renderCard("42", "2");
    await screen.findByTestId("owner-unpaid-balance");

    expect(receivedClinicHeader).toBe("2");
  });
});
