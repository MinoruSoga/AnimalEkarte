import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import { useState, useEffect, memo, useCallback } from "react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { useAuth } from "@/hooks/use-auth";
import { SidebarItemWithPermission } from "./SidebarItems";
import { sidebarMenuSections, type SidebarMenuSection } from "./sidebar-menu";
import { SidebarClinicHeader, SidebarUserFooter } from "./sidebar-chrome";

interface SidebarSectionProps {
  section: SidebarMenuSection;
  collapsed?: boolean;
}

function SidebarSection({ section, collapsed = false }: SidebarSectionProps) {
  return (
    <div className="space-y-px">
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

  const clinicName = user?.clinics.find((c) => c.clinicId === currentClinicId)?.clinicName
    ?? user?.clinic?.name
    ?? "";

  const hasMultipleClinics = (user?.clinics.length ?? 0) > 1;

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
      <SidebarClinicHeader
        collapsed={collapsed}
        clinicName={clinicName}
        hasMultipleClinics={hasMultipleClinics}
        clinicPopoverOpen={clinicPopoverOpen}
        onClinicPopoverOpenChange={setClinicPopoverOpen}
        user={user}
        currentClinicId={currentClinicId}
        onRequestSwitch={handleRequestSwitch}
        onCollapse={() => setCollapsed(true)}
        onExpand={() => setCollapsed(false)}
      />

      <nav aria-label="メインナビゲーション" className="flex-1 px-1 py-2 space-y-6 overflow-y-auto">
        {sidebarMenuSections.map((section) => (
          <SidebarSection key={section.title} section={section} collapsed={collapsed} />
        ))}
      </nav>

      <SidebarUserFooter
        collapsed={collapsed}
        displayName={user?.displayName ?? ""}
        isChangePasswordOpen={isChangePasswordOpen}
        onChangePasswordOpenChange={setIsChangePasswordOpen}
        onLogout={logout}
      />

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
