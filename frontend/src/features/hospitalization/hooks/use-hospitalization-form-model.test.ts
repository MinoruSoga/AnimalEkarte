import { describe, expect, it } from "vitest";

import { calculateNextDate } from "@/components/shared/NextScheduleField";
import type { TreatmentPlanResponse } from "@/types/generated/hospitalization-responses";
import type { BackendHospitalization } from "../api/types";
import {
  buildCreateHospitalizationRequest,
  buildHospitalizationFormDataFromRecord,
  buildPersistableTreatmentPlanRequests,
  buildSelectedPetFromHospitalization,
  buildTreatmentPlansFromRecord,
  buildUpdateHospitalizationRequest,
  createEmptyTreatmentPlan,
  createInitialHospitalizationFormData,
  hospitalizationSubmitFieldErrors,
  mergePetIntoHospitalizationFormData,
  updateTreatmentPlanField,
} from "./use-hospitalization-form-model";

/** HospitalizationResponse wire fixture (no treatment_plans embed). */
const hospitalization = {
  id: 990007,
  clinic_id: 990001,
  owner_id: 990002,
  pet_id: 990003,
  hospitalization_type: "hospitalization",
  start_date: "2026-07-23T00:00:00+09:00",
  end_date: "2026-07-30T00:00:00+09:00",
  status: "admitted",
  memo: "",
  owner_request: "",
  staff_notes: "",
  created_at: "2026-07-23T00:00:00+09:00",
  updated_at: "2026-07-23T00:00:00+09:00",
} satisfies BackendHospitalization;

/** GET /hospitalizations/:id/treatment-plans wire fixture. */
const treatmentPlanWire: TreatmentPlanResponse[] = [
  {
    id: "990018",
    hospitalization_id: "990007",
    treatment_content: "合成監査輸液",
    memo: "既存record由来",
    is_insurance: true,
    unit_price: 3_210,
    quantity: 2,
    discount_rate: 10,
    discount_amount: 642,
    subtotal: 5_778,
    sort_order: 1,
    created_at: "2026-07-23T00:00:00+09:00",
    updated_at: "2026-07-23T00:00:00+09:00",
  },
];

const hospitalizationWithPet = {
  ...hospitalization,
  owner: { id: 990002, name: "合成監査飼主" },
  pet: {
    id: 990003,
    name: "合成監査ペット",
    pet_number: "SYN-PET-990003",
    status: "alive",
    breed: "合成監査品種",
    animal_species: {
      id: 990004,
      name: "合成動物種",
    },
  },
} satisfies BackendHospitalization;

const formData = {
  ...createInitialHospitalizationFormData(),
  hospitalizationType: "ホテル" as const,
  displayDate: "2026-07-23",
  endDate: "2026-07-30",
  ownerRequest: "合成監査要望",
  staffNotes: "合成監査引継ぎ",
  memo: "合成監査メモ",
  cageId: "990019",
  isInsurance: true,
  insuranceCompanyName: "合成監査保険",
  insuranceNumber: "SYN-INS-1",
};

describe("buildTreatmentPlansFromRecord", () => {
  it("GET treatment-plans wire を UI form shape へ欠落なく変換する", () => {
    expect(buildTreatmentPlansFromRecord(treatmentPlanWire)).toEqual([
      {
        id: "990018",
        treatmentContent: "合成監査輸液",
        memo: "既存record由来",
        is_insurance: true,
        unitPrice: 3_210,
        quantity: 2,
        discount: 10,
        discountAmount: 642,
        subtotal: 5_778,
      },
    ]);
  });

  it("明細のない wire 配列に新規作成用 default を混入させない", () => {
    expect(buildTreatmentPlansFromRecord([])).toEqual([]);
  });

  it("id は wire 上 string のまま UI に載せる（models.TreatmentPlan の number ではない）", () => {
    const [plan] = buildTreatmentPlansFromRecord(treatmentPlanWire);
    expect(typeof plan.id).toBe("string");
    expect(plan.id).toBe("990018");
  });

  it("persistable create bodies skip empty content and map UI fields to wire", () => {
    const empty = createEmptyTreatmentPlan();
    const filled = {
      ...createEmptyTreatmentPlan(),
      treatmentContent: "  adm rate  ",
      memo: "memo",
      is_insurance: true,
      unitPrice: 990,
      quantity: 2,
      discount: 10,
      discountAmount: 198,
    };
    const bodies = buildPersistableTreatmentPlanRequests([empty, filled]);
    expect(bodies).toHaveLength(1);
    expect(bodies[0]).toEqual({
      treatment_content: "adm rate",
      memo: "memo",
      is_insurance: true,
      unit_price: 990,
      quantity: 2,
      discount_rate: 10,
      discount_amount: 198,
      sort_order: 0,
    });
  });
});

describe("hospitalization form model", () => {
  it("新規formの安全な初期値を生成する", () => {
    const initial = createInitialHospitalizationFormData();

    expect(initial.hospitalizationType).toBe("入院");
    expect(initial.endDate).not.toBe("");
    expect(initial.isInsurance).toBe(false);
    expect(initial.doctorId).toBe("");
    expect(initial.ownerRequest).toBe("");
  });

  it("編集payloadへ保険・ケージ・臨床メモを欠落なく写す", () => {
    expect(buildUpdateHospitalizationRequest(formData, "")).toEqual({
      hospitalization_type: "hotel",
      start_date: "2026-07-23T00:00:00+09:00",
      end_date: "2026-07-30T00:00:00+09:00",
      owner_request: "合成監査要望",
      staff_notes: "合成監査引継ぎ",
      memo: "合成監査メモ",
      cage_id: "990019",
      expected_cage_is_null: true,
      is_insurance: true,
      insurance_company_name: "合成監査保険",
      insurance_number: "SYN-INS-1",
    });
  });

  it("担当医は doctor_id として create/update に載せる（BUG-001）", () => {
    expect(buildUpdateHospitalizationRequest(formData, "")).not.toHaveProperty("doctor_id");

    const withDoctor = { ...formData, doctorId: "12", doctorName: "山田医師" };
    expect(buildUpdateHospitalizationRequest(withDoctor, "").doctor_id).toBe("12");

    const pet = buildSelectedPetFromHospitalization(hospitalizationWithPet);
    expect(pet).not.toBeNull();
    if (!pet) return;
    expect(buildCreateHospitalizationRequest(withDoctor, pet).doctor_id).toBe("12");
  });

  it("既存recordの doctor nest を担当医表示へ復元する", () => {
    expect(
      buildHospitalizationFormDataFromRecord(formData, {
        ...hospitalization,
        doctor_id: 12,
        doctor: { id: 12, name: "山田医師" },
      }),
    ).toMatchObject({
      doctorId: "12",
      doctorName: "山田医師",
    });
  });

  it("保険未使用なら会社名・番号をnullへ閉じる", () => {
    const request = buildUpdateHospitalizationRequest(
      { ...formData, cageId: "", isInsurance: false },
      "",
    );

    expect(request.cage_id).toBeUndefined();
    expect(request.insurance_company_name).toBeNull();
    expect(request.insurance_number).toBeNull();
  });

  // EMR-316: 編集で「選択なし」に戻すと cage_id を NULL にする指示（clear_cage_id）を送る。
  it("編集payload: 割当済みケージを解除した場合のみ clear_cage_id を送る", () => {
    const cleared = buildUpdateHospitalizationRequest({ ...formData, cageId: "" }, "990019");
    expect(cleared.clear_cage_id).toBe(true);
    expect(cleared.expected_cage_id).toBe("990019");
    expect(cleared.cage_id).toBeUndefined();
    expect(cleared).not.toHaveProperty("expected_cage_is_null");

    const selected = buildUpdateHospitalizationRequest(formData, "");
    expect(selected.cage_id).toBe("990019");
    expect(selected.clear_cage_id).toBeUndefined();
    expect(selected.expected_cage_is_null).toBe(true);
  });

  it("編集payload: ケージ変更は表示時点の expected_cage_id を同送し、初期未割当なら expected_cage_is_null を送る", () => {
    const moved = buildUpdateHospitalizationRequest({ ...formData, cageId: "990020" }, "990019");
    expect(moved.cage_id).toBe("990020");
    expect(moved.expected_cage_id).toBe("990019");
    expect(moved).not.toHaveProperty("expected_cage_is_null");

    const nullToCleared = buildUpdateHospitalizationRequest({ ...formData, cageId: "" }, "");
    expect(nullToCleared).not.toHaveProperty("clear_cage_id");
    expect(nullToCleared).not.toHaveProperty("expected_cage_id");
    expect(nullToCleared).not.toHaveProperty("expected_cage_is_null");
  });

  it.each(["", "990019"])("未変更のケージ %s は更新payloadから除外する", (cageId) => {
    const request = buildUpdateHospitalizationRequest({ ...formData, cageId }, cageId);
    expect(request).not.toHaveProperty("cage_id");
    expect(request).not.toHaveProperty("clear_cage_id");
    expect(request).not.toHaveProperty("expected_cage_id");
    expect(request).not.toHaveProperty("expected_cage_is_null");
  });

  it("新規payload: ケージ未選択は cage_id も clear_cage_id も送らない（ケージ未割当で登録）", () => {
    const pet = buildSelectedPetFromHospitalization(hospitalizationWithPet);
    expect(pet).not.toBeNull();
    if (!pet) return;

    const request = buildCreateHospitalizationRequest({ ...formData, cageId: "" }, pet);
    expect(request).not.toHaveProperty("cage_id");
    expect(request).not.toHaveProperty("clear_cage_id");
  });

  it("送信前検証: ケージ未選択でも通る。ペット未選択・死亡のみ拒否する", () => {
    const alivePet = { id: "1", status: "生存" } as Parameters<
      typeof hospitalizationSubmitFieldErrors
    >[0];
    expect(hospitalizationSubmitFieldErrors(alivePet, false)).toBeNull();
    expect(hospitalizationSubmitFieldErrors(undefined, false)).toEqual({
      pet: "ペットを選択してください",
    });
    const deceasedPet = { id: "1", status: "死亡" } as Parameters<
      typeof hospitalizationSubmitFieldErrors
    >[0];
    expect(hospitalizationSubmitFieldErrors(deceasedPet, false)?.pet).toBe(
      "死亡したペットは入院登録できません",
    );
  });

  it("新規payloadへpet/ownerとJST日付を写す", () => {
    const pet = buildSelectedPetFromHospitalization(hospitalizationWithPet);
    expect(pet).not.toBeNull();
    if (!pet) return;

    const request = buildCreateHospitalizationRequest(formData, pet);
    expect(request).toMatchObject({
      pet_id: "990003",
      owner_id: "990002",
      hospitalization_type: "hotel",
      cage_id: "990019",
      is_insurance: true,
    });
    expect(request.start_date).toContain("2026-07-23");
    expect(request.end_date).toContain("2026-07-30");
    expect(request.treatment_plans).toBeUndefined();
  });

  it("治療内容ありの行は create payload の treatment_plans に同梱し空行は省略する", () => {
    const pet = buildSelectedPetFromHospitalization(hospitalizationWithPet);
    expect(pet).not.toBeNull();
    if (!pet) return;

    const request = buildCreateHospitalizationRequest(formData, pet, [
      {
        id: "a",
        treatmentContent: "adm rate",
        memo: "",
        is_insurance: false,
        unitPrice: 990,
        quantity: 1,
        discount: 0,
        discountAmount: 0,
        subtotal: 990,
      },
      {
        id: "b",
        treatmentContent: "   ",
        memo: "",
        is_insurance: false,
        unitPrice: 0,
        quantity: 1,
        discount: 0,
        discountAmount: 0,
        subtotal: 0,
      },
      {
        id: "c",
        treatmentContent: "monitor",
        memo: "note",
        is_insurance: false,
        unitPrice: 500,
        quantity: 2,
        discount: 10,
        discountAmount: 0,
        subtotal: 900,
      },
    ]);

    expect(request.treatment_plans).toEqual([
      expect.objectContaining({
        treatment_content: "adm rate",
        unit_price: 990,
        quantity: 1,
        sort_order: 0,
      }),
      expect.objectContaining({
        treatment_content: "monitor",
        unit_price: 500,
        quantity: 2,
        discount_rate: 10,
        sort_order: 1,
      }),
    ]);
  });

  it("backend recordから編集formの臨床状態を復元する", () => {
    expect(
      buildHospitalizationFormDataFromRecord(formData, {
        ...hospitalization,
        cage_id: 990019,
        insurance_company_name: "合成監査保険",
        insurance_number: "SYN-INS-1",
      }),
    ).toMatchObject({
      hospitalizationType: "入院",
      cageId: "990019",
      // EMR-353: timestamp 形式の wire 値は共有日付計算が扱えるよう date-only に正規化する。
      displayDate: "2026-07-23",
      endDate: "2026-07-30",
      memo: hospitalization.memo,
      ownerRequest: hospitalization.owner_request,
      staffNotes: hospitalization.staff_notes,
      isInsurance: true,
      insuranceCompanyName: "合成監査保険",
      insuranceNumber: "SYN-INS-1",
    });
  });

  // EMR-353: 次回予定日の wire 往復。create は入力日を送り、PATCH は保存値からの
  // 差分だけを送る（未変更は省略、消去は clear_next_visit）。
  it("新規payload: nextVisit があるとき next_visit を送り、空なら送らない", () => {
    const pet = buildSelectedPetFromHospitalization(hospitalizationWithPet);
    expect(pet).not.toBeNull();
    if (!pet) return;

    const withDate = buildCreateHospitalizationRequest(
      { ...formData, nextVisit: "2026-11-20" },
      pet,
    );
    expect(withDate.next_visit).toBe("2026-11-20");

    const withoutDate = buildCreateHospitalizationRequest({ ...formData, nextVisit: "" }, pet);
    expect(withoutDate).not.toHaveProperty("next_visit");
  });

  it("編集payload: nextVisit が保存値と同じなら next_visit も clear_next_visit も送らない", () => {
    const request = buildUpdateHospitalizationRequest(
      { ...formData, nextVisit: "2026-11-20" },
      "",
      "2026-11-20",
    );
    expect(request).not.toHaveProperty("next_visit");
    expect(request).not.toHaveProperty("clear_next_visit");
  });

  it("編集payload: 保存値から日付を変更したとき next_visit を送る", () => {
    const request = buildUpdateHospitalizationRequest(
      { ...formData, nextVisit: "2026-12-01" },
      "",
      "2026-11-20",
    );
    expect(request.next_visit).toBe("2026-12-01");
    expect(request).not.toHaveProperty("clear_next_visit");
  });

  it("編集payload: 保存済みの日付を消去したとき clear_next_visit: true を送る", () => {
    const request = buildUpdateHospitalizationRequest(
      { ...formData, nextVisit: "" },
      "",
      "2026-11-20",
    );
    expect(request.clear_next_visit).toBe(true);
    expect(request).not.toHaveProperty("next_visit");
  });

  it("編集payload: 未設定のままなら next_visit も clear_next_visit も送らない", () => {
    const request = buildUpdateHospitalizationRequest({ ...formData, nextVisit: "" }, "", "");
    expect(request).not.toHaveProperty("next_visit");
    expect(request).not.toHaveProperty("clear_next_visit");
  });

  it("編集payload: 変更した期間を JST 0時の start_date/end_date として送り、プリセット由来の next_visit を同送する", () => {
    const hydrated = buildHospitalizationFormDataFromRecord(
      createInitialHospitalizationFormData(),
      hospitalization,
    );
    const edited = {
      ...hydrated,
      displayDate: "2026-08-01",
      endDate: "2026-08-08",
      nextScheduleType: "4weeks",
      nextVisit: calculateNextDate("2026-08-01", "4weeks"),
    };

    const request = buildUpdateHospitalizationRequest(edited, hydrated.cageId, hydrated.nextVisit, {
      startDate: hydrated.displayDate,
      endDate: hydrated.endDate,
    });

    expect(request.start_date).toBe("2026-08-01T00:00:00+09:00");
    expect(request.end_date).toBe("2026-08-08T00:00:00+09:00");
    expect(request.next_visit).toBe("2026-08-29");
    expect(request).not.toHaveProperty("clear_next_visit");
  });

  it("編集payload: 保存値から変わっていない期間は start_date/end_date を送らない", () => {
    const hydrated = buildHospitalizationFormDataFromRecord(
      createInitialHospitalizationFormData(),
      hospitalization,
    );

    const request = buildUpdateHospitalizationRequest(
      hydrated,
      hydrated.cageId,
      hydrated.nextVisit,
      { startDate: hydrated.displayDate, endDate: hydrated.endDate },
    );

    expect(request).not.toHaveProperty("start_date");
    expect(request).not.toHaveProperty("end_date");
  });

  it("編集payload: 片方だけ変えた期間は変えた側の日付だけ送る", () => {
    const initialDates = { startDate: "2026-07-23", endDate: "2026-07-30" };

    const startOnly = buildUpdateHospitalizationRequest(
      { ...formData, displayDate: "2026-07-24" },
      formData.cageId,
      "",
      initialDates,
    );
    expect(startOnly.start_date).toBe("2026-07-24T00:00:00+09:00");
    expect(startOnly).not.toHaveProperty("end_date");

    const endOnly = buildUpdateHospitalizationRequest(
      { ...formData, endDate: "2026-07-31" },
      formData.cageId,
      "",
      initialDates,
    );
    expect(endOnly).not.toHaveProperty("start_date");
    expect(endOnly.end_date).toBe("2026-07-31T00:00:00+09:00");
  });

  it("編集payload: 期間が空なら start_date/end_date を送らない", () => {
    const request = buildUpdateHospitalizationRequest(
      { ...formData, displayDate: "", endDate: "" },
      formData.cageId,
      "",
      { startDate: "2026-07-23", endDate: "2026-07-30" },
    );
    expect(request).not.toHaveProperty("start_date");
    expect(request).not.toHaveProperty("end_date");
  });

  it("編集payload: displayDate が空なら start_date を送らない", () => {
    const request = buildUpdateHospitalizationRequest({ ...formData, displayDate: "" }, "");
    expect(request).not.toHaveProperty("start_date");
  });

  it("編集payload: 開始日の変更と手動で変更した nextVisit を両方送る", () => {
    const request = buildUpdateHospitalizationRequest(
      {
        ...formData,
        displayDate: "2026-08-01",
        endDate: "2026-08-08",
        nextVisit: "2026-08-15",
        nextScheduleType: "other",
      },
      "",
      "2026-11-20",
      { startDate: "2026-07-23", endDate: "2026-07-30" },
    );
    expect(request.start_date).toBe("2026-08-01T00:00:00+09:00");
    expect(request.end_date).toBe("2026-08-08T00:00:00+09:00");
    expect(request.next_visit).toBe("2026-08-15");
    expect(request).not.toHaveProperty("clear_next_visit");
  });

  it("編集payload: nextVisit が保存値と同じでも変更した start_date は送る", () => {
    const request = buildUpdateHospitalizationRequest(
      {
        ...formData,
        displayDate: "2026-08-01",
        endDate: "2026-08-08",
        nextVisit: "2026-11-20",
      },
      "",
      "2026-11-20",
      { startDate: "2026-07-23", endDate: "2026-07-30" },
    );
    expect(request.start_date).toBe("2026-08-01T00:00:00+09:00");
    expect(request.end_date).toBe("2026-08-08T00:00:00+09:00");
    expect(request).not.toHaveProperty("next_visit");
    expect(request).not.toHaveProperty("clear_next_visit");
  });

  it("編集recordの next_visit を form の nextVisit へ date-only 正規化し nextScheduleType は other にする", () => {
    expect(
      buildHospitalizationFormDataFromRecord(formData, {
        ...hospitalization,
        next_visit: "2026-08-15",
      }),
    ).toMatchObject({
      nextVisit: "2026-08-15",
      nextScheduleType: "other",
      displayDate: "2026-07-23",
    });
  });

  it("next_visit が無い record は nextVisit 空・nextScheduleType other で復元する", () => {
    expect(buildHospitalizationFormDataFromRecord(formData, hospitalization)).toMatchObject({
      nextVisit: "",
      nextScheduleType: "other",
    });
  });

  it("pet relationがないrecordは選択petを作らない", () => {
    expect(buildSelectedPetFromHospitalization(hospitalization)).toBeNull();
  });

  it("typed pet summary nest を form 選択値へ変換する", () => {
    const pet = buildSelectedPetFromHospitalization(hospitalizationWithPet);

    expect(pet).toMatchObject({
      id: "990003",
      ownerId: "990002",
      name: "合成監査ペット",
      species: "合成動物種",
      breed: "合成監査品種",
      status: "生存",
    });
  });

  it("編集recordの死亡statusを選択petへ明示的に保持する", () => {
    const pet = buildSelectedPetFromHospitalization({
      ...hospitalizationWithPet,
      pet: {
        ...hospitalizationWithPet.pet!,
        status: "deceased",
      },
    });

    expect(pet).toMatchObject({
      id: "990003",
      status: "死亡",
    });
  });

  it("未選択時はformを保持し、選択後はpet表示値だけimmutableに合成する", () => {
    expect(mergePetIntoHospitalizationFormData(formData, [])).toBe(formData);
    const pet = buildSelectedPetFromHospitalization(hospitalizationWithPet);
    expect(pet).not.toBeNull();
    if (!pet) return;

    const merged = mergePetIntoHospitalizationFormData(formData, [
      { ...pet, ownerName: "合成監査飼主", weight: "4.2" },
    ]);
    expect(merged).not.toBe(formData);
    expect(merged).toMatchObject({
      ownerName: "合成監査飼主",
      petName: "合成監査ペット",
      petNumber: "990003",
      species: "合成動物種",
      weight: "4.2kg",
    });
  });

  it("空の治療明細を一意IDと安全な数値初期値で作る", () => {
    expect(createEmptyTreatmentPlan()).toMatchObject({
      treatmentContent: "",
      is_insurance: false,
      unitPrice: 0,
      quantity: 1,
      discount: 0,
      discountAmount: 0,
      subtotal: 0,
    });
  });

  it("単価・数量・割引更新時だけ小計を再計算し元objectを変更しない", () => {
    const original = buildTreatmentPlansFromRecord(treatmentPlanWire)[0];
    const updated = updateTreatmentPlanField(original, "quantity", 3);
    const memoOnly = updateTreatmentPlanField(original, "memo", "更新メモ");

    expect(updated).not.toBe(original);
    expect(updated).toMatchObject({ quantity: 3, discountAmount: 963, subtotal: 8_667 });
    expect(original.quantity).toBe(2);
    expect(memoOnly).toMatchObject({ memo: "更新メモ", subtotal: 5_778 });
  });
});
