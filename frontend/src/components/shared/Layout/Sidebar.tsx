import {
  LayoutDashboard,
  Users,
  Calendar,
  FileText,
  TestTube,
  CreditCard,
  Bed,
  Syringe,
  Scissors,
  Settings,
  ChevronDown,
  PanelLeftClose,
  PanelLeft,
  Pill,
  ShieldCheck,
  Shield,
  Building2,
  Activity,
  Package,
  CalendarDays,
  ClipboardCheck,
  Clipboard,
  ClipboardList,
  LogOut,
  User,
  PawPrint,
  Briefcase,
} from "lucide-react";
import { useState, useEffect, useMemo, memo } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { useAuth } from "@/features/auth/hooks/use-auth";
import { paths } from "@/config/paths";
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
          ? "bg-[#038B94]/8 text-[#37352F] border-l-2 border-l-[#038B94]"
          : "text-[#37352F]/65 hover:bg-[#37352F]/4 hover:text-[#37352F]",
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
              className="p-0.5 rounded hover:bg-[rgba(55,53,47,0.08)] transition-colors"
            >
              <ChevronDown className={`size-3 transition-transform${isExpanded ? " rotate-180" : ""}`} />
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
            <SidebarItem key={sub.label} item={sub} collapsed={collapsed} level={level + 1} />
          ))}
        </div>
      ) : null}
    </div>
  );
});

/* ================================================================== */
/*  Sidebar                                                            */
/* ================================================================== */

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(
    () => typeof window !== "undefined" && window.innerWidth < 1280,
  );
  const { user, logout } = useAuth();
  const clinicName = user?.clinic?.name ?? "";

  useEffect(() => {
    const mq = window.matchMedia("(max-width: 1279px)");
    const handler = (e: MediaQueryListEvent) => setCollapsed(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  const menuItems: MenuItem[] = useMemo(() => [
    { icon: <LayoutDashboard className="size-[18px]" />, label: "当日の受付",  path: paths.home.getHref() },
    { icon: <Users         className="size-[18px]" />, label: "飼主・ペット", path: paths.owners.getHref() },
    { icon: <Calendar      className="size-[18px]" />, label: "予約管理",     path: paths.reservations.getHref() },
    { icon: <FileText      className="size-[18px]" />, label: "カルテ",       path: paths.medicalRecords.getHref() },
    { icon: <TestTube      className="size-[18px]" />, label: "検査管理",     path: paths.examinations.getHref() },
    { icon: <CreditCard    className="size-[18px]" />, label: "会計管理",     path: paths.accounting.getHref() },
    { icon: <Bed           className="size-[18px]" />, label: "入院・ホテル", path: paths.hospitalization.getHref() },
    { icon: <Syringe       className="size-[18px]" />, label: "予防接種",     path: paths.vaccinations.getHref() },
    { icon: <ClipboardCheck className="size-[18px]" />, label: "定期健診",    path: "/checkups" },
    { icon: <Package       className="size-[18px]" />, label: "在庫管理",     path: paths.inventory.getHref() },
    { icon: <CalendarDays  className="size-[18px]" />, label: "シフト管理",   path: paths.shifts.getHref() },
    { icon: <Scissors      className="size-[18px]" />, label: "トリミング",   path: paths.trimming.getHref() },
    {
      icon: <Settings className="size-[18px]" />,
      label: "マスタ設定",
      path: paths.settings.getHref(),
      subItems: [
        // 基本設定
        { icon: <Building2    className="size-[18px]" />, label: "医院",       path: paths.settings.clinic.getHref() },
        { icon: <PawPrint     className="size-[18px]" />, label: "動物種類",   path: paths.settings.animalSpecies.getHref() },
        // カルテ
        {
          icon: <FileText className="size-[18px]" />,
          label: "カルテ",
          subItems: [
            { icon: <ClipboardList  className="size-[18px]" />, label: "診療項目", path: paths.settings.treatmentItems.getHref() },
            { icon: <Clipboard      className="size-[18px]" />, label: "診断病名", path: paths.settings.diagnosis.getHref() },
            { icon: <ClipboardCheck className="size-[18px]" />, label: "問診",     path: paths.settings.inquiryTemplates.getHref() },
            { icon: <Pill           className="size-[18px]" />, label: "薬剤",     path: paths.settings.medicine.getHref() },
          ],
        },
        // 診療関連
        { icon: <Activity     className="size-[18px]" />, label: "予約区分",   path: paths.settings.serviceType.getHref() },
        // 入院・ケージ管理
        { icon: <Bed          className="size-[18px]" />, label: "入院",       path: paths.settings.hospitalization.getHref() },
        { icon: <Building2    className="size-[18px]" />, label: "ケージ",     path: paths.settings.cage.getHref() },
        // トリミング
        { icon: <Scissors     className="size-[18px]" />, label: "トリミング", path: paths.settings.trimming.getHref() },
        // スタッフ・保険
        { icon: <Users        className="size-[18px]" />, label: "スタッフ",   path: paths.settings.staff.getHref() },
        { icon: <Briefcase    className="size-[18px]" />, label: "職種",       path: paths.settings.jobTitle.getHref() },
        { icon: <ShieldCheck  className="size-[18px]" />, label: "保険",       path: paths.settings.insurance.getHref() },
        { icon: <Package      className="size-[18px]" />, label: "物販",       path: paths.settings.merchandiseItems.getHref() },
        // 権限
        { icon: <Shield       className="size-[18px]" />, label: "権限グループ", path: paths.settings.permissionGroups.getHref() },
      ],
    },
  ], []);

  return (
    <div
      className={`h-full bg-[#F7F6F3] border-r border-[rgba(55,53,47,0.09)] flex flex-col transition-all duration-300 ${
        collapsed ? "w-[56px]" : "w-[220px]"
      }`}
    >
      {/* Header */}
      <div className="h-[53px] flex items-center px-2.5 border-b border-[rgba(55,53,47,0.06)]">
        {!collapsed ? (
          <div className="flex items-center justify-between w-full">
            <button
              type="button"
              className="flex items-center gap-1 min-w-0 text-base font-semibold text-[#37352F] hover:bg-[rgba(55,53,47,0.04)] rounded-[3px] px-1.5 py-1 transition-colors outline-none"
            >
              <span className="size-2 rounded-full bg-[#038B94] shrink-0" />
              <span className="truncate">{clinicName}</span>
              <ChevronDown className="size-3 opacity-40 shrink-0" />
            </button>
            <button
              type="button"
              onClick={() => setCollapsed(true)}
              aria-label="サイドバーを折りたたむ"
              className="size-7 flex items-center justify-center text-[#37352F]/40 hover:text-[#37352F] hover:bg-[rgba(55,53,47,0.08)] rounded-[3px] transition-colors"
            >
              <PanelLeftClose className="size-[18px]" />
            </button>
          </div>
        ) : (
          <button
            type="button"
            onClick={() => setCollapsed(false)}
            aria-label="サイドバーを展開"
            className="w-full flex items-center justify-center h-[30px] text-[#37352F]/50 hover:bg-[rgba(55,53,47,0.04)] rounded-[3px] transition-colors"
          >
            <PanelLeft className="size-[18px]" />
          </button>
        )}
      </div>

      {/* Navigation */}
      <nav aria-label="メインナビゲーション" className="flex-1 px-1 py-2 space-y-px overflow-y-auto">
        {menuItems.map(item => (
          <SidebarItem key={item.label} item={item} collapsed={collapsed} />
        ))}
      </nav>

      {/* Footer */}
      <div className="border-t border-[rgba(55,53,47,0.06)] px-1 py-1.5">
        {!collapsed ? (
          <div className="flex items-center gap-2 px-2 py-1 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] transition-colors">
            <div className="size-[26px] rounded-full flex items-center justify-center shrink-0 bg-[rgba(55,53,47,0.08)]">
              <User className="size-[13px] text-[#37352F]/50" />
            </div>
            <p className="flex-1 min-w-0 text-base text-[#37352F] truncate">
              {user?.displayName ?? ""}
            </p>
            <button
              type="button"
              onClick={logout}
              aria-label="ログアウト"
              title="ログアウト"
              className="size-[26px] flex items-center justify-center rounded-[3px] text-[#37352F]/35 hover:text-[#37352F] hover:bg-[rgba(55,53,47,0.08)] transition-colors shrink-0"
            >
              <LogOut className="size-[15px]" />
            </button>
          </div>
        ) : (
          <div className="flex items-center justify-center h-[30px]">
            <LogOut className="size-[15px] text-[#37352F]/40" />
          </div>
        )}
      </div>
    </div>
  );
}
