import { memo } from "react";

import BedDouble from "lucide-react/dist/esm/icons/bed-double";
import Clock from "lucide-react/dist/esm/icons/clock";
import CreditCard from "lucide-react/dist/esm/icons/credit-card";
import Dog from "lucide-react/dist/esm/icons/dog";
import ExternalLink from "lucide-react/dist/esm/icons/external-link";
import FileText from "lucide-react/dist/esm/icons/file-text";
import Scissors from "lucide-react/dist/esm/icons/scissors";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import User from "lucide-react/dist/esm/icons/user";

import { Badge } from "@/components/ui/badge";
import { DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { C, ICON } from "@/lib/design-tokens";
import {
  RECEPTION_STATUS_COLOR_FALLBACK,
  RECEPTION_STATUS_COLORS,
} from "@/constants/status-colors";

import { ActionButtons, type ActionButtonsProps } from "./ReceptionDialogActionButtons";
import type { ReceptionAppointment as Appointment } from "../api/types";

const SECTION_LABEL = `text-2xs font-semibold ${C.text60} uppercase`;
const DIVIDER_ROW = `flex items-center justify-between border-b ${C.borderLight} pb-2`;
const ROW_ICON = `flex items-center gap-2 ${C.text60}`;

const RELATED_BTN_BASE =
  "flex items-center gap-1.5 text-sm border rounded-md px-3 py-1.5 transition-colors group";
const RELATED_BTN_KARTE = `${RELATED_BTN_BASE} ${C.textBrand} ${C.bgBrandLight40} ${C.hoverBgBrandLight} ${C.borderBrandLight}`;
const RELATED_BTN_ACCOUNTING = `${RELATED_BTN_BASE} ${C.textStatusGreen} ${C.bgStatusGreen40} ${C.hoverBgStatusGreen} ${C.borderStatusGreen}`;
const RELATED_BTN_HOSPITAL = `${RELATED_BTN_BASE} ${C.textStatusPurple} ${C.bgStatusPurple40} ${C.hoverBgStatusPurple} ${C.borderStatusPurple}`;

interface ReceptionDialogHeaderProps {
  appointment: Appointment;
  currentStatus?: string;
}

export function ReceptionDialogHeader({
  appointment,
  currentStatus,
}: ReceptionDialogHeaderProps) {
  const visitTypeClass =
    appointment.visitType === "初診"
      ? `${C.bgBrandLight} ${C.textBrand}`
      : `${C.bgActive} ${C.text60}`;

  return (
    <DialogHeader className={`p-6 pb-4 border-b ${C.borderLight} pr-12`}>
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <span
            className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${visitTypeClass}`}
          >
            {appointment.visitType === "初診" ? "初" : "再"}
          </span>
          <DialogTitle className={`text-xl font-bold ${C.text}`}>
            {appointment.reservationType}
          </DialogTitle>
        </div>
        {currentStatus ? (
          <Badge
            variant="outline"
            className={`${
              RECEPTION_STATUS_COLORS[currentStatus] ?? RECEPTION_STATUS_COLOR_FALLBACK
            } px-3 py-1 text-sm font-medium border shrink-0`}
          >
            {currentStatus}
          </Badge>
        ) : null}
      </div>
      <DialogDescription className="sr-only">
        予約の詳細情報
      </DialogDescription>
    </DialogHeader>
  );
}

interface RelatedPagesProps {
  isTrimming: boolean;
  onCreateMedicalRecord: (tab?: string) => void;
  onCreateTrimming: () => void;
  onCreateAccounting: () => void;
  onCreateHospitalization: () => void;
  canCreateMedicalRecord?: boolean;
  canCreateAccounting?: boolean;
  canCreateHospitalization?: boolean;
}

const RelatedPages = memo(function RelatedPages({
  isTrimming,
  onCreateMedicalRecord,
  onCreateTrimming,
  onCreateAccounting,
  onCreateHospitalization,
  canCreateMedicalRecord = false,
  canCreateAccounting = false,
  canCreateHospitalization = false,
}: RelatedPagesProps) {
  return (
    <div className="space-y-2">
      <h3 className={SECTION_LABEL}>関連ページ</h3>
      <div className="flex flex-wrap gap-2">
        {canCreateMedicalRecord ? (
          <button
            type="button"
            className={RELATED_BTN_KARTE}
            onClick={() => {
              if (isTrimming) onCreateTrimming();
              else onCreateMedicalRecord();
            }}
          >
            {isTrimming ? (
              <Scissors className={ICON.xs} />
            ) : (
              <FileText className={ICON.xs} />
            )}
            <span>{isTrimming ? "施術" : "カルテ"}</span>
            <ExternalLink className={`${ICON.xs} opacity-0 group-hover:opacity-100 transition-opacity`} />
          </button>
        ) : null}

        {canCreateAccounting ? (
          <button type="button" className={RELATED_BTN_ACCOUNTING} onClick={onCreateAccounting}>
            <CreditCard className={ICON.xs} />
            <span>会計</span>
            <ExternalLink className={`${ICON.xs} opacity-0 group-hover:opacity-100 transition-opacity`} />
          </button>
        ) : null}

        {canCreateHospitalization ? (
          <button
            type="button"
            className={RELATED_BTN_HOSPITAL}
            onClick={onCreateHospitalization}
          >
            <BedDouble className={ICON.xs} />
            <span>入院</span>
            <ExternalLink className={`${ICON.xs} opacity-0 group-hover:opacity-100 transition-opacity`} />
          </button>
        ) : null}
      </div>
    </div>
  );
});

interface ReceptionDialogBodyProps extends RelatedPagesProps {
  appointment: Appointment;
}

export function ReceptionDialogBody({
  appointment,
  isTrimming,
  onCreateMedicalRecord,
  onCreateTrimming,
  onCreateAccounting,
  onCreateHospitalization,
  canCreateMedicalRecord,
  canCreateAccounting,
  canCreateHospitalization,
}: ReceptionDialogBodyProps) {
  return (
    <div className="p-6 space-y-4 overflow-y-auto">
      <div className={`flex items-center gap-3 p-3 ${C.bgPage} rounded-lg`}>
        <Clock className={`${ICON.page} ${C.text60} shrink-0`} />
        <span className={`font-mono text-xl font-medium ${C.text}`}>{appointment.time}</span>
      </div>

      <div className="space-y-3">
        <h3 className={SECTION_LABEL}>患者情報</h3>
        <div className={DIVIDER_ROW}>
          <div className={ROW_ICON}>
            <Dog className={ICON.action} />
            <span className="text-sm">ペット</span>
          </div>
          <div className="text-right">
            <div className="font-bold text-base">{appointment.petName}</div>
            <div className={`text-sm ${C.text60}`}>{appointment.petType}</div>
          </div>
        </div>
        <div className={DIVIDER_ROW}>
          <div className={ROW_ICON}>
            <User className={ICON.action} />
            <span className="text-sm">飼い主</span>
          </div>
          <span className={`font-medium ${C.text}`}>{appointment.ownerName}</span>
        </div>
      </div>

      <div className="space-y-3">
        <h3 className={SECTION_LABEL}>診療詳細</h3>
        <div className={DIVIDER_ROW}>
          <div className={ROW_ICON}>
            <Stethoscope className={ICON.action} />
            <span className="text-sm">担当医</span>
          </div>
          <div className="flex items-center gap-2">
            <span className={`font-medium ${C.text}`}>{appointment.doctor || "未定"}</span>
            {appointment.isDesignated ? (
              <Badge
                variant="outline"
                className={`text-xs h-6 ${C.bgDiscountLight} ${C.textDiscount} ${C.borderDiscount20}`}
              >
                指名
              </Badge>
            ) : null}
          </div>
        </div>
      </div>

      <RelatedPages
        isTrimming={isTrimming}
        onCreateMedicalRecord={onCreateMedicalRecord}
        onCreateTrimming={onCreateTrimming}
        onCreateAccounting={onCreateAccounting}
        onCreateHospitalization={onCreateHospitalization}
        canCreateMedicalRecord={canCreateMedicalRecord}
        canCreateAccounting={canCreateAccounting}
        canCreateHospitalization={canCreateHospitalization}
      />
    </div>
  );
}

type ReceptionDialogFooterProps = ActionButtonsProps;

export function ReceptionDialogFooter(props: ReceptionDialogFooterProps) {
  return (
    <DialogFooter className={`p-4 ${C.bgPage}`}>
      <div className="flex flex-wrap gap-2 justify-end w-full">
        <ActionButtons {...props} />
      </div>
    </DialogFooter>
  );
}
