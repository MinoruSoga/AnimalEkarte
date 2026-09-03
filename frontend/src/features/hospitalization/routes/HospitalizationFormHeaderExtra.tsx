import { FileText, Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { C, STYLE, ICON } from "@/lib/design-tokens";

export function HospitalizationFormHeaderExtra({
  hospitalizationId,
  canShowDelete,
  onOpenDetail,
  onOpenDeleteConfirm,
}: {
  hospitalizationId: string | undefined;
  canShowDelete: boolean;
  onOpenDetail: () => void;
  onOpenDeleteConfirm: () => void;
}) {
  return (
    <>
      {hospitalizationId ? (
        <Button
          variant="outline"
          type="button"
          className={`gap-2 h-10 text-sm px-4 ${C.text}`}
          onClick={onOpenDetail}
        >
          <FileText className={ICON.action} />
          デイリーカルテ
        </Button>
      ) : null}
      {hospitalizationId && canShowDelete ? (
        <Button
          variant="ghost"
          type="button"
          className={`${STYLE.btnDangerGhost} h-10 text-sm px-4`}
          onClick={onOpenDeleteConfirm}
        >
          <Trash2 className={`mr-1.5 ${ICON.action}`} />
          削除
        </Button>
      ) : null}
    </>
  );
}
