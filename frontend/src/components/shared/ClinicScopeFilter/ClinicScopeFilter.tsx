import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import type { ClinicMembership } from "@/types/auth";

interface ClinicScopeFilterProps {
  /** ユーザー所属医院（選択肢はこれに限定。所属検証はサーバ側でも行う） */
  clinics: ClinicMembership[];
  /** 表示中の医院ID（最低1件） */
  selectedIds: string[];
  onToggle: (clinicId: string) => void;
}

/**
 * #86: 一覧画面の拠点横断表示フィルタ。
 * 所属医院をトグルボタンで複数選択する。単一所属ユーザーには表示しない。
 */
export function ClinicScopeFilter({ clinics, selectedIds, onToggle }: ClinicScopeFilterProps) {
  if (clinics.length < 2) return null;

  return (
    <div className="flex items-center gap-1.5 flex-wrap" data-testid="clinic-scope-filter">
      <span className={`text-sm ${C.text60}`}>表示拠点:</span>
      {clinics.map((membership) => {
        const selected = selectedIds.includes(membership.clinicId);
        return (
          <Button
            key={membership.clinicId}
            type="button"
            size="sm"
            variant={selected ? "default" : "outline"}
            aria-pressed={selected}
            className={
              selected
                ? `${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full transition-colors shadow-none border-transparent text-sm px-3`
                : `text-sm ${C.text} ${C.hoverBgMedium} ${C.borderMedium} px-3`
            }
            onClick={() => onToggle(membership.clinicId)}
          >
            {membership.clinicName}
          </Button>
        );
      })}
    </div>
  );
}
