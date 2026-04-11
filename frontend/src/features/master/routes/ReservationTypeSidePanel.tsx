import { memo, useState, useCallback } from "react";
import { Activity, MessageCircle } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, PALETTE, STYLE } from "@/lib/design-tokens";
import type { ReservationType } from "@/features/master/api/reservation-types";

// ── 静的 SelectItem JSX (rendering-hoist-jsx) ──────────────────
export const RESERVATION_DAY_OPTION_ITEMS = (
  <>
    <SelectItem value="none">制限なし</SelectItem>
    <SelectItem value="weekday">平日のみ</SelectItem>
    <SelectItem value="saturday">土曜含む</SelectItem>
    <SelectItem value="anyday">毎日</SelectItem>
  </>
);

export interface CategoryFormData {
  name: string;
  description: string;
  isActive: boolean;
  groupId: string | undefined;
  reservationDisplayName: string;
  durationMinutes: number;
  shortName: string;
  reservationVisible: boolean;
  reservationComment: string;
  reservationImageUrl: string;
  showShortName: boolean;
  reservationDayOption: string;
  isInternal: boolean;
}

export interface GroupOption { id: string; name: string; color: string; }

export const CategorySidePanel = memo(function CategorySidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, groups, defaultGroupId,
}: {
  item: ReservationType | null;
  onClose: () => void;
  onSave: (d: CategoryFormData) => void;
  onDeleteRequest?: (i: ReservationType) => void;
  readOnly?: boolean;
  groups: GroupOption[];
  defaultGroupId?: string;
}) {
  const [f, setF] = useState<CategoryFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
    groupId: item?.groupId ?? defaultGroupId,
    reservationDisplayName: item?.reservationDisplayName ?? "",
    durationMinutes: item?.durationMinutes ?? 15,
    shortName: item?.shortName ?? "",
    reservationVisible: item?.reservationVisible ?? true,
    reservationComment: item?.reservationComment ?? "",
    reservationImageUrl: item?.reservationImageUrl ?? "",
    showShortName: item?.showShortName ?? false,
    reservationDayOption: item?.reservationDayOption ?? "none",
    isInternal: item?.isInternal ?? false,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleDescriptionChange = useCallback((v: string) => {
    setF((p) => ({ ...p, description: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!f.name.trim()) { setNameError("名称を入力してください"); return; }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleClose = useCallback(() => { setIsDirty(false); onClose(); }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Activity className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty} titleError={nameError} titleMaxLength={100} readOnly={readOnly}>
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="グループ">
        <Select
          value={f.groupId ?? "none"}
          onValueChange={(v) => { setF((p) => ({ ...p, groupId: v === "none" ? undefined : v })); setIsDirty(true); }}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="グループを選択" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="none">未分類</SelectItem>
            {groups.map((g) => (
              <SelectItem key={g.id} value={g.id}>
                <div className="flex items-center gap-2">
                  <span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: g.color }} />
                  {g.name}
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput value={f.description} onChange={handleDescriptionChange} placeholder="補足情報など" />
      </PropertyRow>

      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-3">
          <MessageCircle className="size-3.5" style={{ color: PALETTE.lineGreen }} />
          <p className={`text-xs font-medium ${C.text50}`}>LINE予約設定</p>
        </div>
        <PropertyRow label="LINE表示名">
          <PropertyInput value={f.reservationDisplayName}
            onChange={(v) => { setF((p) => ({ ...p, reservationDisplayName: v })); setIsDirty(true); }}
            placeholder={f.name || "空欄なら名称を使用"} />
        </PropertyRow>
        <PropertyRow label="予約ページに表示">
          <Switch checked={f.reservationVisible}
            onCheckedChange={(v) => { setF((p) => ({ ...p, reservationVisible: v })); setIsDirty(true); }} />
        </PropertyRow>
        <PropertyRow label="内部サービス">
          <Switch checked={f.isInternal}
            onCheckedChange={(v) => { setF((p) => ({ ...p, isInternal: v })); setIsDirty(true); }} />
        </PropertyRow>
        <PropertyRow label="所要時間（分）">
          <input type="number" min={5} max={480}
            className={`w-20 rounded-[3px] border ${C.borderMedium} px-2 py-1 text-base ${C.text}`}
            value={f.durationMinutes}
            onChange={(e) => { setF((p) => ({ ...p, durationMinutes: Number(e.target.value) || 15 })); setIsDirty(true); }} />
        </PropertyRow>
        <PropertyRow label="略称">
          <PropertyInput value={f.shortName}
            onChange={(v) => { setF((p) => ({ ...p, shortName: v })); setIsDirty(true); }}
            placeholder="LINE表示用の略称" />
        </PropertyRow>
        <PropertyRow label="略称を使用">
          <Switch checked={f.showShortName}
            onCheckedChange={(v) => { setF((p) => ({ ...p, showShortName: v })); setIsDirty(true); }} />
        </PropertyRow>
        <PropertyRow label="画像URL">
          <PropertyInput value={f.reservationImageUrl}
            onChange={(v) => { setF((p) => ({ ...p, reservationImageUrl: v })); setIsDirty(true); }}
            placeholder="https://..." />
        </PropertyRow>
        <PropertyRow label="予約可能曜日">
          <Select value={f.reservationDayOption}
            onValueChange={(v) => { setF((p) => ({ ...p, reservationDayOption: v })); setIsDirty(true); }}>
            <SelectTrigger className={STYLE.selectCompact}><SelectValue /></SelectTrigger>
            <SelectContent>{RESERVATION_DAY_OPTION_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>
        <PropertyRow label="LINE説明文">
          <PropertyInput value={f.reservationComment}
            onChange={(v) => { setF((p) => ({ ...p, reservationComment: v })); setIsDirty(true); }}
            placeholder="LINE予約画面に表示する説明" />
        </PropertyRow>
      </div>
    </MasterSidePanel>
  );
});
