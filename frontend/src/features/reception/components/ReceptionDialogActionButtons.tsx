import { memo } from "react";

import BedDouble from "lucide-react/dist/esm/icons/bed-double";
import CreditCard from "lucide-react/dist/esm/icons/credit-card";
import FileText from "lucide-react/dist/esm/icons/file-text";
import Pencil from "lucide-react/dist/esm/icons/pencil";
import Scissors from "lucide-react/dist/esm/icons/scissors";
import TestTube from "lucide-react/dist/esm/icons/test-tube";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import User from "lucide-react/dist/esm/icons/user";

import { Button } from "@/components/ui/button";
import { PRIMARY_ACTION_BUTTON_CLASSES } from "@/components/shared/Form/primary-action-button-classes";
import { C, ICON } from "@/lib/design-tokens";

import type { ReceptionAppointment as Appointment } from "../api/types";

export interface ActionButtonsProps {
  currentStatus?: string;
  appointment: Appointment;
  isTrimming: boolean;
  isHospitalization: boolean;
  isMedical: boolean;
  onConfirm?: () => void;
  onEdit?: (appointment: Appointment) => void;
  onCancel?: (appointment: Appointment) => void;
  onOpenOwnerDetail: () => void;
  onCreateMedicalRecord: (tab?: string) => void;
  onCreateTrimming: () => void;
  onCreateAccounting: () => void;
  onCreateHospitalization: () => void;
}

export const ActionButtons = memo(function ActionButtons({
  currentStatus,
  appointment,
  isTrimming,
  isHospitalization,
  isMedical,
  onConfirm,
  onEdit,
  onCancel,
  onOpenOwnerDetail,
  onCreateMedicalRecord,
  onCreateTrimming,
  onCreateAccounting,
  onCreateHospitalization,
}: ActionButtonsProps) {
  const ownerDetailBtn = (
    <Button variant="ghost" onClick={onOpenOwnerDetail} className={`h-10 text-sm ${C.text}`}>
      <User className={ICON.action} />
      飼主詳細
    </Button>
  );

  if (currentStatus === "受付予約") {
    return (
      <>
        {onCancel ? (
          <Button
            variant="ghost"
            onClick={() => onCancel(appointment)}
            className={`h-10 text-sm ${C.danger} ${C.hoverBgDanger5}`}
          >
            <Trash2 className={ICON.action} />
            取消
          </Button>
        ) : null}
        {onEdit ? (
          <Button
            variant="outline"
            onClick={() => onEdit(appointment)}
            className={`h-10 text-sm ${C.text} ${C.borderMedium}`}
          >
            <Pencil className={ICON.action} />
            編集
          </Button>
        ) : null}
        {ownerDetailBtn}
        {onConfirm ? (
          <Button onClick={onConfirm} className={PRIMARY_ACTION_BUTTON_CLASSES}>
            受付済にする
          </Button>
        ) : null}
      </>
    );
  }

  if (currentStatus === "受付済") {
    return (
      <>
        {ownerDetailBtn}
        {isTrimming ? (
          <div className="flex flex-col items-end gap-1">
            <Button
              onClick={() => {
                if (onConfirm) onConfirm();
                onCreateTrimming();
              }}
              className={PRIMARY_ACTION_BUTTON_CLASSES}
            >
              <Scissors className={ICON.action} />
              トリミングカルテ作成
            </Button>
            <span className={`text-2xs ${C.text40}`}>
              ※カルテ作成と同時に「診療中」へ移動します
            </span>
          </div>
        ) : isMedical ? (
          <div className="flex flex-col items-end gap-1">
            <Button
              onClick={() => {
                if (onConfirm) onConfirm();
                onCreateMedicalRecord();
              }}
              className={PRIMARY_ACTION_BUTTON_CLASSES}
            >
              <FileText className={ICON.action} />
              カルテ作成
            </Button>
            <span className={`text-2xs ${C.text40}`}>
              ※カルテ作成と同時に「診療中」へ移動します
            </span>
          </div>
        ) : onConfirm ? (
          <Button onClick={onConfirm} className={PRIMARY_ACTION_BUTTON_CLASSES}>
            診察を開始する
          </Button>
        ) : null}
      </>
    );
  }

  if (currentStatus === "診療中") {
    return (
      <>
        {ownerDetailBtn}
        {onConfirm ? (
          <Button
            variant="outline"
            onClick={onConfirm}
            className={`h-10 text-sm ${C.danger} ${C.borderDanger} ${C.hoverBgDanger5}`}
          >
            診察を終了する
          </Button>
        ) : null}
        {isMedical ? (
          <>
            <Button onClick={() => onCreateMedicalRecord()} className={PRIMARY_ACTION_BUTTON_CLASSES}>
              <FileText className={ICON.action} />
              カルテ入力
            </Button>
            <Button
              variant="outline"
              onClick={() => onCreateMedicalRecord("test")}
              className={`h-10 text-sm ${C.text} ${C.borderMedium}`}
            >
              <TestTube className={ICON.action} />
              検査
            </Button>
          </>
        ) : null}
      </>
    );
  }

  if (currentStatus === "会計待ち") {
    return (
      <>
        {ownerDetailBtn}
        <Button onClick={onCreateAccounting} className={PRIMARY_ACTION_BUTTON_CLASSES}>
          <CreditCard className={ICON.action} />
          会計へ進む
        </Button>
      </>
    );
  }

  if (currentStatus === "会計済") {
    return (
      <>
        {ownerDetailBtn}
        {onConfirm ? (
          <Button onClick={onConfirm} className={PRIMARY_ACTION_BUTTON_CLASSES}>
            完了/リストから削除
          </Button>
        ) : null}
      </>
    );
  }

  return (
    <>
      {ownerDetailBtn}
      {isMedical ? (
        <Button onClick={() => onCreateMedicalRecord()} className={PRIMARY_ACTION_BUTTON_CLASSES}>
          <FileText className={ICON.action} />
          カルテ確認
        </Button>
      ) : null}
      {isTrimming ? (
        <Button onClick={onCreateTrimming} className={PRIMARY_ACTION_BUTTON_CLASSES}>
          <Scissors className={ICON.action} />
          トリミング
        </Button>
      ) : null}
      {isHospitalization ? (
        <Button onClick={onCreateHospitalization} className={PRIMARY_ACTION_BUTTON_CLASSES}>
          <BedDouble className={ICON.action} />
          入院登録
        </Button>
      ) : null}
      {onConfirm ? (
        <Button onClick={onConfirm} className={PRIMARY_ACTION_BUTTON_CLASSES}>
          ステータス変更
        </Button>
      ) : null}
    </>
  );
});
