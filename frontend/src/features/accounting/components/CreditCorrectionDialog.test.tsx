import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";

import { CreditCorrectionDialog } from "./CreditCorrectionDialog";
import type { Accounting } from "../types";

// BUG-008/018: 理由未入力時のアプリ独自 toast を assert する
vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

function makeAccounting(overrides: Partial<Accounting>): Accounting {
  return {
    id: "10",
    clinicId: "1",
    status: "completed",
    paymentSplits: [
      { id: "1", method: "credit_card", amount: 10000, receivedAmount: 0, changeAmount: 0 },
    ],
    ...overrides,
  } as unknown as Accounting;
}

afterEach(() => {
  localStorage.clear();
  vi.mocked(toast.error).mockClear();
  vi.mocked(toast.success).mockClear();
});

function renderDialog(accounting: Accounting, isPostClose = false, canPostCloseEdit = true) {
  return render(
    <CreditCorrectionDialog
      accounting={accounting}
      isPostClose={isPostClose}
      canPostCloseEdit={canPostCloseEdit}
    />,
    {
      wrapper: createTestWrapper(),
    },
  );
}

describe("CreditCorrectionDialog 表示ゲート (#189)", () => {
  it("確定済み + カード支払い: 訂正導線(ボタン)を表示する", () => {
    renderDialog(makeAccounting({}));
    expect(screen.getByRole("button", { name: "クレジット訂正" })).toBeInTheDocument();
  });

  it("確定前(waiting): カードでも導線を出さない", () => {
    renderDialog(makeAccounting({ status: "waiting" as Accounting["status"] }));
    expect(screen.queryByRole("button", { name: "クレジット訂正" })).not.toBeInTheDocument();
  });

  it("確定済みだが現金のみ: 導線を出さない", () => {
    renderDialog(
      makeAccounting({
        paymentSplits: [
          { id: "1", method: "cash", amount: 10000, receivedAmount: 10000, changeAmount: 0 },
        ] as Accounting["paymentSplits"],
      }),
    );
    expect(screen.queryByRole("button", { name: "クレジット訂正" })).not.toBeInTheDocument();
  });
});

describe("CreditCorrectionDialog 締め済み警告 (#189/M-2)", () => {
  it("isPostClose=true: ダイアログ内に締め済み期間の警告を表示する", async () => {
    const user = userEvent.setup();
    renderDialog(makeAccounting({}), true);
    await user.click(screen.getByRole("button", { name: "クレジット訂正" }));
    expect(await screen.findByText(/レジ締め確定済み期間です/)).toBeInTheDocument();
  });

  it("isPostClose=false: 警告を表示しない", async () => {
    const user = userEvent.setup();
    renderDialog(makeAccounting({}));
    await user.click(screen.getByRole("button", { name: "クレジット訂正" }));
    await screen.findByLabelText("訂正理由（必須）");
    expect(screen.queryByText(/レジ締め確定済み期間です/)).not.toBeInTheDocument();
  });
});

describe("CreditCorrectionDialog 送信 (#189)", () => {
  it("専用 credit-correction エンドポイントを呼び、通常の PATCH(updateAccounting) は呼ばない", async () => {
    let captured: { method?: string; amount?: number; reason?: string } | null = null;
    let patchCalled = false;
    server.use(
      http.post("*/v1/accountings/10/credit-correction", async ({ request }) => {
        captured = (await request.json()) as { method: string; amount: number; reason: string };
        return HttpResponse.json({
          id: 10,
          clinic_id: 1,
          status: "completed",
          payment_splits: [{ id: 1, method: "credit_card", amount: 12000 }],
          payments: [{ billing_id: 10, method: "credit_card", billing_amount: 12000 }],
        });
      }),
      http.patch("*/v1/accountings/10", () => {
        patchCalled = true;
        return HttpResponse.json({});
      }),
    );

    const user = userEvent.setup();
    renderDialog(makeAccounting({}));

    await user.click(screen.getByRole("button", { name: "クレジット訂正" }));
    // ダイアログ内のフォームが開く
    const reason = await screen.findByLabelText("訂正理由（必須）");
    await user.type(reason, "端末への入力金額を打ち間違えたため");
    await user.click(screen.getByRole("button", { name: "訂正を保存" }));

    await waitFor(() => expect(captured).not.toBeNull());
    expect(captured!.method).toBe("credit_card");
    expect(captured!.amount).toBe(10000); // 既定値=現在のカード金額
    expect(captured!.reason).toBe("端末への入力金額を打ち間違えたため");
    expect(patchCalled).toBe(false); // 確定済みカードの訂正は専用経路のみ
  });

  it("理由未入力では送信せず、アプリ独自 toast を出す (BUG-008/018)", async () => {
    let posted = false;
    server.use(
      http.post("*/v1/accountings/10/credit-correction", () => {
        posted = true;
        return HttpResponse.json({
          id: 10,
          clinic_id: 1,
          status: "completed",
          payment_splits: [],
          payments: [],
        });
      }),
    );

    const user = userEvent.setup();
    renderDialog(makeAccounting({}));
    await user.click(screen.getByRole("button", { name: "クレジット訂正" }));
    await screen.findByLabelText("訂正理由（必須）");
    // form に noValidate があること（HTML5 が先にインターセプトしない）
    expect(document.querySelector("form")).toHaveAttribute("novalidate");
    await user.click(screen.getByRole("button", { name: "訂正を保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("訂正理由を入力してください");
    });
    expect(posted).toBe(false);
  });

  it("金額が1円未満のとき toast でブロックし POST しない", async () => {
    let posted = false;
    server.use(
      http.post("*/v1/accountings/10/credit-correction", () => {
        posted = true;
        return HttpResponse.json({
          id: 10,
          clinic_id: 1,
          status: "completed",
          payment_splits: [],
          payments: [],
        });
      }),
    );

    const user = userEvent.setup();
    renderDialog(makeAccounting({}));
    await user.click(screen.getByRole("button", { name: "クレジット訂正" }));
    const amount = await screen.findByLabelText("訂正後の金額");
    await user.clear(amount);
    await user.type(amount, "0");
    await user.type(await screen.findByLabelText("訂正理由（必須）"), "理由あり");
    await user.click(screen.getByRole("button", { name: "訂正を保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("金額は1円以上で入力してください");
    });
    expect(posted).toBe(false);
  });
});

describe("CreditCorrectionDialog 権限再チェック (FE-RC-110)", () => {
  it("canPostCloseEdit=false では POST せず toast する", async () => {
    let posted = false;
    server.use(
      http.post("*/v1/accountings/10/credit-correction", () => {
        posted = true;
        return HttpResponse.json({
          id: 10,
          clinic_id: 1,
          status: "completed",
          payment_splits: [],
          payments: [],
        });
      }),
    );

    const user = userEvent.setup();
    renderDialog(makeAccounting({}), false, false);
    await user.click(screen.getByRole("button", { name: "クレジット訂正" }));
    await user.type(
      await screen.findByLabelText("訂正理由（必須）"),
      "端末への入力金額を打ち間違えたため",
    );
    await user.click(screen.getByRole("button", { name: "訂正を保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(posted).toBe(false);
  });

  it("開いたあと canPostCloseEdit が false になると最新の ref で送信を拒否する", async () => {
    let posted = false;
    server.use(
      http.post("*/v1/accountings/10/credit-correction", () => {
        posted = true;
        return HttpResponse.json({
          id: 10,
          clinic_id: 1,
          status: "completed",
          payment_splits: [],
          payments: [],
        });
      }),
    );

    const accounting = makeAccounting({});
    const user = userEvent.setup();
    const { rerender } = renderDialog(accounting);
    await user.click(screen.getByRole("button", { name: "クレジット訂正" }));
    await user.type(
      await screen.findByLabelText("訂正理由（必須）"),
      "端末への入力金額を打ち間違えたため",
    );

    rerender(
      <CreditCorrectionDialog
        accounting={accounting}
        isPostClose={false}
        canPostCloseEdit={false}
      />,
    );

    await user.click(screen.getByRole("button", { name: "訂正を保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(posted).toBe(false);
  });
});

describe("CreditCorrectionDialog クリニックスコープ (#186 review P2-12)", () => {
  it("グローバル選択クリニックと異なる accounting.clinicId を持つ会計では、X-Clinic-ID に accounting.clinicId を送信する", async () => {
    // グローバル選択クリニックは "1"
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.post("*/v1/accountings/10/credit-correction", async ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json({
          id: 10,
          clinic_id: 2,
          status: "completed",
          payment_splits: [{ id: 1, method: "credit_card", amount: 12000 }],
          payments: [{ billing_id: 10, method: "credit_card", billing_amount: 12000 }],
        });
      }),
    );

    const user = userEvent.setup();
    // 対象会計のクリニックは "2"（グローバル選択とは異なる拠点横断ケース）
    renderDialog(makeAccounting({ clinicId: "2" }));

    await user.click(screen.getByRole("button", { name: "クレジット訂正" }));
    const reason = await screen.findByLabelText("訂正理由（必須）");
    await user.type(reason, "端末への入力金額を打ち間違えたため");
    await user.click(screen.getByRole("button", { name: "訂正を保存" }));

    await waitFor(() => expect(receivedClinicHeader).not.toBeNull());
    expect(receivedClinicHeader).toBe("2");
  });
});
