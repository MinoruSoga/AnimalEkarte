import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { TreatmentsTab } from "./TreatmentsTab";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

// PageLayout 非経由でも usePermission → useAuth が呼ばれるため最小モックを用意する
// （CashRegisterHistoryPage.test.tsx と同パターン）。
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }] },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

// TreatmentSearchDialog は検索 UI を持つ重量コンポーネントのため、統合境界として
// onSelect のみを露出するスタブに差し替える（実検索 UX は TreatmentSearchDialog 自身の
// テスト対象であり、ここでの関心事は handleSelectFromMaster 側の hard gate 分岐）。
const UNSAFE_MASTER_ITEM: TreatmentMasterItem = {
  id: "med-1",
  name: "劇薬A",
  unitPrice: 100,
  category: "薬剤",
  medicineId: "1",
};

// 負のコントロール: medicineId を持たない（per-weight 計算対象外の）選択は
// doseCalcInput が常に null → computeDoseGate は常に requiresDeviationReason=false。
// hard gate が「常にダイアログを出す」実装に劣化していないことを確認する。
const SAFE_MASTER_ITEM: TreatmentMasterItem = {
  id: "cons-1",
  name: "一般診察",
  unitPrice: 500,
  category: "診察",
};

vi.mock("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog", () => ({
  TreatmentSearchDialog: ({ onSelect }: { onSelect: (item: TreatmentMasterItem) => void }) => (
    <>
      <button type="button" onClick={() => onSelect(UNSAFE_MASTER_ITEM)}>
        劇薬Aを選択
      </button>
      <button type="button" onClick={() => onSelect(SAFE_MASTER_ITEM)}>
        一般診察を選択
      </button>
    </>
  ),
}));

const MEDICAL_RECORD_ID = "10";

// P1-7 (PR #186 review) 回帰: 体重4kg・strength=50mg/tablet・dose_per_kg=0.5・max_mg_per_kg=1・
// rounding_step=1/up の per-weight 薬剤。推奨プリフィルは丸め up で quantity=1 になった時点で
// 実効用量50mg > 上限4mg となり exceedsMax=true → プリフィルは使われず既定値1にフェイルクローズする。
// しかしフォールバック値1自体も安全域外（同じ50mg > 4mg）——create 前に hard gate が必須。
function installUnsafeMedicineHandlers() {
  server.use(
    http.get("*/v1/medical-records/:id/treatments", () => HttpResponse.json([])),
    http.get("*/v1/medical-records/:id/vitals", () =>
      HttpResponse.json([
        {
          id: "1",
          medical_record_id: MEDICAL_RECORD_ID,
          recorded_at: "2026-07-15T00:00:00Z",
          weight: 4,
          weight_unit: "Kg",
          created_at: "2026-07-15T00:00:00Z",
          updated_at: "2026-07-15T00:00:00Z",
        },
      ])
    ),
    http.get("*/v1/masters/medicines", () =>
      HttpResponse.json([
        {
          id: 1,
          name: "劇薬A",
          medicine_unit: "per_tablet",
          price: 100,
          is_active: true,
          sort_order: 0,
          tax_type: "excluded",
          tax_rate: 0.1,
          is_non_insurance: false,
          calculation_type: "per_weight",
          strength: 50,
        },
      ])
    ),
    http.get("*/v1/masters/medicines/1/dose-params", () =>
      HttpResponse.json([
        {
          id: 1,
          clinic_id: 1,
          medicine_id: 1,
          species: "dog",
          dose_basis: "per_administration",
          dose_per_kg: 0.5,
          max_mg_per_kg: 1,
          rounding_step: 1,
          rounding_mode: "up",
          notes: "",
          created_at: "2026-07-15T00:00:00Z",
          updated_at: "2026-07-15T00:00:00Z",
        },
      ])
    )
  );
}

describe("TreatmentsTab — マスタ選択時の投与量 hard gate (P1-7)", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("フォールバック quantity=1 が上限超過の場合、create せず inline で保存不可理由を表示する", async () => {
    installUnsafeMedicineHandlers();
    let createCalled = false;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", async () => {
        createCalled = true;
        return HttpResponse.json({ id: "new-1" }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    render(<TreatmentsTab medicalRecordId={MEDICAL_RECORD_ID} petSpecies="dog" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: "劇薬Aを選択" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/上限.*保存できません/);
    expect(createCalled).toBe(false);
    expect(screen.queryByText("投与量を確認してください")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /この数量で追加/ })).not.toBeInTheDocument();
  });

  it("medicineId を持たない選択（per-weight 対象外）では gate を出さず即座に create する", async () => {
    installUnsafeMedicineHandlers();
    let createCalled = false;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", async () => {
        createCalled = true;
        return HttpResponse.json({ id: "new-2" }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    render(<TreatmentsTab medicalRecordId={MEDICAL_RECORD_ID} petSpecies="dog" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: "一般診察を選択" }));

    await waitFor(() => expect(createCalled).toBe(true));
    expect(screen.queryByText("投与量を確認してください")).not.toBeInTheDocument();
  });
});

// #201 TASK-025: dose-params technical failure は silent fallback せず保存を止める。
// missing data（体重・species・param 未設定）の挙動変更は TASK-033 まで HOLD。
describe("TreatmentsTab — dose-params technical failure (TASK-025)", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  const UPSTREAM_LEAK = "database timeout; SQLSTATE 57P01 internal stack";

  function installLookupFailureHandlers() {
    installUnsafeMedicineHandlers();
    server.use(
      http.get("*/v1/masters/medicines/1/dose-params", () =>
        HttpResponse.json({ message: UPSTREAM_LEAK }, { status: 500 })
      )
    );
  }

  it("dose-params が 500 のとき create せず technical failure を role=alert で表示する", async () => {
    installLookupFailureHandlers();
    let createCount = 0;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", async () => {
        createCount += 1;
        return HttpResponse.json({ id: "new-fail-1" }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    render(<TreatmentsTab medicalRecordId={MEDICAL_RECORD_ID} petSpecies="dog" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: "劇薬Aを選択" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(/投与量パラメータを取得できなかった/);
    expect(alert).not.toHaveTextContent(UPSTREAM_LEAK);
    expect(screen.queryByText(UPSTREAM_LEAK)).not.toBeInTheDocument();
    expect(createCount).toBe(0);
  });

  it("dose-params technical failure 中は POST /treatments が 0 回のままである", async () => {
    installLookupFailureHandlers();
    let createCount = 0;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", async () => {
        createCount += 1;
        return HttpResponse.json({ id: "new-fail-2" }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    render(<TreatmentsTab medicalRecordId={MEDICAL_RECORD_ID} petSpecies="dog" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: "劇薬Aを選択" }));
    await screen.findByRole("alert");
    expect(createCount).toBe(0);
  });

  it("retry で dose-params が成功したら通常保存経路が再開する", async () => {
    installUnsafeMedicineHandlers();
    // 安全域内の param（strength=10, weight=4, dose_per_kg=2.5, max 十分）で create まで到達させる
    server.use(
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([
          {
            id: 1,
            name: "劇薬A",
            medicine_unit: "per_tablet",
            price: 100,
            is_active: true,
            sort_order: 0,
            tax_type: "excluded",
            tax_rate: 0.1,
            is_non_insurance: false,
            calculation_type: "per_weight",
            strength: 10,
          },
        ])
      )
    );

    let doseParamsCalls = 0;
    server.use(
      http.get("*/v1/masters/medicines/1/dose-params", () => {
        doseParamsCalls += 1;
        if (doseParamsCalls === 1) {
          return HttpResponse.json({ message: UPSTREAM_LEAK }, { status: 500 });
        }
        return HttpResponse.json([
          {
            id: 1,
            clinic_id: 1,
            medicine_id: 1,
            species: "dog",
            dose_basis: "per_administration",
            dose_per_kg: 2.5,
            max_mg_per_kg: 100,
            rounding_step: 1,
            rounding_mode: "up",
            notes: "",
            created_at: "2026-07-15T00:00:00Z",
            updated_at: "2026-07-15T00:00:00Z",
          },
        ]);
      })
    );

    let createCount = 0;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", async () => {
        createCount += 1;
        return HttpResponse.json({ id: "new-retry-1" }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    render(<TreatmentsTab medicalRecordId={MEDICAL_RECORD_ID} petSpecies="dog" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: "劇薬Aを選択" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(/投与量パラメータを取得できなかった/);
    expect(createCount).toBe(0);

    // QueryClient が失敗をキャッシュしている可能性があるため、再試行前に失敗キャッシュを捨てる経路は
    // 実装側（retry ボタン）が担う。ここでは UI の再試行 action を押す。
    await user.click(screen.getByRole("button", { name: /投与量パラメータの取得を再試行/ }));

    await waitFor(() => expect(createCount).toBe(1));
  });

  it("体重未記録でも technical failure alert にせず create を継続する（missing data）", async () => {
    installUnsafeMedicineHandlers();
    server.use(
      http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([]))
    );
    let createCount = 0;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", async () => {
        createCount += 1;
        return HttpResponse.json({ id: "new-miss-weight" }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    render(<TreatmentsTab medicalRecordId={MEDICAL_RECORD_ID} petSpecies="dog" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: "劇薬Aを選択" }));
    await waitFor(() => expect(createCount).toBe(1));
    expect(screen.queryByText(/投与量パラメータを取得できなかった/)).not.toBeInTheDocument();
  });

  it("petSpecies 未設定でも technical failure alert にせず create を継続する（missing data）", async () => {
    installUnsafeMedicineHandlers();
    let createCount = 0;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", async () => {
        createCount += 1;
        return HttpResponse.json({ id: "new-miss-species" }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    render(<TreatmentsTab medicalRecordId={MEDICAL_RECORD_ID} petSpecies={null} />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: "劇薬Aを選択" }));
    await waitFor(() => expect(createCount).toBe(1));
    expect(screen.queryByText(/投与量パラメータを取得できなかった/)).not.toBeInTheDocument();
  });

  it("dose-params 200 で該当 species なし（[]）でも technical failure にせず create を継続する", async () => {
    installUnsafeMedicineHandlers();
    server.use(
      http.get("*/v1/masters/medicines/1/dose-params", () => HttpResponse.json([]))
    );
    let createCount = 0;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", async () => {
        createCount += 1;
        return HttpResponse.json({ id: "new-miss-param" }, { status: 201 });
      })
    );

    const user = userEvent.setup();
    render(<TreatmentsTab medicalRecordId={MEDICAL_RECORD_ID} petSpecies="dog" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(await screen.findByRole("button", { name: "劇薬Aを選択" }));
    await waitFor(() => expect(createCount).toBe(1));
    expect(screen.queryByText(/投与量パラメータを取得できなかった/)).not.toBeInTheDocument();
  });
});
