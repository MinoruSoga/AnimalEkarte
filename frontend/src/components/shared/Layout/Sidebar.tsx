import { C, ICON, STYLE } from "@/lib/design-tokens";
import { LayoutDashboard, Users, Calendar, FileText, TestTube, CreditCard, Bed, Syringe, Scissors, Settings, ChevronDown, PanelLeftClose, PanelLeft, Pill, ShieldCheck, Shield, Building2, Activity, Package, CalendarDays, ClipboardCheck, Clipboard, ClipboardList, KeyRound, LogOut, User, PawPrint } from "lucide-react";
import { useState, useEffect, memo } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { useAuth, ChangePasswordDialog, usePermission } from "@/features/auth";
import { paths } from "@/config/paths";
import { ResourceDashboard, ResourceOwners, ResourceReservations, ResourceMedicalRecords, ResourceExaminations, ResourceAccounting, ResourceHospitalization, ResourceVaccinations, ResourceCheckups, ResourceInventory, ResourceShifts, ResourceTrimming, ResourceMaster, ResourceHospitalSettings } from "@/types/generated/models";
import type { MenuItem } from "@/types";

/* ================================================================== */
/*  SidebarItem                                                        */
/* ================================================================== */

interface SidebarItemProps {
  item: MenuItem;
  collapsed?: boolean;
  level?: number;
}

const SidebarItem = memo(function SidebarItem({ item, collapsed = false, level = 0 }: SidebarItemProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [manualExpanded, setManualExpanded] = useState<boolean | null>(null);

  const isActive = item.path
    ? (item.path === "/" ? location.pathname === "/" : location.pathname.startsWith(item.path))
    : false;

  const hasSubItems = !!item.subItems?.length;

  const checkAnyChildActive = (items: MenuItem[]): boolean =>
    items.some(sub =>
      (sub.path ? location.pathname.startsWith(sub.path) : false) ||
      (sub.subItems ? checkAnyChildActive(sub.subItems) : false)
    );

  const hasActiveChild = hasSubItems
    ? checkAnyChildActive(item.subItems ?? [])
    : false;

  const isExpanded = manualExpanded ?? hasActiveChild;

  const handleClick = (e: React.MouseEvent) => {
    if (hasSubItems) {
      e.preventDefault();
      setManualExpanded(!isExpanded);
      if (item.path) navigate(item.path);
    }
  };

  const handleChevronClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setManualExpanded(!isExpanded);
  };

  const content = (
    <div
      className={[
        "w-full flex items-center gap-3 px-3 h-12 rounded-[3px] text-base transition-colors",
        isActive
          ? STYLE.sidebarItemActive
          : `${C.text65} ${C.hoverBgLight} ${C.hoverText}`,
        collapsed ? "justify-center" : "",
        level === 1 ? "pl-8" : level > 1 ? "pl-14" : "",
      ].join(" ")}
    >
      <div className={`size-[18px] flex items-center justify-center shrink-0${level > 0 && !item.icon ? " invisible" : ""}`}>
        {item.icon}
      </div>
      {!collapsed ? (
        <>
          <span className="truncate flex-1 text-left">{item.label}</span>
          {hasSubItems ? (
            <span
              role="button"
              tabIndex={0}
              onClick={handleChevronClick}
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === " ") {
                  e.preventDefault();
                  e.stopPropagation();
                  setManualExpanded(prev => !(prev ?? hasActiveChild));
                }
              }}
              aria-label={isExpanded ? `${item.label}を折りたむ` : `${item.label}を展開`}
              className={`p-0.5 rounded ${C.hoverBgMedium} transition-colors`}
            >
              <ChevronDown className={`${ICON.xs} transition-transform${isExpanded ? " rotate-180" : ""}`} />
            </span>
          ) : null}
        </>
      ) : null}
    </div>
  );

  return (
    <div className="w-full">
      {hasSubItems ? (
        <button
          type="button"
          onClick={handleClick}
          className="w-full block"
          title={collapsed ? item.label : undefined}
          aria-expanded={isExpanded}
        >
          {content}
        </button>
      ) : (
        <Link
          to={item.path || "#"}
          className="w-full block"
          title={collapsed ? item.label : undefined}
          aria-current={isActive ? "page" : undefined}
        >
          {content}
        </Link>
      )}

      {hasSubItems && isExpanded && !collapsed ? (
        <div className="space-y-0.5 mt-0.5 mb-1">
          {item.subItems?.map(sub => (
            <SidebarItemWithPermission key={sub.label} item={sub} collapsed={collapsed} level={level + 1} />
          ))}
        </div>
      ) : null}
    </div>
  );
});

/* ================================================================== */
/*  SidebarItemWithPermission — resource がある場合のみ権限チェック    */
/* ================================================================== */

interface SidebarItemWithPermissionProps {
  item: MenuItem;
  collapsed?: boolean;
  level?: number;
}

/**
 * メニュー項目を権限チェックしてレンダリングするコンポーネント。
 *
 * - `item.resource` が設定されている場合、`usePermission` で `canView` を確認し、
 *   `false` なら `null` を返す（非表示）。
 * - `item.resource` が未設定の場合（サブアイテムの一部等）は常に表示する。
 * - `clinic_admin` / `system_admin` は `useAuth.hasPermission` 内部で常に `true` を返す。
 *
 * フックのルール（ループ内での呼び出し禁止）を遵守するため、
 * 各メニュー項目をこのコンポーネントに切り出し、フック呼び出しをここで行う。
 */
const SidebarItemWithPermission = memo(function SidebarItemWithPermission({
  item,
  collapsed = false,
  level = 0,
}: SidebarItemWithPermissionProps) {
  // hooks のルール: 常に呼び出す（条件付き呼び出し禁止）。
  // resource が undefined の場合は空文字を渡し、下記の条件分岐で権限チェック結果を無視する。
  const { canView } = usePermission(item.resource ?? "");
  if (item.resource !== undefined && !canView) return null;
  return <SidebarItem item={item} collapsed={collapsed} level={level} />;
});

/* ================================================================== */
/*  Sidebar                                                            */
/* ================================================================== */

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(
    () => typeof window !== "undefined" && window.innerWidth < 1280,
  );
  const [isChangePasswordOpen, setIsChangePasswordOpen] = useState(false);
  const { user, logout } = useAuth();
  const clinicName = user?.clinic?.name ?? "";

  useEffect(() => {
    const mq = window.matchMedia("(max-width: 1279px)");
    const handler = (e: MediaQueryListEvent) => setCollapsed(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  return (
    <div
      className={`h-full ${C.bgPage} border-r ${C.borderLight} flex flex-col transition-all duration-300 ${
        collapsed ? "w-[56px]" : "w-[220px]"
      }`}
    >
      {/* Header */}
      <div className={`h-[53px] flex items-center px-2.5 border-b ${C.borderDivider}`}>
        {!collapsed ? (
          <div className="flex items-center justify-between w-full">
            <button
              type="button"
              className={`flex items-center gap-1 min-w-0 text-base font-semibold ${C.text} ${C.hoverBgLight} rounded-[3px] px-1.5 py-1 transition-colors outline-none`}
            >
              <span className={`size-2 rounded-full ${C.bgBrand} shrink-0`} />
              <span className="truncate">{clinicName}</span>
              <ChevronDown className={`${ICON.xs} opacity-40 shrink-0`} />
            </button>
            <button
              type="button"
              onClick={() => setCollapsed(true)}
              aria-label="サイドバーを折りたたむ"
              className={`size-7 flex items-center justify-center ${C.text40} ${C.hoverText} ${C.hoverBgMedium} rounded-[3px] transition-colors`}
            >
              <PanelLeftClose className={ICON.page} />
            </button>
          </div>
        ) : (
          <button
            type="button"
            onClick={() => setCollapsed(false)}
            aria-label="サイドバーを展開"
            className={`w-full flex items-center justify-center h-[30px] ${C.text50} ${C.hoverBgLight} rounded-[3px] transition-colors`}
          >
            <PanelLeft className={ICON.page} />
          </button>
        )}
      </div>

      {/* Navigation */}
      <nav aria-label="メインナビゲーション" className="flex-1 px-1 py-2 space-y-6 overflow-y-auto">
        {/* Clinical Section */}
        <div className="space-y-px">
          {!collapsed ? <p className={`px-3 mb-1 text-[10px] font-bold ${C.text40} uppercase tracking-wider`}>診療業務</p> : null}
          {[
            { icon: <LayoutDashboard className={ICON.toolbar} />, label: "当日の受付",  path: paths.home.getHref(),            resource: ResourceDashboard },
            { icon: <Users         className={ICON.toolbar} />, label: "飼主・ペット", path: paths.owners.getHref(),          resource: ResourceOwners },
            { icon: <Calendar      className={ICON.toolbar} />, label: "予約管理",     path: paths.reservations.getHref(),    resource: ResourceReservations },
            { icon: <FileText      className={ICON.toolbar} />, label: "カルテ",       path: paths.medicalRecords.getHref(),  resource: ResourceMedicalRecords },
            { icon: <TestTube      className={ICON.toolbar} />, label: "検査管理",     path: paths.examinations.getHref(),    resource: ResourceExaminations },
            { icon: <Scissors      className={ICON.toolbar} />, label: "トリミング",   path: paths.trimming.getHref(),        resource: ResourceTrimming },
            { icon: <Syringe       className={ICON.toolbar} />, label: "予防接種",     path: paths.vaccinations.getHref(),    resource: ResourceVaccinations },
            { icon: <ClipboardCheck className={ICON.toolbar} />, label: "定期健診",    path: "/checkups",                     resource: ResourceCheckups },
          ].map(item => (
            <SidebarItemWithPermission key={item.label} item={item as MenuItem} collapsed={collapsed} />
          ))}
        </div>

        {/* Operations Section */}
        <div className="space-y-px">
          {!collapsed ? <p className={`px-3 mb-1 text-[10px] font-bold ${C.text40} uppercase tracking-wider`}>運用・管理</p> : null}
          {[
            { icon: <CreditCard    className={ICON.toolbar} />, label: "会計管理",     path: paths.accounting.getHref(),      resource: ResourceAccounting },
            { icon: <Bed           className={ICON.toolbar} />, label: "入院・ホテル", path: paths.hospitalization.getHref(), resource: ResourceHospitalization },
            { icon: <Package       className={ICON.toolbar} />, label: "在庫管理",     path: paths.inventory.getHref(),       resource: ResourceInventory },
            { icon: <CalendarDays  className={ICON.toolbar} />, label: "シフト管理",   path: paths.shifts.getHref(),          resource: ResourceShifts },
          ].map(item => (
            <SidebarItemWithPermission key={item.label} item={item as MenuItem} collapsed={collapsed} />
          ))}
        </div>

        {/* Settings Section */}
        <div className="space-y-px">
          {!collapsed ? <p className={`px-3 mb-1 text-[10px] font-bold ${C.text40} uppercase tracking-wider`}>システム設定</p> : null}
          <SidebarItemWithPermission 
            item={{
              icon: <Settings className={ICON.toolbar} />,
              label: "マスタ設定",
              path: paths.settings.getHref(),
              resource: ResourceMaster,
              subItems: [
                { icon: <Building2    className={ICON.toolbar} />, label: "医院",       path: paths.settings.clinic.getHref(), resource: ResourceHospitalSettings },
                { icon: <PawPrint     className={ICON.toolbar} />, label: "動物種類",   path: paths.settings.animalSpecies.getHref() },
                {
                  icon: <FileText className={ICON.toolbar} />,
                  label: "カルテ関連",
                  subItems: [
                    { icon: <ClipboardList  className={ICON.toolbar} />, label: "診療項目", path: paths.settings.treatmentItems.getHref() },
                    { icon: <Clipboard      className={ICON.toolbar} />, label: "診断病名", path: paths.settings.diagnosis.getHref() },
                    { icon: <ClipboardCheck className={ICON.toolbar} />, label: "問診設定", path: paths.settings.inquiryTemplates.getHref() },
                    { icon: <Pill           className={ICON.toolbar} />, label: "薬剤マスタ", path: paths.settings.medicine.getHref() },
                  ],
                },
                { icon: <Activity     className={ICON.toolbar} />, label: "診療サービス", path: paths.settings.serviceType.getHref() },
                { icon: <Bed          className={ICON.toolbar} />, label: "入院・ケージ", path: paths.settings.hospitalization.getHref() },
                { icon: <Scissors     className={ICON.toolbar} />, label: "トリミング", path: paths.settings.trimming.getHref() },
                { icon: <Users        className={ICON.toolbar} />, label: "スタッフ管理", path: paths.settings.staff.getHref() },
                { icon: <Shield       className={ICON.toolbar} />, label: "権限グループ", path: paths.settings.permissionGroups.getHref() },
                { icon: <ShieldCheck  className={ICON.toolbar} />, label: "保険マスタ", path: paths.settings.insurance.getHref() },
                { icon: <Package      className={ICON.toolbar} />, label: "物販・フード", path: paths.settings.merchandiseItems.getHref() },
              ]
            } as MenuItem} 
            collapsed={collapsed} 
          />
        </div>
      </nav>

      {/* Footer */}
      <div className={`border-t ${C.borderDivider} px-1 py-1.5`}>
        {!collapsed ? (
          <>
            <div className={`flex items-center gap-2 px-2 py-1 rounded-[3px] ${C.hoverBgLight} transition-colors`}>
              <div className={`size-[26px] rounded-full flex items-center justify-center shrink-0 ${C.bgHoverMd}`}>
                <User className={`size-[13px] ${C.text50}`} />
              </div>
              <p className={`flex-1 min-w-0 text-base ${C.text} truncate`}>
                {user?.displayName ?? ""}
              </p>
              <button
                type="button"
                onClick={() => setIsChangePasswordOpen(true)}
                aria-label="パスワード変更"
                title="パスワード変更"
                className={`size-7 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} ${C.hoverBgMedium} transition-colors shrink-0`}
              >
                <KeyRound className={ICON.action} />
              </button>
              <button
                type="button"
                onClick={logout}
                aria-label="ログアウト"
                title="ログアウト"
                className={`size-7 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} ${C.hoverBgMedium} transition-colors shrink-0`}
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
          <div className="flex items-center justify-center h-[30px]">
            <LogOut className={`${ICON.action} ${C.text40}`} />
          </div>
        )}
      </div>
    </div>
  );
}
