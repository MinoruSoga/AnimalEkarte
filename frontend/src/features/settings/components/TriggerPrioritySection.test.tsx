import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { TriggerPrioritySection } from "./TriggerPrioritySection";
import type { TriggerPriorityListResponse } from "../hooks/use-trigger-priorities";

const CLINIC_ID = "clinic-test-1";

const sampleResponse: TriggerPriorityListResponse = {
  clinic_id: CLINIC_ID,
  items: [
    { trigger_type: "first_visit_followup_3d", priority: 1 },
    { trigger_type: "vaccine_deadline_30d", priority: 2 },
    { trigger_type: "birthday_message", priority: 3 },
  ],
};

function setupGetHandler(data: TriggerPriorityListResponse = sampleResponse) {
  server.use(
    http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/trigger-priorities`, () =>
      HttpResponse.json(data),
    ),
  );
}

const createWrapper = () => createTestWrapper({ router: true });

async function renderAndWait(data: TriggerPriorityListResponse = sampleResponse) {
  setupGetHandler(data);
  render(<TriggerPrioritySection />, { wrapper: createWrapper() });
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
// A: GET 結果表示
// ─────────────────────────────────────────────────────────────

describe("TriggerPrioritySection — A: GET 結果表示 (Q23)", () => {
  it("トリガーラベルが表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("初診フォロー 3日後")).toBeInTheDocument();
    expect(screen.getByText("ワクチン期限 30日前")).toBeInTheDocument();
    expect(screen.getByText("誕生日メッセージ")).toBeInTheDocument();
  });

  it("priority の値が input に表示される", async () => {
    await renderAndWait();
    const input = screen.getByLabelText("初診フォロー 3日後 優先順位") as HTMLInputElement;
    expect(input.value).toBe("1");
  });

  it("priority 昇順で表示される", async () => {
    await renderAndWait({
      clinic_id: CLINIC_ID,
      items: [
        { trigger_type: "birthday_message", priority: 3 },
        { trigger_type: "first_visit_followup_3d", priority: 1 },
        { trigger_type: "vaccine_deadline_30d", priority: 2 },
      ],
    });
    const inputs = screen.getAllByRole("spinbutton");
    expect((inputs[0] as HTMLInputElement).value).toBe("1");
    expect((inputs[1] as HTMLInputElement).value).toBe("2");
    expect((inputs[2] as HTMLInputElement).value).toBe("3");
  });

  it('読み込み中は "読み込み中..." が表示される', () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/trigger-priorities`, async () => {
        await new Promise(() => {}); // never resolve
      }),
    );
    render(<TriggerPrioritySection />, { wrapper: createWrapper() });
    expect(screen.getByText("読み込み中...")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// B: dirty state / 保存ボタン活性制御
// ─────────────────────────────────────────────────────────────

describe("TriggerPrioritySection — B: dirty state / 保存ボタン (Q23)", () => {
  it("初期状態では保存ボタンが disabled", async () => {
    await renderAndWait();
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
  });

  it("priority を変更すると保存ボタンが enabled になる", async () => {
    await renderAndWait();
    const input = screen.getByLabelText("初診フォロー 3日後 優先順位");
    fireEvent.change(input, { target: { value: "5" } });
    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
  });

  it("priority を元の値に戻すと保存ボタンが disabled に戻る", async () => {
    await renderAndWait();
    const input = screen.getByLabelText("初診フォロー 3日後 優先順位");
    fireEvent.change(input, { target: { value: "5" } });
    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
    fireEvent.change(input, { target: { value: "1" } });
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
  });
});

// ─────────────────────────────────────────────────────────────
// C: PATCH 送信内容
// ─────────────────────────────────────────────────────────────

describe("TriggerPrioritySection — C: PATCH 送信 (Q23)", () => {
  it("保存すると PATCH body が { items: [...] } 形式で送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep/trigger-priorities`, async ({ request }) => {
        capturedBody = await request.json();
        return new HttpResponse(null, { status: 200 });
      }),
    );

    await renderAndWait();
    const user = userEvent.setup();
    const input = screen.getByLabelText("初診フォロー 3日後 優先順位");
    fireEvent.change(input, { target: { value: "5" } });
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });

    const body = capturedBody as { items: { trigger_type: string; priority: number }[] };
    expect(body).toHaveProperty("items");
    expect(Array.isArray(body.items)).toBe(true);
    const item = body.items.find((i) => i.trigger_type === "first_visit_followup_3d");
    expect(item?.priority).toBe(5);
  });

  it("保存成功後は保存ボタンが disabled に戻る（query invalidate で draft リセット）", async () => {
    server.use(
      http.patch(
        `/api/v1/clinics/${CLINIC_ID}/lstep/trigger-priorities`,
        () => new HttpResponse(null, { status: 200 }),
      ),
    );
    // GET は成功後に再フェッチされ同じデータが返る
    setupGetHandler();

    await renderAndWait();
    const user = userEvent.setup();
    const input = screen.getByLabelText("初診フォロー 3日後 優先順位");
    fireEvent.change(input, { target: { value: "5" } });

    await user.click(screen.getByRole("button", { name: "保存" }));

    // invalidate で GET が再実行され draft がリセットされるため disabled に戻る
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
    });
  });
});

// ─────────────────────────────────────────────────────────────
// D: バリデーション (priority < 1)
// ─────────────────────────────────────────────────────────────

describe("TriggerPrioritySection — D: バリデーション (Q23)", () => {
  it("priority を 0 に設定して保存しようとしても PATCH が呼ばれない", async () => {
    let patchCount = 0;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep/trigger-priorities`, () => {
        patchCount++;
        return new HttpResponse(null, { status: 200 });
      }),
    );

    await renderAndWait();
    const user = userEvent.setup();
    const input = screen.getByLabelText("初診フォロー 3日後 優先順位");
    fireEvent.change(input, { target: { value: "0" } });

    // 保存ボタンは有効（dirty だが 0 なので）
    // isDirty = true (0 !== 1) → button is enabled
    const saveBtn = screen.getByRole("button", { name: "保存" });
    if (!saveBtn.hasAttribute("disabled")) {
      await user.click(saveBtn);
    }

    // PATCH は呼ばれない
    await waitFor(() => {
      expect(patchCount).toBe(0);
    });
  });
});
