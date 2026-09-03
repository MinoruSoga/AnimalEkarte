import { describe, it, expect, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { RecommendationReasonSelect } from "./RecommendationReasonSelect";
import type { RecommendationReason } from "../constants/recommendation-reason";

const RECORD_ID = "rec-test-1";

// transformMedicalRecord が通せる最低限レスポンス
const minimalMedicalRecordApiResponse = {
  id: 1,
  record_no: "R001",
  date: "2026-05-04",
  pet_id: 1,
  owner_id: 1,
  doctor_id: 1,
  status: "draft",
  visit_type: "再診",
  version: 1,
  created_at: "2026-05-04T00:00:00Z",
  updated_at: "2026-05-04T00:00:00Z",
};

function createWrapper() {
  return createTestWrapper({ router: true });
}

function renderSelect(value: RecommendationReason | null, disabled = false) {
  render(
    <RecommendationReasonSelect
      mode="edit"
      medicalRecordId={RECORD_ID}
      value={value}
      disabled={disabled}
    />,
    { wrapper: createWrapper() },
  );
}

afterEach(() => {
  server.resetHandlers();
});

// ─────────────────────────────────────────────────────────────
// A: 初期表示
// ─────────────────────────────────────────────────────────────

describe("RecommendationReasonSelect — A: 初期表示", () => {
  it("value=null → 「未選択」プレースホルダーが表示される", () => {
    renderSelect(null);
    expect(screen.getByTestId("recommendation-reason-trigger")).toBeInTheDocument();
    expect(screen.getByText("未選択")).toBeInTheDocument();
  });

  it("value='revisit' → 「再診」が表示される", () => {
    renderSelect("revisit");
    expect(screen.getByText("再診")).toBeInTheDocument();
  });

  it("value='exam' → 「検査」が表示される", () => {
    renderSelect("exam");
    expect(screen.getByText("検査")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// B: 選択肢の表示
// ─────────────────────────────────────────────────────────────

describe("RecommendationReasonSelect — B: 選択肢の表示", () => {
  it("トリガークリック → 4選択肢（再診/健診/予防/検査）がすべて表示される", async () => {
    renderSelect(null);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("recommendation-reason-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "再診" })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: "健診" })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: "予防" })).toBeInTheDocument();
      expect(screen.getByRole("option", { name: "検査" })).toBeInTheDocument();
    });
  });

  it("value が設定済みの場合「クリア」選択肢が表示される", async () => {
    renderSelect("checkup");
    const user = userEvent.setup();
    await user.click(screen.getByTestId("recommendation-reason-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "クリア" })).toBeInTheDocument();
    });
  });

  it("value=null の場合「クリア」選択肢は表示されない", async () => {
    renderSelect(null);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("recommendation-reason-trigger"));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "再診" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("option", { name: "クリア" })).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// C: PATCH エンドポイント呼び出し
// ─────────────────────────────────────────────────────────────

describe("RecommendationReasonSelect — C: PATCH エンドポイント呼び出し", () => {
  it("「再診」を選択 → PATCH に reason:'revisit' が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/medical-records/${RECORD_ID}/recommendation-reason`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({
            ...minimalMedicalRecordApiResponse,
            recommendation_reason: "revisit",
          });
        },
      ),
    );
    renderSelect(null);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("recommendation-reason-trigger"));
    await waitFor(() => screen.getByRole("option", { name: "再診" }));
    await user.click(screen.getByRole("option", { name: "再診" }));
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ reason: "revisit" }));
    });
  });

  it("「予防」を選択 → PATCH に reason:'prevention' が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/medical-records/${RECORD_ID}/recommendation-reason`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({
            ...minimalMedicalRecordApiResponse,
            recommendation_reason: "prevention",
          });
        },
      ),
    );
    renderSelect(null);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("recommendation-reason-trigger"));
    await waitFor(() => screen.getByRole("option", { name: "予防" }));
    await user.click(screen.getByRole("option", { name: "予防" }));
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ reason: "prevention" }));
    });
  });

  it("「クリア」選択 → PATCH に reason:null が送られる", async () => {
    let capturedBody: unknown = null;
    server.use(
      http.patch(
        `/api/v1/medical-records/${RECORD_ID}/recommendation-reason`,
        async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({
            ...minimalMedicalRecordApiResponse,
            recommendation_reason: null,
          });
        },
      ),
    );
    renderSelect("checkup");
    const user = userEvent.setup();
    await user.click(screen.getByTestId("recommendation-reason-trigger"));
    await waitFor(() => screen.getByRole("option", { name: "クリア" }));
    await user.click(screen.getByRole("option", { name: "クリア" }));
    await waitFor(() => {
      expect(capturedBody).toEqual(expect.objectContaining({ reason: null }));
    });
  });
});

// ─────────────────────────────────────────────────────────────
// D: disabled 状態
// ─────────────────────────────────────────────────────────────

describe("RecommendationReasonSelect — D: disabled 状態", () => {
  it("disabled=true → Select トリガーが disabled 属性を持つ", () => {
    renderSelect(null, true);
    const trigger = screen.getByTestId("recommendation-reason-trigger");
    expect(trigger).toBeDisabled();
  });
});

// ─────────────────────────────────────────────────────────────
// E: create mode
// ─────────────────────────────────────────────────────────────

describe("RecommendationReasonSelect — E: create mode", () => {
  it("create モードで Select が表示される", () => {
    render(<RecommendationReasonSelect mode="create" value={null} onChange={vi.fn()} />, {
      wrapper: createWrapper(),
    });
    expect(screen.getByTestId("recommendation-reason-trigger")).toBeInTheDocument();
    expect(screen.getByText("未選択")).toBeInTheDocument();
  });

  it("create モードで値選択時 onChange が呼ばれ PATCH は送信されない", async () => {
    let patchCalled = false;
    server.use(
      http.patch(`/api/v1/medical-records/${RECORD_ID}/recommendation-reason`, () => {
        patchCalled = true;
        return HttpResponse.json({});
      }),
    );
    const onChange = vi.fn();
    render(<RecommendationReasonSelect mode="create" value={null} onChange={onChange} />, {
      wrapper: createWrapper(),
    });
    const user = userEvent.setup();
    await user.click(screen.getByTestId("recommendation-reason-trigger"));
    await waitFor(() => screen.getByRole("option", { name: "健診" }));
    await user.click(screen.getByRole("option", { name: "健診" }));
    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith("checkup");
    });
    expect(patchCalled).toBe(false);
  });

  it("create モードで「クリア」選択時 onChange(null) が呼ばれる", async () => {
    const onChange = vi.fn();
    render(<RecommendationReasonSelect mode="create" value={"checkup"} onChange={onChange} />, {
      wrapper: createWrapper(),
    });
    const user = userEvent.setup();
    await user.click(screen.getByTestId("recommendation-reason-trigger"));
    await waitFor(() => screen.getByRole("option", { name: "クリア" }));
    await user.click(screen.getByRole("option", { name: "クリア" }));
    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith(null);
    });
  });
});
