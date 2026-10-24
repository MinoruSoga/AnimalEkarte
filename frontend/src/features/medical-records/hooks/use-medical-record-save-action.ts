import { useActionState, useLayoutEffect, useRef } from "react";
import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type {
  ClinicalPlan,
  ClinicalPlanDiagnosisInput,
  ClinicalPlanDiagnosisRow,
} from "../api/clinical-plan";
import type { ClinicalPlanDraft } from "./use-apply-clinical-plan";
import { findCachedDiagnosisLabels, type DiagnosisOptionLabels } from "../lib/diagnosis-option-labels";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";
const DECEASED_SAVE_MESSAGE = "死亡したペットのカルテは保存できません";

function deniedState(error: string): ActionState {
  return { success: false, error, timestamp: Date.now() };
}

function isNullableString(value: unknown): boolean {
  return value === null || typeof value === "string";
}

// BE の診断行（clinical_plan_response.go）。ID 系は文字列 / null、slot と sort_order は数値。
function isClinicalPlanDiagnosisRow(value: unknown): boolean {
  return (
    typeof value === "object" &&
    value !== null &&
    "id" in value &&
    typeof value.id === "string" &&
    "version" in value &&
    typeof value.version === "number" &&
    "diagnosis_slot" in value &&
    typeof value.diagnosis_slot === "number" &&
    "diagnosis_type_id" in value &&
    isNullableString(value.diagnosis_type_id) &&
    "diagnosis_name_id" in value &&
    isNullableString(value.diagnosis_name_id) &&
    "diagnosis_name_snapshot" in value &&
    isNullableString(value.diagnosis_name_snapshot) &&
    "category_name_snapshot" in value &&
    isNullableString(value.category_name_snapshot) &&
    "sort_order" in value &&
    typeof value.sort_order === "number"
  );
}

function isClinicalPlan(value: unknown): value is ClinicalPlan {
  return (
    typeof value === "object" &&
    value !== null &&
    "id" in value &&
    typeof value.id === "string" &&
    "medical_record_id" in value &&
    typeof value.medical_record_id === "string" &&
    "version" in value &&
    typeof value.version === "number" &&
    "physical_exam" in value &&
    typeof value.physical_exam === "string" &&
    "treatment_policy" in value &&
    typeof value.treatment_policy === "string" &&
    "diagnosis_details" in value &&
    typeof value.diagnosis_details === "string" &&
    "created_at" in value &&
    typeof value.created_at === "string" &&
    "updated_at" in value &&
    typeof value.updated_at === "string" &&
    "diagnoses" in value &&
    Array.isArray(value.diagnoses) &&
    value.diagnoses.every(isClinicalPlanDiagnosisRow)
  );
}

/** 診断 slot（BE の CHECK(diagnosis_slot IN (1,2)) と同じ値域）。 */
type ClinicalPlanDiagnosisSlot = 1 | 2;

interface ClinicalPlanDiagnosisSource {
  diagnosis1CategoryId: number | null;
  diagnosis1NameId: number | null;
  diagnosis2CategoryId: number | null;
  diagnosis2NameId: number | null;
  /** 読込済み baseline の診断行（BE が slot, sort_order, id 昇順で返した順のまま）。 */
  existingClinicalPlanDiagnoses?: ClinicalPlanDiagnosisRow[];
}

interface SlotDiagnoses {
  /** UI の選択が読込済みの代表行と異なるか。true の slot だけが置換・クリア対象。 */
  changed: boolean;
  rows: ClinicalPlanDiagnosisInput[];
}

function toNumericId(value: string | null): number | null {
  if (value === null) return null;
  const id = Number(value);
  if (!Number.isSafeInteger(id)) {
    throw new Error(`診断行の ID を数値へ変換できません: ${value}`);
  }
  return id;
}

/** 選択中の区分・名称 ID について、UI が表示しているマスタ名を返す関数（保存時のスナップショット用）。 */
type DiagnosisLabelResolver = (
  categoryId: number | null,
  nameId: number | null,
) => DiagnosisOptionLabels;

/**
 * 1 つの slot の送信行を決める（EMR-495）。
 * - 代表行 = その slot の先頭行。BE は slot, sort_order, id 昇順で返す（ListDiagnoses）。
 * - UI の区分・名称が代表行と同じ: 読込済みの全行を値のまま返す（複数行・スナップショット・sort_order を保持）。
 * - 異なり、両方未選択: その slot の行を持たない（クリア）。
 * - 異なり、選択あり: 新しい 1 行へ置換する。スナップショットは BE が現在のマスタから補完しないため、
 *   UI が表示している区分名・病名を送る（react-query キャッシュ、無ければ未変更の代表行のスナップショット、判明しなければ null）。
 */
function buildSlotDiagnoses(
  slot: ClinicalPlanDiagnosisSlot,
  categoryId: number | null,
  nameId: number | null,
  loadedRows: ClinicalPlanDiagnosisRow[],
  findLabels: DiagnosisLabelResolver,
): SlotDiagnoses {
  const slotRows = loadedRows.filter((row) => row.diagnosis_slot === slot);
  const representative: ClinicalPlanDiagnosisRow | undefined = slotRows[0];
  const loadedCategoryId = toNumericId(representative?.diagnosis_type_id ?? null);
  const loadedNameId = toNumericId(representative?.diagnosis_name_id ?? null);
  if (categoryId === loadedCategoryId && nameId === loadedNameId) {
    return {
      changed: false,
      rows: slotRows.map((row) => ({
        diagnosis_slot: slot,
        diagnosis_type_id: toNumericId(row.diagnosis_type_id),
        diagnosis_name_id: toNumericId(row.diagnosis_name_id),
        diagnosis_name_snapshot: row.diagnosis_name_snapshot,
        category_name_snapshot: row.category_name_snapshot,
        sort_order: row.sort_order,
      })),
    };
  }
  if (categoryId === null && nameId === null) return { changed: true, rows: [] };
  // 新規選択の行: UI が実際に表示しているマスタ名をスナップショットにする。判明しない値は推測せず null。
  const cachedLabels = findLabels(categoryId, nameId);
  const categoryNameSnapshot =
    cachedLabels.categoryName ??
    (categoryId !== null && categoryId === loadedCategoryId
      ? (representative?.category_name_snapshot ?? null)
      : null);
  const diagnosisNameSnapshot =
    cachedLabels.diagnosisName ??
    (nameId !== null && nameId === loadedNameId
      ? (representative?.diagnosis_name_snapshot ?? null)
      : null);
  return {
    changed: true,
    rows: [
      {
        diagnosis_slot: slot,
        diagnosis_type_id: categoryId,
        diagnosis_name_id: nameId,
        diagnosis_name_snapshot: diagnosisNameSnapshot,
        category_name_snapshot: categoryNameSnapshot,
        sort_order: 0,
      },
    ],
  };
}

/**
 * 診断行の PATCH 値を作る。診断1 → slot 1、診断2 → slot 2。
 * どの slot も変わっていなければ undefined（= diagnoses キーを送らず「変更なし」）。
 * 1 つでも変わっていれば、変えていない slot の読込済み行も含めた全集合を返す（BE は全置換のため）。
 * 両 slot をユーザーがクリアした場合だけ `[]`（全行クリア）になる。
 */
function buildClinicalPlanDiagnoses(
  source: ClinicalPlanDiagnosisSource,
  findLabels: DiagnosisLabelResolver,
): ClinicalPlanDiagnosisInput[] | undefined {
  const loadedRows = source.existingClinicalPlanDiagnoses ?? [];
  const slots = [
    buildSlotDiagnoses(
      1,
      source.diagnosis1CategoryId,
      source.diagnosis1NameId,
      loadedRows,
      findLabels,
    ),
    buildSlotDiagnoses(
      2,
      source.diagnosis2CategoryId,
      source.diagnosis2NameId,
      loadedRows,
      findLabels,
    ),
  ];
  if (!slots.some((slotDiagnoses) => slotDiagnoses.changed)) return undefined;
  return slots.flatMap((slotDiagnoses) => slotDiagnoses.rows);
}

interface UseMedicalRecordSaveActionArgs {
  recordId?: string;
  activeTab: string;
  canEdit: boolean;
  isSelectedPetDeceased: boolean;
  isFinalized: boolean;
  diagnosis1CategoryId: number | null;
  diagnosis1NameId: number | null;
  diagnosis2CategoryId: number | null;
  diagnosis2NameId: number | null;
  /** 身体検査所見 → clinical_plan.physical_exam */
  physicalExam: string;
  /** 治療方針 → clinical_plan.treatment_policy */
  plan: string;
  /** 診断詳細 → clinical_plan.diagnosis_details */
  assessment: string;
  chiefComplaint: string;
  chiefComplaintDefault: string;
  chiefComplaintBaseline: string;
  chiefComplaintTypeId: number | null;
  treatmentPolicy: string;
  treatmentPolicyDefault: string;
  treatmentPolicyBaseline: string;
  existingRecordVersion?: number;
  existingClinicalPlanVersion?: number;
  /** 読込済み baseline の診断行（変更していない slot の行をそのまま送り返すために保持する）。 */
  existingClinicalPlanDiagnoses?: ClinicalPlanDiagnosisRow[];
  onClinicalPlanSaved?: (clinicalPlan: ClinicalPlan, submittedDraft: ClinicalPlanDraft) => void;
  setManualErrors: (errors: Record<string, string>) => void;
  queryClient: QueryClient;
  updateInquiryMutation: {
    mutateAsync: (variables: {
      chief_complaint?: string;
      chief_complaint_type_id: number | null;
      notes?: string;
    }) => Promise<unknown>;
  };
  updateTreatmentPlanMutation: {
    mutateAsync: (variables: {
      physical_exam: string;
      treatment_policy: string;
      diagnosis_details: string;
      diagnoses?: ClinicalPlanDiagnosisInput[];
      version?: number;
    }) => Promise<unknown>;
  };
}

export function useMedicalRecordSaveAction({
  recordId,
  activeTab,
  canEdit,
  isSelectedPetDeceased,
  isFinalized,
  diagnosis1CategoryId,
  diagnosis1NameId,
  diagnosis2CategoryId,
  diagnosis2NameId,
  physicalExam,
  plan,
  assessment,
  chiefComplaint,
  chiefComplaintDefault,
  chiefComplaintBaseline,
  chiefComplaintTypeId,
  treatmentPolicy,
  treatmentPolicyDefault,
  treatmentPolicyBaseline,
  existingRecordVersion,
  existingClinicalPlanVersion,
  existingClinicalPlanDiagnoses,
  onClinicalPlanSaved,
  setManualErrors,
  queryClient,
  updateInquiryMutation,
  updateTreatmentPlanMutation,
}: UseMedicalRecordSaveActionArgs) {
  const canEditRef = useRef(canEdit);
  const isSelectedPetDeceasedRef = useRef(isSelectedPetDeceased);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  useLayoutEffect(() => {
    isSelectedPetDeceasedRef.current = isSelectedPetDeceased;
  }, [isSelectedPetDeceased]);

  // activeTab を保存時に正確に参照するための ref（commit 直後 callback 用）
  const activeTabRef = useRef(activeTab);
  useLayoutEffect(() => {
    activeTabRef.current = activeTab;
  }, [activeTab]);

  const saveSnapshotRef = useRef({
    recordId,
    isFinalized,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
    physicalExam,
    plan,
    assessment,
    chiefComplaint,
    chiefComplaintDefault,
    chiefComplaintBaseline,
    chiefComplaintTypeId,
    treatmentPolicy,
    treatmentPolicyDefault,
    treatmentPolicyBaseline,
    existingRecordVersion,
    existingClinicalPlanVersion,
    existingClinicalPlanDiagnoses,
  });
  useLayoutEffect(() => {
    saveSnapshotRef.current = {
      recordId,
      isFinalized,
      diagnosis1CategoryId,
      diagnosis1NameId,
      diagnosis2CategoryId,
      diagnosis2NameId,
      physicalExam,
      plan,
      assessment,
      chiefComplaint,
      chiefComplaintDefault,
      chiefComplaintBaseline,
      chiefComplaintTypeId,
      treatmentPolicy,
      treatmentPolicyDefault,
      treatmentPolicyBaseline,
      existingRecordVersion,
      existingClinicalPlanVersion,
      existingClinicalPlanDiagnoses,
    };
  });

  const [formState, formAction, isSaving] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const snapshot = saveSnapshotRef.current;
      // UI の disabled は操作補助にすぎない。programmatic submit や race でも
      // 確定済み・権限なしカルテを更新しないよう action 境界で拒否する。
      if (isSelectedPetDeceasedRef.current) {
        return deniedState(DECEASED_SAVE_MESSAGE);
      }
      if (!snapshot.recordId || canEditRef.current !== true || snapshot.isFinalized) {
        return deniedState(PERMISSION_DENIED_MESSAGE);
      }

      try {
        setManualErrors({});
        const currentTab = activeTabRef.current;

        switch (currentTab) {
          case "問診":
            if (isSelectedPetDeceasedRef.current) {
              return deniedState(DECEASED_SAVE_MESSAGE);
            }
            if (canEditRef.current !== true) {
              return deniedState(PERMISSION_DENIED_MESSAGE);
            }
            await updateInquiryMutation.mutateAsync({
              // EMR-215: 入力が既定文と一致していても、保存値(baseline)が既定文と
              // 異なるなら「既定文へ戻す更新」として送信する。省略は baseline も
              // 既定文の場合（未保存・未取得）に限る。
              chief_complaint:
                snapshot.chiefComplaint !== snapshot.chiefComplaintDefault ||
                snapshot.chiefComplaintBaseline !== snapshot.chiefComplaintDefault
                  ? snapshot.chiefComplaint
                  : undefined,
              chief_complaint_type_id: snapshot.chiefComplaintTypeId,
              notes:
                snapshot.treatmentPolicy !== snapshot.treatmentPolicyDefault ||
                snapshot.treatmentPolicyBaseline !== snapshot.treatmentPolicyDefault
                  ? snapshot.treatmentPolicy
                  : undefined,
            });
            break;

          case "診察/治療プラン": {
            // BUG-010: clinical-plan GET/hydrate 前の空文字 PATCH は既存所見を無音で消す。
            // version 未確定（undefined）は BE が楽観ロックをスキップするため fail-closed で拒否する。
            if (typeof snapshot.existingClinicalPlanVersion !== "number") {
              toast.error("診察プランの読み込みが完了してから保存してください");
              return { success: false, timestamp: Date.now() };
            }
            if (snapshot.diagnosis1CategoryId && !snapshot.diagnosis1NameId) {
              const diagError = { diagnosis1_name_id: "診断名を選択してください" };
              setManualErrors(diagError);
              return { success: false, fieldErrors: diagError, timestamp: Date.now() };
            }
            // BUG-416 ②: diagnosis1 と同じバリデーションを diagnosis2 にも適用する（FE validation parity）
            if (snapshot.diagnosis2CategoryId && !snapshot.diagnosis2NameId) {
              const diagError = { diagnosis2_name_id: "診断名を選択してください" };
              setManualErrors(diagError);
              return { success: false, fieldErrors: diagError, timestamp: Date.now() };
            }
            // BUG-010 / BUG-102: 3欄は常に送信する（undefined 欠落は「未更新」になり、
            // テンプレ既定や別 writer の last-write-wins で入力が消える）。空文字は明示クリア。
            const treatmentPlanPayload = {
              physical_exam: snapshot.physicalExam,
              treatment_policy: snapshot.plan,
              diagnosis_details: snapshot.assessment,
              // EMR-495: 診断は diagnoses 行配列。変更した slot だけ置換し、変えていない slot の読込済み行は
              // 値のまま送り返す。どの slot も変えていなければ undefined（キーを送らない = 変更なし）。
              // 新規選択の診断は、保存時に UI が表示しているマスタ名（react-query キャッシュ）をスナップショットに載せる。
              diagnoses: buildClinicalPlanDiagnoses(snapshot, (categoryId, nameId) =>
                findCachedDiagnosisLabels(queryClient, categoryId, nameId),
              ),
              // BUG-416③: clinical_plan 楽観ロック。undefined を送ると BE は
              // バージョンチェックをスキップする（後方互換）ため常に送信する。
              version: snapshot.existingClinicalPlanVersion,
            };
            const submittedClinicalPlanDraft: ClinicalPlanDraft = {
              physicalExam: snapshot.physicalExam,
              plan: snapshot.plan,
              assessment: snapshot.assessment,
              diagnosis1CategoryId: snapshot.diagnosis1CategoryId,
              diagnosis1NameId: snapshot.diagnosis1NameId,
              diagnosis2CategoryId: snapshot.diagnosis2CategoryId,
              diagnosis2NameId: snapshot.diagnosis2NameId,
            };
            if (isSelectedPetDeceasedRef.current) {
              return deniedState(DECEASED_SAVE_MESSAGE);
            }
            if (canEditRef.current !== true) {
              return deniedState(PERMISSION_DENIED_MESSAGE);
            }
            const savedClinicalPlan =
              await updateTreatmentPlanMutation.mutateAsync(treatmentPlanPayload);
            if (isClinicalPlan(savedClinicalPlan)) {
              onClinicalPlanSaved?.(savedClinicalPlan, submittedClinicalPlanDraft);
            }
            // EMR-308: 次回来院日の意思は tentative appointment（NextVisitButton 経路）が正本。
            // 旧カルテの来院推奨日カラムは撤廃済みのため、ここではカルテ PATCH を発行しない。
            break;
          }

          case "見積書":
            // BUG-016: 見積の永続化は post-save の estimateSave が正本。
            // ここで汎用「保存しました」を出すと API 未送信でも成功に見える。
            // 成功トーストは MedicalRecordEstimate が実 API 成功時のみ出す。
            queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
            return { success: true, timestamp: Date.now() };

          case "予防接種":
            // BUG-001: 永続化は inner VaccinationForm の createVaccination が正本。
            // success:true だと useMedicalRecordPostSave が markClean し、
            // 未保存の問診/診察 dirty を偽クリーンする。トーストも出さない。
            return { success: false, timestamp: Date.now() };

          default:
            // 治療 / 定期健診 / 検査 / 画像は子フォームが正本。未送信を成功にしない。
            return { success: false, timestamp: Date.now() };
        }

        toast.success("保存しました");
        queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE,
  );

  return { formState, formAction, isSaving };
}
