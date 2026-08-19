import { Button } from "@/components/ui/button";
import { usePermission } from "@/hooks/use-permission";
import { C } from "@/lib/design-tokens";
import { ResourceLabImport } from "@/types/generated/models";

import { useAttachLabDeviceJob, useGetLabDeviceUnlinked } from "../api/lab-device";
import { labDeviceCardTitle } from "../lib/lab-device-board-model";

export function LabDeviceUnlinkedBanner({ petId }: { petId: string }) {
  const numericPetId = Number(petId);
  const { canView, canEdit } = usePermission(ResourceLabImport);
  const { data: unlinked = [] } = useGetLabDeviceUnlinked(canView && Number.isFinite(numericPetId) && numericPetId > 0);
  const attach = useAttachLabDeviceJob();

  if (!canView || unlinked.length === 0) {
    return null;
  }

  return (
    <section className={`rounded-lg border ${C.borderNotice} ${C.bgWhite} p-3 space-y-2`} aria-live="polite">
      <p className="font-medium">未紐付けの受信があります</p>
      <ul className="space-y-2">
        {unlinked.map((card) => (
          <li key={card.jobId} className="flex items-center justify-between gap-3">
            <span className={C.textInkMuted}>
              {labDeviceCardTitle(card)}
              {card.unmappedItemCount > 0 ? " · 未対応あり" : ""}
            </span>
            {canEdit ? (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => void attach.mutateAsync({ jobId: card.jobId, petId: numericPetId })}
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
