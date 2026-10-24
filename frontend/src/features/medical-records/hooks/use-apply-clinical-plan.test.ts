import { act, renderHook, waitFor } from "@testing-library/react";
import { useState } from "react";
import { describe, expect, it } from "vitest";

import type { ClinicalPlan, ClinicalPlanDiagnosisRow } from "../api/clinical-plan";
import { useApplyClinicalPlan } from "./use-apply-clinical-plan";

function diagnosisRow(
  slot: number,
  typeId: string | null,
  nameId: string | null,
  overrides: Partial<ClinicalPlanDiagnosisRow> = {},
): ClinicalPlanDiagnosisRow {
  return {
    id: `d-${slot}-${nameId ?? "none"}`,
    version: 1,
    diagnosis_slot: slot,
    diagnosis_type_id: typeId,
    diagnosis_name_id: nameId,
    diagnosis_name_snapshot: null,
    category_name_snapshot: null,
    sort_order: 0,
    ...overrides,
  };
}

function clinicalPlan(overrides: Partial<ClinicalPlan> = {}): ClinicalPlan {
  return {
    id: "cp-10",
    medical_record_id: "10",
    physical_exam: "v1 身体所見",
    // 診断1 = 区分3/名称7、診断2 = 区分4/名称9（BE は slot, sort_order, id 昇順で返す）
    diagnoses: [diagnosisRow(1, "3", "7"), diagnosisRow(2, "4", "9")],
    diagnosis_details: "v1 診断詳細",
    treatment_policy: "v1 治療方針",
    created_at: "2026-09-05T00:00:00Z",
    updated_at: "2026-09-05T00:00:00Z",
    version: 1,
    ...overrides,
  };
}

function draftFrom(plan: ClinicalPlan) {
  const slot1Row = plan.diagnoses.find((row) => row.diagnosis_slot === 1);
  const slot2Row = plan.diagnoses.find((row) => row.diagnosis_slot === 2);
  const toNumber = (id: string | null | undefined) => (id == null ? null : Number(id));
  return {
    physicalExam: plan.physical_exam,
    plan: plan.treatment_policy,
    assessment: plan.diagnosis_details,
    diagnosis1CategoryId: toNumber(slot1Row?.diagnosis_type_id),
    diagnosis1NameId: toNumber(slot1Row?.diagnosis_name_id),
    diagnosis2CategoryId: toNumber(slot2Row?.diagnosis_type_id),
    diagnosis2NameId: toNumber(slot2Row?.diagnosis_name_id),
  };
}

function useClinicalPlanHarness(recordId: string, plan?: ClinicalPlan) {
  const [physicalExam, setPhysicalExam] = useState("");
  const [planValue, setPlan] = useState("");
  const [assessment, setAssessment] = useState("");
  const [diagnosis1CategoryId, setDiagnosis1CategoryId] = useState<number | null>(null);
  const [diagnosis1NameId, setDiagnosis1NameId] = useState<number | null>(null);
  const [diagnosis2CategoryId, setDiagnosis2CategoryId] = useState<number | null>(null);
  const [diagnosis2NameId, setDiagnosis2NameId] = useState<number | null>(null);
  const snapshot = useApplyClinicalPlan({
    recordId,
    clinicalPlan: plan,
    physicalExam,
    plan: planValue,
    assessment,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
    setPhysicalExam,
    setPlan,
    setAssessment,
    setDiagnosis1CategoryId,
    setDiagnosis1NameId,
    setDiagnosis2CategoryId,
    setDiagnosis2NameId,
  });

  return {
    ...snapshot,
    physicalExam,
    plan: planValue,
    assessment,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
    setPlan,
    setDiagnosis1CategoryId,
  };
}

describe("useApplyClinicalPlan", () => {
  it("dirty な v1 編集中の remote v2 を採用せず、baseline version も v1 に保つ", async () => {
    const v1 = clinicalPlan();
    const v2 = clinicalPlan({
      version: 2,
      physical_exam: "B の身体所見",
      treatment_policy: "B の治療方針",
      diagnosis_details: "B の診断詳細",
      diagnoses: [diagnosisRow(1, "11", "12"), diagnosisRow(2, "13", "14")],
    });
    const { result, rerender } = renderHook(({ plan }) => useClinicalPlanHarness("10", plan), {
      initialProps: { plan: v1 },
    });

    await waitFor(() => {
      expect(result.current.plan).toBe("v1 治療方針");
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
    act(() => {
      result.current.setPlan("A の編集値");
    });
    rerender({ plan: v2 });

    await waitFor(() => {
      expect(result.current.plan).toBe("A の編集値");
      expect(result.current.diagnosis1CategoryId).toBe(3);
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
  });

  it("clean な remote v2 は全 field と baseline version を一緒に更新する", async () => {
    const v1 = clinicalPlan();
    const v2 = clinicalPlan({
      version: 2,
      physical_exam: "v2 身体所見",
      treatment_policy: "v2 治療方針",
      diagnosis_details: "v2 診断詳細",
      diagnoses: [diagnosisRow(1, "11", "12"), diagnosisRow(2, "13", "14")],
    });
    const { result, rerender } = renderHook(({ plan }) => useClinicalPlanHarness("10", plan), {
      initialProps: { plan: v1 },
    });

    await waitFor(() => {
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
    rerender({ plan: v2 });

    await waitFor(() => {
      expect(result.current.physicalExam).toBe("v2 身体所見");
      expect(result.current.plan).toBe("v2 治療方針");
      expect(result.current.assessment).toBe("v2 診断詳細");
      expect(result.current.diagnosis1CategoryId).toBe(11);
      expect(result.current.diagnosis2NameId).toBe(14);
      expect(result.current.clinicalPlanVersion).toBe(2);
    });
  });

  it("own PATCH response は新しい clean snapshot として採用する", async () => {
    const v1 = clinicalPlan();
    const saved = clinicalPlan({ version: 2, treatment_policy: "保存済みの治療方針" });
    const { result } = renderHook(() => useClinicalPlanHarness("10", v1));

    await waitFor(() => {
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
    act(() => {
      result.current.onClinicalPlanSaved(saved, draftFrom(v1));
    });

    await waitFor(() => {
      expect(result.current.plan).toBe("保存済みの治療方針");
      expect(result.current.clinicalPlanVersion).toBe(2);
    });
  });

  it("own PATCH response の後に遅延した古い refetch は baseline を巻き戻さない", async () => {
    const v1 = clinicalPlan();
    const saved = clinicalPlan({ version: 2, treatment_policy: "保存済みの治療方針" });
    const { result, rerender } = renderHook(({ plan }) => useClinicalPlanHarness("10", plan), {
      initialProps: { plan: v1 },
    });

    await waitFor(() => {
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
    act(() => {
      result.current.onClinicalPlanSaved(saved, draftFrom(v1));
    });
    rerender({ plan: v1 });

    await waitFor(() => {
      expect(result.current.plan).toBe("保存済みの治療方針");
      expect(result.current.clinicalPlanVersion).toBe(2);
    });
  });

  it("PATCH 待機中に追加した入力は成功応答で上書きせず、新しい baseline version で保持する", async () => {
    const v1 = clinicalPlan();
    const submitted = clinicalPlan({ treatment_policy: "送信した治療方針" });
    const saved = clinicalPlan({ version: 2, treatment_policy: "送信した治療方針" });
    const { result } = renderHook(() => useClinicalPlanHarness("10", v1));

    await waitFor(() => {
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
    act(() => {
      result.current.setPlan("PATCH 待機中の追加入力");
    });
    await waitFor(() => {
      expect(result.current.plan).toBe("PATCH 待機中の追加入力");
    });
    act(() => {
      result.current.onClinicalPlanSaved(saved, draftFrom(submitted));
    });

    await waitFor(() => {
      expect(result.current.plan).toBe("PATCH 待機中の追加入力");
      expect(result.current.clinicalPlanVersion).toBe(2);
    });
  });

  it("同一 version の遅延診断 hydrate は未編集の診断 snapshot にのみ取り込む", async () => {
    const withoutDiagnosis = clinicalPlan({
      diagnoses: [],
    });
    const withDiagnosis = clinicalPlan();
    const { result, rerender } = renderHook(({ plan }) => useClinicalPlanHarness("10", plan), {
      initialProps: { plan: withoutDiagnosis },
    });

    await waitFor(() => {
      expect(result.current.diagnosis1CategoryId).toBeNull();
    });
    rerender({ plan: withDiagnosis });

    await waitFor(() => {
      expect(result.current.diagnosis1CategoryId).toBe(3);
      expect(result.current.diagnosis2NameId).toBe(9);
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
  });

  it("ユーザーが先に選んだ診断は、同一 version の遅延 hydrate で上書きしない", async () => {
    const withoutDiagnosis = clinicalPlan({
      diagnoses: [],
    });
    const withDiagnosis = clinicalPlan();
    const { result, rerender } = renderHook(({ plan }) => useClinicalPlanHarness("10", plan), {
      initialProps: { plan: withoutDiagnosis },
    });

    await waitFor(() => {
      expect(result.current.diagnosis1CategoryId).toBeNull();
    });
    act(() => {
      result.current.setDiagnosis1CategoryId(99);
    });
    rerender({ plan: withDiagnosis });

    await waitFor(() => {
      expect(result.current.diagnosis1CategoryId).toBe(99);
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
  });

  it("record switch は旧 snapshot を破棄し、次 record の値と version を採用する", async () => {
    const record10 = clinicalPlan();
    const record11 = clinicalPlan({
      id: "cp-11",
      medical_record_id: "11",
      version: 4,
      treatment_policy: "record 11 治療方針",
    });
    const { result, rerender } = renderHook(
      ({ recordId, plan }) => useClinicalPlanHarness(recordId, plan),
      { initialProps: { recordId: "10", plan: record10 } },
    );

    await waitFor(() => {
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
    rerender({ recordId: "11", plan: undefined });

    await waitFor(() => {
      expect(result.current.plan).toBe("");
      expect(result.current.clinicalPlanVersion).toBeUndefined();
    });
    rerender({ recordId: "11", plan: record11 });

    await waitFor(() => {
      expect(result.current.plan).toBe("record 11 治療方針");
      expect(result.current.clinicalPlanVersion).toBe(4);
    });
  });

  it("record switch 後に完了した旧 record の PATCH 応答を破棄する", async () => {
    const record10 = clinicalPlan();
    const savedRecord10 = clinicalPlan({ version: 2, treatment_policy: "旧 record の保存結果" });
    const { result, rerender } = renderHook(
      ({ recordId, plan }) => useClinicalPlanHarness(recordId, plan),
      { initialProps: { recordId: "10", plan: record10 } },
    );

    await waitFor(() => {
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
    const onRecord10Saved = result.current.onClinicalPlanSaved;
    rerender({ recordId: "11", plan: undefined });
    await waitFor(() => {
      expect(result.current.clinicalPlanVersion).toBeUndefined();
    });
    act(() => {
      onRecord10Saved(savedRecord10, draftFrom(record10));
    });

    await waitFor(() => {
      expect(result.current.plan).toBe("");
      expect(result.current.clinicalPlanVersion).toBeUndefined();
    });
  });

  it("診断は slot ごとに返却順の先頭行を代表として hydrate し、baseline には全行を保持する", async () => {
    const rows = [
      diagnosisRow(1, "3", "7", { id: "d-1", sort_order: 0 }),
      diagnosisRow(1, "3", "8", { id: "d-2", sort_order: 1 }),
      diagnosisRow(2, "4", "9", { id: "d-3" }),
    ];
    const v1 = clinicalPlan({ diagnoses: rows });
    const { result } = renderHook(() => useClinicalPlanHarness("10", v1));

    await waitFor(() => {
      expect(result.current.diagnosis1NameId).toBe(7);
      expect(result.current.diagnosis2NameId).toBe(9);
      expect(result.current.clinicalPlanDiagnoses).toEqual(rows);
    });
  });

  it("own PATCH response の診断行を次の baseline 診断行として採用する", async () => {
    const v1 = clinicalPlan();
    const savedRows = [diagnosisRow(1, "3", "7"), diagnosisRow(2, "4", "10")];
    const saved = clinicalPlan({ version: 2, diagnoses: savedRows });
    const { result } = renderHook(() => useClinicalPlanHarness("10", v1));

    await waitFor(() => {
      expect(result.current.clinicalPlanVersion).toBe(1);
    });
    act(() => {
      result.current.onClinicalPlanSaved(saved, draftFrom(v1));
    });

    await waitFor(() => {
      expect(result.current.diagnosis2NameId).toBe(10);
      expect(result.current.clinicalPlanDiagnoses).toEqual(savedRows);
      expect(result.current.clinicalPlanVersion).toBe(2);
    });
  });
});
