import { useCallback } from "react";
import { useLocation, useNavigate } from "react-router";
import { Scissors, Stethoscope } from "lucide-react";

import { Button } from "@/components/ui/button";
import { ICON } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { usePartnerRecordLink, type PartnerRecordKind } from "@/hooks/use-partner-record-link";

const PARTNER_LABELS: Record<PartnerRecordKind, { open: string; create: string }> = {
  "medical-record": { open: "診察カルテを開く", create: "診察カルテを作成" },
  trimming: { open: "トリミング記録を開く", create: "トリミング記録を作成" },
};

interface PartnerRecordLinkProps {
  /** 遷移先の record 種別。"medical-record" は診察カルテ、"trimming" はトリミング記録。 */
  kind: PartnerRecordKind;
  petId: string | undefined;
  /** 相方解決の基準日（YYYY-MM-DD）。不正値では query を走らせず非表示。 */
  visitDate: string;
  /** 死亡ペットでは相方 record への遷移導線自体を出さない（FE12 の表示側防壁）。 */
  isPetDeceased?: boolean;
}

/**
 * EMR-168 案A: 診察カルテ ⇔ トリミング記録の相互ショートカットボタン。
 * 同日・同一ペットの相方が既に存在すれば「開く」、なければ record_shortcut 入力経路へ「作成」。
 * 遷移先リソースの権限がない場合は描画しない（開く=view、作成=create）。
 */
export function PartnerRecordLink({
  kind,
  petId,
  visitDate,
  isPetDeceased = false,
}: PartnerRecordLinkProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { canView, canCreate } = usePermission(
    kind === "medical-record" ? "medical-records" : "trimming",
  );
  const { target } = usePartnerRecordLink({
    kind,
    petId,
    visitDate,
    enabled: !isPetDeceased,
  });

  const handleClick = useCallback(() => {
    if (!target) return;
    navigate(target.href, {
      state: {
        from: `${location.pathname}${location.search}`,
        appointmentId: target.appointmentId,
        visitDate,
      },
    });
  }, [target, location.pathname, location.search, navigate, visitDate]);

  if (isPetDeceased || !target) {
    return null;
  }
  const allowed = target.mode === "open" ? canView : canCreate;
  if (allowed !== true) {
    return null;
  }

  const Icon = kind === "medical-record" ? Stethoscope : Scissors;
  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      className="h-10 text-sm"
      onClick={handleClick}
    >
      <Icon className={`mr-1.5 ${ICON.action}`} />
      {PARTNER_LABELS[kind][target.mode]}
    </Button>
  );
}
