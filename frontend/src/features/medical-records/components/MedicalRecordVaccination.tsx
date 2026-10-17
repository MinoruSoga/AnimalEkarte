// React/Framework
import { memo } from "react";

// Internal
import { BADGE, C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";
import { Plus } from "lucide-react";
import { getLockedEmptyStateMessage } from "../lib/record-empty-state";

// Relative
import { EmptyState } from "@/components/shared/DataStates";
import { MEDICAL_RECORD_FIXED_ACTION_LAYOUT } from "../constants/fixed-action-layout";
import { VaccinationForm } from "./VaccinationForm";
import { VaccinationHistory } from "./VaccinationHistory";
import { useMedicalRecordVaccinationForm } from "../hooks/use-medical-record-vaccination-form";

type LstepStatus = "synced" | "not-linked" | "opt-out";

interface MedicalRecordVaccinationProps {
  petId?: string;
  medicalRecordId?: string;
  lstepStatus?: LstepStatus;
  /** BUG-035: 確定済み/送信権限なし。記録追加・複写のみ無効化し履歴閲覧は残す。 */
  isLocked?: boolean;
  /** EMR-412 K5: 確定済み判定（K5 空状態文言の優先理由）。既定 false。 */
  isFinalized?: boolean;
}

function LstepStatusBadge({ status }: { status: LstepStatus }) {
  if (status === "synced") {
    return (
      <span
        className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full text-white"
        style={{ backgroundColor: PALETTE.lineGreen }}
      >
        LINE通知対象
      </span>
    );
  }
  if (status === "not-linked") {
    return (
      <span
        className={`inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full border ${BADGE.yellow}`}
      >
        LINE未連携
      </span>
    );
  }
  return (
    <span
      className={`inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full border ${C.text40} ${C.borderMediumLight} ${C.bgPage30}`}
    >
      LINE受信拒否
    </span>
  );
}

export const MedicalRecordVaccination = memo(function MedicalRecordVaccination({
  petId,
  medicalRecordId,
  lstepStatus,
  isLocked = false,
  isFinalized = false,
}: MedicalRecordVaccinationProps) {
  // FE-RC-008: 埋め込みフォームの状態と保存アクションは extract hook に集約。
  // 本体は useActionState の formAction を <form action> に渡すだけの薄い belt。
  const {
    isAdding,
    setIsAdding,
    vaccineOptions,
    vaccineName,
    setVaccineName,
    date,
    setDate,
    supplemental,
    setSupplemental,
    lot1,
    setLot1,
    lot2,
    setLot2,
    lot3,
    setLot3,
    lot4,
    setLot4,
    nextScheduleType,
    setNextScheduleType,
    nextDate,
    setNextDate,
    remarks,
    setRemarks,
    fieldErrors,
    formAction,
    handleDuplicate,
    historyItems,
    isLoading,
  } = useMedicalRecordVaccinationForm(petId, medicalRecordId);

  // jsx-no-leaked-render: JSX 属性内の && 連鎖を避けて変数化
  const canCreateVaccination = !!petId && !isLocked;
  // EMR-412 K5: ロック中（確定済/編集権限なし）は理由で文言を分ける。
  const lockedEmptyMessage = getLockedEmptyStateMessage({ isFinalized, isLocked });

  return (
    <>
      {lstepStatus !== undefined ? (
        <div className="flex items-center gap-2 px-1 py-1.5 shrink-0">
          <LstepStatusBadge status={lstepStatus} />
        </div>
      ) : null}
      {/* EMR-212: MedicalRecordFormReadyPanels がページ全体を <form action>（カルテ保存）で
          包むため、ここに <form> を置くとネスト form となりブラウザが破棄する。form 要素は
          使わず、VaccinationForm 内の SubmitButton の formAction で送信する（EMR-208 と同型） */}
      <div
        className={`grid grid-cols-1 gap-4 flex-1 min-h-0 overflow-y-auto relative ${MEDICAL_RECORD_FIXED_ACTION_LAYOUT.scrollBottomClearance} pr-1 lg:grid-cols-5`}
      >
        <fieldset disabled={isLocked} className={`contents ${STYLE.readonlyInputs}`}>
          {isAdding ? (
            <VaccinationForm
              formAction={formAction}
              vaccineOptions={vaccineOptions}
              vaccineName={vaccineName}
              setVaccineName={setVaccineName}
              date={date}
              setDate={setDate}
              supplemental={supplemental}
              setSupplemental={setSupplemental}
              lot1={lot1}
              setLot1={setLot1}
              lot2={lot2}
              setLot2={setLot2}
              lot3={lot3}
              setLot3={setLot3}
              lot4={lot4}
              setLot4={setLot4}
              nextScheduleType={nextScheduleType}
              setNextScheduleType={setNextScheduleType}
              nextDate={nextDate}
              setNextDate={setNextDate}
              remarks={remarks}
              setRemarks={setRemarks}
              fieldErrors={fieldErrors}
            />
          ) : historyItems.length > 0 ? (
            <div className="lg:col-span-3 flex flex-col gap-3">
              <ul
                className={`divide-y ${C.borderLight} ${C.bgWhite} rounded-lg border ${C.borderMedium}`}
              >
                {historyItems.map((item) => (
                  <li key={item.id} className="px-3 py-2 text-sm">
                    <div className={`font-medium ${C.text}`}>{item.name}</div>
                    <div className={C.text60}>接種日 {item.date}</div>
                  </li>
                ))}
              </ul>
              {canCreateVaccination ? (
                <button
                  type="button"
                  className={STYLE.inlineAddBtn}
                  onClick={() => setIsAdding(true)}
                >
                  <Plus className={ICON.xs} />
                  <span>記録を追加</span>
                </button>
              ) : null}
            </div>
          ) : (
            <EmptyState
              className="lg:col-span-3"
              message={
                lockedEmptyMessage ??
                "接種記録がありません。「記録を追加」ボタンから追加してください。"
              }
            >
              {/* EMR-412 K4: 追加ボタンは STYLE.inlineAddBtn に統一。ロック中は出さない。 */}
              {canCreateVaccination ? (
                <button
                  type="button"
                  className={STYLE.inlineAddBtn}
                  onClick={() => setIsAdding(true)}
                >
                  <Plus className={ICON.xs} />
                  <span>記録を追加</span>
                </button>
              ) : null}
            </EmptyState>
          )}
        </fieldset>

        <VaccinationHistory
          historyItems={historyItems}
          isLoading={isLoading}
          onDuplicate={handleDuplicate}
          canCreate={canCreateVaccination}
        />
      </div>
    </>
  );
});
