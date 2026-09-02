import { useState } from "react";

import { Button } from "@/components/ui/button";
import { usePermission } from "@/hooks/use-permission";
import { C } from "@/lib/design-tokens";

import {
  useAttachLabDeviceJob,
  useDetachLabDeviceJob,
  useGetLabDeviceUnlinked,
  type LabDeviceJobCard,
} from "@/features/lab-device/api/lab-device";
import {
  isLabDeviceAttachPersisted,
  labDeviceCardTitle,
  labDeviceClockSkewLabel,
  labDeviceNeedsReviewReason,
} from "@/features/lab-device/lib/lab-device-board-model";

// components/shared は @/types/generated/models の import allowlist 対象外（TASK-444-S1）。
// リテラル値は models.ts の `ResourceLabImport = "lab-import"` と同値（usePermission の
// Resource union に構造的に一致するため型 import 無しで済む）。
const LAB_IMPORT_RESOURCE = "lab-import";

export function LabDeviceUnlinkedBanner({ petId }: { petId: string }) {
  const numericPetId = Number(petId);
  const { canView, canEdit } = usePermission(LAB_IMPORT_RESOURCE);
  const { data: unlinked = [] } = useGetLabDeviceUnlinked(canView && Number.isFinite(numericPetId) && numericPetId > 0);
  const attach = useAttachLabDeviceJob();
  const detach = useDetachLabDeviceJob();
  const [justAttached, setJustAttached] = useState<LabDeviceJobCard | null>(null);
  const [attachError, setAttachError] = useState<string | null>(null);

  const visibleUnlinked = unlinked.filter((card) => card.jobId !== justAttached?.jobId);

  if (!canView || (visibleUnlinked.length === 0 && justAttached == null && attachError == null)) {
    return null;
  }

  return (
    <section className={`rounded-lg border ${C.borderNotice} ${C.bgWhite} p-3 space-y-2`} aria-live="polite">
      <p className="font-medium">未紐付けの受信があります</p>
      <ul className="space-y-2">
        {justAttached ? (
          <li className="flex items-center justify-between gap-3">
            <span className={C.textInkMuted}>
              {labDeviceCardTitle(justAttached)}
              {" · 付けました"}
              {labDeviceClockSkewLabel(justAttached) ? ` · ${labDeviceClockSkewLabel(justAttached)}` : ""}
            </span>
            {canEdit ? (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => {
                  void detach.mutateAsync(justAttached.jobId).then(() => setJustAttached(null));
                }}
              >
                取り消す
              </Button>
            ) : null}
          </li>
        ) : null}
        {attachError ? (
          <li>
            <span className={`text-sm ${C.textRed700}`}>{attachError}</span>
          </li>
        ) : null}
        {visibleUnlinked.map((card) => (
          <li key={card.jobId} className="flex items-center justify-between gap-3">
            <span className={C.textInkMuted}>
              {labDeviceCardTitle(card)}
              {card.unmappedItemCount > 0 ? " · 未対応あり" : ""}
              {labDeviceNeedsReviewReason(card) ? ` · ${labDeviceNeedsReviewReason(card)}` : ""}
              {labDeviceClockSkewLabel(card) ? ` · ${labDeviceClockSkewLabel(card)}` : ""}
            </span>
            {canEdit ? (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => {
                  setAttachError(null);
                  void attach.mutateAsync({ jobId: card.jobId, petId: numericPetId }).then((attached) => {
                    if (isLabDeviceAttachPersisted(attached)) {
                      setJustAttached(attached);
                    } else {
                      const reviewMsg = labDeviceNeedsReviewReason(attached);
                      setAttachError(
                        reviewMsg
                          ? `保存できませんでした（${reviewMsg}）`
                          : "保存できませんでした。未紐付けのままです",
                      );
                    }
                  });
                }}
              >
                付ける
              </Button>
            ) : null}
          </li>
        ))}
      </ul>
    </section>
  );
}
