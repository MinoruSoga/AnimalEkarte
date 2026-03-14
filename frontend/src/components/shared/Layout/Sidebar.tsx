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
  Building2,
  Activity,
  Package,
  CalendarDays,
  ClipboardCheck,
  Clipboard,
  ClipboardList,
  LogOut,
  User,
} from "lucide-react";
import { useState, useEffect, useMemo } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { useClinicInfo } from "@/hooks/use-clinic-info";
import { useAuth } from "@/features/auth/hooks/useAuth";
import type { MenuItem } from "@/types";

/* ================================================================== */
/*  SidebarItem                                                        */
/* ================================================================== */

interface SidebarItemProps {
  item: MenuItem;
  collapsed?: boolean;
  level?: number;
}

const SidebarItem = ({ item, collapsed = false, level = 0 }: SidebarItemProps) => {
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
        level > 0 ? "pl-8" : "",
      ].join(" ")}
    >
      <div className={`size-[18px] flex items-center justify-center shrink-0${level > 0 && !item.icon ? " invisible" : ""}`}>
        {item.icon}
      </div>
      {!collapsed && (
        <>
          <span className="truncate flex-1 text-left">{item.label}</span>
          {hasSubItems && (
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
          )}
        </>
      )}
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

      {hasSubItems && isExpanded && !collapsed && (
        <div className="space-y-0.5 mt-0.5 mb-1">
          {item.subItems?.map(sub => (
            <SidebarItem key={sub.label} item={sub} collapsed={collapsed} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  );
};

/* ================================================================== */
/*  Sidebar                                                            */
/* ================================================================== */

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(
    () => typeof window !== "undefined" && window.innerWidth < 1280,
  );
  const { clinicInfo } = useClinicInfo();
  const { user, logout } = useAuth();

  useEffect(() => {
    const mq = window.matchMedia("(max-width: 1279px)");
    const handler = (e: MediaQueryListEvent) => setCollapsed(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  const menuItems: MenuItem[] = useMemo(() => [
    { icon: <LayoutDashboard className="size-[18px]" />, label: "当日の受付",  path: "/" },
    { icon: <Users         className="size-[18px]" />, label: "飼主・ペット", path: "/owners" },
    { icon: <Calendar      className="size-[18px]" />, label: "予約管理",     path: "/reservations" },
    { icon: <FileText      className="size-[18px]" />, label: "カルテ",       path: "/medical-records" },
    { icon: <TestTube      className="size-[18px]" />, label: "検査管理",     path: "/examinations" },
    { icon: <CreditCard    className="size-[18px]" />, label: "会計管理",     path: "/accounting" },
    { icon: <Bed           className="size-[18px]" />, label: "入院・ホテル", path: "/hospitalization" },
    { icon: <Syringe       className="size-[18px]" />, label: "予防接種",     path: "/vaccinations" },
    { icon: <ClipboardCheck className="size-[18px]" />, label: "定期健診",    path: "/checkups" },
    { icon: <Package       className="size-[18px]" />, label: "在庫管理",     path: "/inventory" },
    { icon: <CalendarDays  className="size-[18px]" />, label: "シフト管理",   path: "/shifts" },
    { icon: <Scissors      className="size-[18px]" />, label: "トリミング",   path: "/trimming" },
    {
      icon: <Settings className="size-[18px]" />,
      label: "マスタ設定",
      path: "/settings",
      subItems: [
        { icon: <Activity      className="size-[18px]" />, label: "予約区分マスタ",   path: "/settings/service-type" },
        {
          icon: <FileText className="size-[18px]" />,
          label: "カルテ",
          subItems: [
            { icon: <ClipboardList className="size-[18px]" />, label: "治療プランマスタ", path: "/settings/treatment-items" },
            { icon: <Pill          className="size-[18px]" />, label: "薬剤マスタ",       path: "/settings/medicine" },
            { icon: <Clipboard     className="size-[18px]" />, label: "診断病名マスタ",   path: "/settings/diagnosis" },
            { icon: <ClipboardCheck className="size-[18px]" />, label: "問診マスタ",      path: "/settings/inquiry-template" },
          ],
        },
        { icon: <Bed          className="size-[18px]" />, label: "入院マスタ",       path: "/settings/hospitalization" },
        { icon: <Building2    className="size-[18px]" />, label: "ケージマスタ",     path: "/settings/cage" },
        { icon: <Scissors     className="size-[18px]" />, label: "トリミングマスタ", path: "/settings/trimming" },
        { icon: <Users        className="size-[18px]" />, label: "スタッフマスタ",   path: "/settings/staff" },
        { icon: <ShieldCheck  className="size-[18px]" />, label: "保険マスタ",       path: "/settings/insurance" },
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
              <span className="truncate">{clinicInfo.name}</span>
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
