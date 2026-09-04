import { ChevronDown, PanelLeftClose, PanelLeft, KeyRound, LogOut, User } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { ChangePasswordDialog } from "@/components/shared/ChangePasswordDialog/ChangePasswordDialog";
import { C, ICON, STYLE, LAYOUT } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";
import type { AuthUser } from "@/types/auth";

interface SidebarClinicHeaderProps {
  collapsed: boolean;
  clinicName: string;
  hasMultipleClinics: boolean;
  clinicPopoverOpen: boolean;
  onClinicPopoverOpenChange: (open: boolean) => void;
  user: AuthUser | null;
  currentClinicId: string | null;
  onRequestSwitch: (clinicId: string) => void;
  onCollapse: () => void;
  onExpand: () => void;
}

export function SidebarClinicHeader({
  collapsed,
  clinicName,
  hasMultipleClinics,
  clinicPopoverOpen,
  onClinicPopoverOpenChange,
  user,
  currentClinicId,
  onRequestSwitch,
  onCollapse,
  onExpand,
}: SidebarClinicHeaderProps) {
  return (
    <div className={cn(STYLE.sidebarHeader, collapsed ? "px-1.5" : undefined)}>
      {!collapsed ? (
        <div className="flex items-center justify-between w-full">
          {hasMultipleClinics ? (
            <Popover open={clinicPopoverOpen} onOpenChange={onClinicPopoverOpenChange}>
              <PopoverTrigger asChild>
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
              <PopoverContent align="start" className="w-[200px] p-1">
                {user?.clinics.map((c) => {
                  const isCurrentClinic = c.clinicId === currentClinicId;
                  return (
                    <button
                      key={c.clinicId}
                      type="button"
                      onClick={() => onRequestSwitch(c.clinicId)}
                      aria-pressed={isCurrentClinic}
                      disabled={isCurrentClinic}
                      className={`flex min-h-11 min-w-11 w-full items-center gap-2 rounded-xxs px-2.5 py-1.5 text-left text-sm outline-none transition-colors focus-visible:ring-2 ${C.focusRingAccent40} ${
                        isCurrentClinic
                          ? `font-medium ${C.text} ${C.bgBrand10}`
                          : `${C.text65} ${C.hoverBgLight}`
                      }`}
                    >
                      <span
                        className={`${ICON.dotSm} rounded-full shrink-0 ${
                          isCurrentClinic ? C.bgBrand : C.bgInactive
                        }`}
                      />
                      {c.clinicName}
                    </button>
                  );
                })}
              </PopoverContent>
            </Popover>
          ) : (
            <div
              className={`flex items-center gap-1 min-w-0 text-base font-semibold px-1.5 py-1 ${C.text}`}
            >
              <span className={`${ICON.dot} rounded-full ${C.bgBrand} shrink-0`} />
              <span className="truncate">{clinicName}</span>
            </div>
          )}
          <button
            type="button"
            onClick={onCollapse}
            aria-label="サイドバーを折りたたむ"
            className={STYLE.sidebarToggle}
          >
            <PanelLeftClose className={ICON.page} />
          </button>
        </div>
      ) : (
        <button
          type="button"
          onClick={onExpand}
          aria-label="サイドバーを展開"
          className={`w-full min-w-11 flex items-center justify-center ${LAYOUT.sidebar.collapsedItemH} ${C.text50} ${C.hoverBgLight} rounded-xxs transition-colors`}
        >
          <PanelLeft className={ICON.page} />
        </button>
      )}
    </div>
  );
}

interface SidebarUserFooterProps {
  collapsed: boolean;
  displayName: string;
  isChangePasswordOpen: boolean;
  onChangePasswordOpenChange: (open: boolean) => void;
  onLogout: () => void;
}

export function SidebarUserFooter({
  collapsed,
  displayName,
  isChangePasswordOpen,
  onChangePasswordOpenChange,
  onLogout,
}: SidebarUserFooterProps) {
  return (
    <div className={`border-t ${C.borderDivider} px-1 py-1.5`}>
      {!collapsed ? (
        <>
          <div
            className={`flex items-center gap-2 px-2 py-1 rounded-xxs ${C.hoverBgLight} transition-colors`}
          >
            <div
              className={`${ICON.avatar} rounded-full flex items-center justify-center shrink-0 ${C.bgHoverMd}`}
            >
              <User className={`${ICON.avatarGlyph} ${C.text50}`} />
            </div>
            <p className={`flex-1 min-w-0 text-base ${C.text} truncate`}>{displayName}</p>
            <button
              type="button"
              onClick={() => onChangePasswordOpenChange(true)}
              aria-label="パスワード変更"
              title="パスワード変更"
              className={`${STYLE.iconBtn28} ${C.text35} ${C.hoverText} ${C.hoverBgMedium} shrink-0`}
            >
              <KeyRound className={ICON.action} />
            </button>
            <button
              type="button"
              onClick={onLogout}
              aria-label="ログアウト"
              title="ログアウト"
              className={`${STYLE.iconBtn28} ${C.text35} ${C.hoverText} ${C.hoverBgMedium} shrink-0`}
            >
              <LogOut className={ICON.action} />
            </button>
          </div>
          <ChangePasswordDialog
            open={isChangePasswordOpen}
            onOpenChange={onChangePasswordOpenChange}
            onSuccess={onLogout}
          />
        </>
      ) : (
        <button
          type="button"
          onClick={onLogout}
          aria-label="ログアウト"
          title="ログアウト"
          className={`w-full min-w-11 ${LAYOUT.sidebar.collapsedItemH} flex items-center justify-center rounded-xxs ${C.hoverBgMedium} transition-colors`}
        >
          <LogOut className={`${ICON.action} ${C.text40}`} />
        </button>
      )}
    </div>
  );
}
