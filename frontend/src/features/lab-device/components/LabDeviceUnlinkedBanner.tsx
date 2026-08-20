import { useState } from "react";

import { Button } from "@/components/ui/button";
import { usePermission } from "@/hooks/use-permission";
import { C } from "@/lib/design-tokens";
import { ResourceLabImport } from "@/types/generated/models";

import {
  useAttachLabDeviceJob,
  useDetachLabDeviceJob,
  useGetLabDeviceUnlinked,
  type LabDeviceJobCard,
} from "../api/lab-device";
import { labDeviceCardTitle, labDeviceClockSkewLabel } from "../lib/lab-device-board-model";

export function LabDeviceUnlinkedBanner({ petId }: { petId: string }) {
  const numericPetId = Number(petId);
  const { canView, canEdit } = usePermission(ResourceLabImport);
  const { data: unlinked = [] } = useGetLabDeviceUnlinked(canView && Number.isFinite(numericPetId) && numericPetId > 0);
  const attach = useAttachLabDeviceJob();
  const detach = useDetachLabDeviceJob();
  const [justAttached, setJustAttached] = useState<LabDeviceJobCard | null>(null);

  const visibleUnlinked = unlinked.filter((card) => card.jobId !== justAttached?.jobId);

  if (!canView || (visibleUnlinked.length === 0 && justAttached == null)) {
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
        {visibleUnlinked.map((card) => (
          <li key={card.jobId} className="flex items-center justify-between gap-3">
            <span className={C.textInkMuted}>
              {labDeviceCardTitle(card)}
              {card.unmappedItemCount > 0 ? " · 未対応あり" : ""}
              {labDeviceClockSkewLabel(card) ? ` · ${labDeviceClockSkewLabel(card)}` : ""}
            </span>
            {canEdit ? (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => {
                  void attach.mutateAsync({ jobId: card.jobId, petId: numericPetId }).then((attached) => {
                    setJustAttached(attached);
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
