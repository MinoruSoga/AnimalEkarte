import { C, ICON, STYLE, LAYOUT } from "@/lib/design-tokens";
import { ChevronDown, PanelLeftClose, PanelLeft, KeyRound, LogOut, User } from "lucide-react";
import { useState, useEffect, memo, useCallback } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { ChangePasswordDialog } from "@/components/shared/ChangePasswordDialog/ChangePasswordDialog";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { useAuth } from "@/hooks/use-auth";
import { cn } from "@/lib/utils";
import { SidebarItemWithPermission } from "./SidebarItems";
import { sidebarMenuSections, type SidebarMenuSection } from "./sidebar-menu";

interface SidebarSectionProps {
  section: SidebarMenuSection;
  collapsed?: boolean;
}

function SidebarSection({ section, collapsed = false }: SidebarSectionProps) {
  return (
    <div className="space-y-px">
      {/* FE9-2: design-system.md §3.4 micro ロールへ全面統一。旧 10px 任意値は廃止。 */}
      {!collapsed ? (
        <p className={`px-3 mb-1 text-2xs font-semibold ${C.text40} uppercase`}>
          {section.title}
        </p>
      ) : null}
      {section.items.map((item) => (
        <SidebarItemWithPermission key={item.label} item={item} collapsed={collapsed} />
      ))}
    </div>
  );
}

export const Sidebar = memo(function Sidebar() {
  const [collapsed, setCollapsed] = useState(
    () => typeof window !== "undefined" && window.innerWidth < 1280,
  );
  const [isChangePasswordOpen, setIsChangePasswordOpen] = useState(false);
  const { user, logout, switchClinic, currentClinicId } = useAuth();
  const [clinicPopoverOpen, setClinicPopoverOpen] = useState(false);
  const [pendingSwitchClinicId, setPendingSwitchClinicId] = useState<string | null>(null);

  // 現在のクリニック名: currentClinicId に一致する clinics から取得
  const clinicName = user?.clinics.find((c) => c.clinicId === currentClinicId)?.clinicName
    ?? user?.clinic?.name
    ?? "";

  // 所属医院が1つの場合は選択UIを無効化
  const hasMultipleClinics = (user?.clinics.length ?? 0) > 1;

  // 切替先のクリニック名
  const pendingSwitchClinicName = pendingSwitchClinicId
    ? user?.clinics.find((c) => c.clinicId === pendingSwitchClinicId)?.clinicName ?? ""
    : "";

  const handleRequestSwitch = useCallback((clinicId: string) => {
    setClinicPopoverOpen(false);
    setPendingSwitchClinicId(clinicId);
  }, []);

  const handleConfirmSwitch = useCallback(() => {
    if (pendingSwitchClinicId) {
      switchClinic(pendingSwitchClinicId);
    }
    setPendingSwitchClinicId(null);
  }, [pendingSwitchClinicId, switchClinic]);

  const handleCancelSwitch = useCallback(() => {
    setPendingSwitchClinicId(null);
  }, []);

  useEffect(() => {
    const mq = window.matchMedia("(max-width: 1279px)");
    const handler = (e: MediaQueryListEvent) => setCollapsed(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  return (
    <div
      className={`${STYLE.sidebarContainer} ${collapsed ? LAYOUT.sidebar.collapsed : LAYOUT.sidebar.expanded}`}
    >
      {/* Header */}
      <div className={cn(STYLE.sidebarHeader, collapsed ? "px-1.5" : undefined)}>
        {!collapsed ? (
          <div className="flex items-center justify-between w-full">
            {hasMultipleClinics ? (
              <Popover open={clinicPopoverOpen} onOpenChange={setClinicPopoverOpen}>
                <PopoverTrigger asChild>
                  {/* rounded-xxs: コードベース全体112箇所の既存compact-control標準値。以降の同値も同様(全面改修は範囲外) */}
                  <button
                    type="button"
                    aria-label={clinicName || "医院を切り替える"}
                    className={`flex min-h-11 min-w-11 max-w-full items-center gap-1 text-base font-semibold ${C.text} ${C.hoverBgLight} rounded-xxs px-1.5 py-1 transition-colors outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
                  >
                    <span className={`${ICON.dot} rounded-full ${C.bgBrand} shrink-0`} />
                    <span className="truncate">{clinicName}</span>
                    <ChevronDown className={`${ICON.xs} opacity-40 shrink-0`} />
                  </button>
                </PopoverTrigger>
                {/* w-[200px]: 15ファイルで使われる既存の慣習値。Sidebar単体でのtoken化は他箇所との不整合を招くため範囲外 */}
                <PopoverContent align="start" className="w-[200px] p-1">
                  {user?.clinics.map((c) => {
                    const isCurrentClinic = c.clinicId === currentClinicId;
                    return (
                      <button
                        key={c.clinicId}
                        type="button"
                        onClick={() => handleRequestSwitch(c.clinicId)}
                        aria-pressed={isCurrentClinic}
                        disabled={isCurrentClinic}
                        className={`flex min-h-11 min-w-11 w-full items-center gap-2 rounded-xxs px-2.5 py-1.5 text-left text-sm outline-none transition-colors focus-visible:ring-2 ${C.focusRingAccent40} ${
                          isCurrentClinic
                            ? `font-medium ${C.text} ${C.bgBrand10}`
                            : `${C.text65} ${C.hoverBgLight}`
                        }`}
                      >
                        <span className={`${ICON.dotSm} rounded-full shrink-0 ${
                          isCurrentClinic ? C.bgBrand : C.bgInactive
                        }`} />
                        {c.clinicName}
                      </button>
                    );
                  })}
                </PopoverContent>
              </Popover>
            ) : (
              <div className={`flex items-center gap-1 min-w-0 text-base font-semibold px-1.5 py-1 ${C.text}`}>
                <span className={`${ICON.dot} rounded-full ${C.bgBrand} shrink-0`} />
                <span className="truncate">{clinicName}</span>
              </div>
            )}
            <button
              type="button"
              onClick={() => setCollapsed(true)}
              aria-label="サイドバーを折りたたむ"
              className={STYLE.sidebarToggle}
            >
              <PanelLeftClose className={ICON.page} />
            </button>
          </div>
        ) : (
          <button
            type="button"
            onClick={() => setCollapsed(false)}
            aria-label="サイドバーを展開"
            // rounded-xxs: コードベース全体112箇所の既存compact-control標準値(全面改修は範囲外)
            className={`w-full min-w-11 flex items-center justify-center ${LAYOUT.sidebar.collapsedItemH} ${C.text50} ${C.hoverBgLight} rounded-xxs transition-colors`}
          >
            <PanelLeft className={ICON.page} />
          </button>
        )}
      </div>

      {/* Navigation */}
      <nav aria-label="メインナビゲーション" className="flex-1 px-1 py-2 space-y-6 overflow-y-auto">
        {sidebarMenuSections.map((section) => (
          <SidebarSection key={section.title} section={section} collapsed={collapsed} />
        ))}
      </nav>

      {/* Footer */}
      <div className={`border-t ${C.borderDivider} px-1 py-1.5`}>
        {!collapsed ? (
          <>
            <div className={`flex items-center gap-2 px-2 py-1 rounded-xxs ${C.hoverBgLight} transition-colors`}>
              <div className={`${ICON.avatar} rounded-full flex items-center justify-center shrink-0 ${C.bgHoverMd}`}>
                <User className={`${ICON.avatarGlyph} ${C.text50}`} />
              </div>
              <p className={`flex-1 min-w-0 text-base ${C.text} truncate`}>
                {user?.displayName ?? ""}
              </p>
              <button
                type="button"
                onClick={() => setIsChangePasswordOpen(true)}
                aria-label="パスワード変更"
                title="パスワード変更"
                className={`${STYLE.iconBtn28} ${C.text35} ${C.hoverText} ${C.hoverBgMedium} shrink-0`}
              >
                <KeyRound className={ICON.action} />
              </button>
              <button
                type="button"
                onClick={logout}
                aria-label="ログアウト"
                title="ログアウト"
                className={`${STYLE.iconBtn28} ${C.text35} ${C.hoverText} ${C.hoverBgMedium} shrink-0`}
              >
                <LogOut className={ICON.action} />
              </button>
            </div>
            <ChangePasswordDialog
              open={isChangePasswordOpen}
              onOpenChange={setIsChangePasswordOpen}
              onSuccess={logout}
            />
          </>
        ) : (
          <button
            type="button"
            onClick={logout}
            aria-label="ログアウト"
            title="ログアウト"
            className={`w-full min-w-11 ${LAYOUT.sidebar.collapsedItemH} flex items-center justify-center rounded-xxs ${C.hoverBgMedium} transition-colors`}
          >
            <LogOut className={`${ICON.action} ${C.text40}`} />
          </button>
        )}
      </div>

      {/* クリニック切替確認ダイアログ */}
      <ConfirmDialog
        open={pendingSwitchClinicId !== null}
        onClose={handleCancelSwitch}
        onConfirm={handleConfirmSwitch}
        title="医院を切り替えますか？"
        description={`「${pendingSwitchClinicName}」に切り替えます。現在の画面データは新しい医院のデータに更新されます。`}
        confirmLabel="切り替える"
        cancelLabel="キャンセル"
      />
    </div>
  );
});
