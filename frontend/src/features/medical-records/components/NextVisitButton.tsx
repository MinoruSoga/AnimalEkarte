import { lazy, Suspense, useCallback, useMemo, useState } from "react";
import { CalendarPlus, Calendar, CalendarCheck2, Pencil } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Tooltip } from "@/components/ui/tooltip";
import { toast } from "sonner";
import { C, ICON } from "@/lib/design-tokens";
import { useConfirmTentative } from "@/hooks/use-confirm-tentative";
import { buildTentativeConfirmInitialData } from "@/components/shared/ReservationFormModal/confirm-tentative";
import type { Reservation } from "@/types";
import { useClearNextVisitIntent, useSetNextVisitIntent } from "../api/next-visit-intent";
import { useNextVisitIntent } from "../hooks/use-next-visit-intent";
import { NextVisitDateField } from "./NextVisitDateField";

// EMR-308 A1: tentative 確定は共有 ReservationFormModal を遅延ロードで再利用する
const ReservationFormModal = lazy(() =>
  import("@/components/shared/ReservationFormModal/ReservationFormModal").then((m) => ({
    default: m.ReservationFormModal,
  })),
);

interface NextVisitButtonProps {
  petId: string;
  /** 保存済みカルテの id。PUT 時に source_medical_record_id として送る。 */
  medicalRecordId?: string;
  /** EMR-308/Codex: 拠点横断カルテ（record.clinicId）ではその医院を X-Clinic-ID として送る */
  clinicId?: string;
  isPetDeceased?: boolean;
  hasLineIntegration?: boolean;
}

/**
 * EMR-308: ヘッダー「次回予定」コントロール。
 * 表示はペットの予約（実予約 > tentative > 未設定）から導出し、
 * Popover の日付入力は tentative appointment の upsert/除去を行う。
 */
export function NextVisitButton({
  petId,
  medicalRecordId,
  clinicId,
  isPetDeceased = false,
  hasLineIntegration,
}: NextVisitButtonProps) {
  const setIntentMutation = useSetNextVisitIntent(clinicId);
  const clearIntentMutation = useClearNextVisitIntent(clinicId);
  const confirmTentativeMutation = useConfirmTentative(clinicId);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const {
    nextVisitLabel,
    intentFieldValue,
    handleIntentDateChange,
    resetIntentDraft,
    canMutateIntent,
    canConfirmTentative,
    tentativeReservation,
  } = useNextVisitIntent({
    petId,
    medicalRecordId,
    clinicId,
    isPetDeceased,
    setIntentMutation,
    clearIntentMutation,
  });

  // tentative 行の意図日・区分・担当医を初期値にした確定フォーム。
  // petId は tentative 行のものを優先（行と患者の不一致を防ぐ）。
  const confirmInitialData = useMemo(
    () =>
      tentativeReservation
        ? buildTentativeConfirmInitialData(
            tentativeReservation,
            tentativeReservation.petId ?? petId,
          )
        : null,
    [tentativeReservation, petId],
  );

  const handleConfirmTentative = useCallback(
    async (tentativeId: string, data: Partial<Reservation>) => {
      try {
        await confirmTentativeMutation.mutateAsync({
          reservationId: tentativeId,
          formData: data,
        });
      } catch {
        // onError(handleApiError) が toast 済み。モーダル内にもインライン表示して開いたままにする。
        return "予約の確定に失敗しました。入力内容を確認して再度お試しください。";
      }
      toast.success("次回予定を確定しました");
      setConfirmOpen(false);
      return null;
    },
    [confirmTentativeMutation],
  );

  const hasValue = nextVisitLabel !== "";
  const tooltipText = hasValue ? "次回予定を変更" : "次回予定を追加";
  const ariaLabel = tooltipText;

  return (
    <div className="flex flex-col gap-0 shrink-0">
      <span className={`text-xs ${C.text50}`}>次回予定</span>
      <div className="flex items-center gap-1">
        <Popover
          onOpenChange={(open) => {
            if (!open) resetIntentDraft();
          }}
        >
          <Tooltip content={tooltipText}>
            <PopoverTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                disabled={!canMutateIntent}
                aria-label={ariaLabel}
                className={`h-8 px-2 gap-1 text-sm font-medium ${C.text} ${C.hoverBgPage} border-none`}
              >
                {hasValue ? (
                  <>
                    <Calendar className={`${ICON.sm} ${C.text60}`} aria-hidden="true" />
                    <span>{nextVisitLabel}</span>
                    <Pencil className={`${ICON.sm} ${C.text40}`} aria-hidden="true" />
                  </>
                ) : (
                  <CalendarPlus className={`${ICON.sm} ${C.text40}`} aria-hidden="true" />
                )}
              </Button>
            </PopoverTrigger>
          </Tooltip>
          <PopoverContent className="w-auto p-4" align="start">
            <NextVisitDateField
              value={intentFieldValue}
              onChange={handleIntentDateChange}
              hasLineIntegration={hasLineIntegration}
              disabled={!canMutateIntent}
            />
          </PopoverContent>
        </Popover>

        {/* EMR-308 A1: tentative が表示中なら「予約を確定」で確定モーダルを開く */}
        {canConfirmTentative ? (
          <Tooltip content="この次回予定を本予約に確定">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setConfirmOpen(true)}
              aria-label="次回予定を確定"
              className={`h-8 px-2 gap-1 text-xs font-medium ${C.textActionPrimary} ${C.hoverBgPage} border-none`}
            >
              <CalendarCheck2 className={ICON.sm} aria-hidden="true" />
              <span>予約を確定</span>
            </Button>
          </Tooltip>
        ) : null}
      </div>

      {tentativeReservation ? (
        <Suspense fallback={null}>
          <ReservationFormModal
            isOpen={confirmOpen}
            onClose={() => setConfirmOpen(false)}
            tentativeConfirmId={tentativeReservation.id}
            onConfirmTentative={handleConfirmTentative}
            initialData={confirmInitialData}
            canCreate={false}
            // POST confirm-tentative は reservations:edit 権限。死亡ペットは canConfirmTentative
            // で既に弾いているが、モーダル側にも fail-closed で渡す。
            canEdit={canConfirmTentative}
            // Codex PR606 再指摘: 拠点横断カルテでは tentative 行の医院で
            // マスタ/空き枠/休診日を引く（mutation 側は hook 引数で同値済み）
            clinicId={tentativeReservation.clinicId ?? clinicId}
          />
        </Suspense>
      ) : null}
    </div>
  );
}
