import { useState, useMemo } from "react";
import { Link, useLocation } from "react-router";
import {
  LayoutDashboard,
  Users,
  Calendar,
  FileText,
  TestTube,
  CreditCard,
  Bed,
  Syringe,
  Package,
  Scissors,
  Settings,
  ChevronDown,
  Pill,
  ShieldCheck,
  Building2,
  Stethoscope,
  Activity,
  PanelLeft,
} from "lucide-react";
import { useClinicInfo } from "@/hooks/use-clinic-info";
import { useSidebarStore } from "@/stores/sidebar-store";
import type { MenuItem } from "@/types";

interface SidebarItemProps {
  item: MenuItem;
  collapsed?: boolean;
  level?: number;
}

const SidebarItem = ({ item, collapsed = false, level = 0 }: SidebarItemProps) => {
  const location = useLocation();

  // Track manual toggle state (null = auto, true/false = user toggled)
  const [manualExpanded, setManualExpanded] = useState<boolean | null>(null);

  // Compute if any child is active
  const hasActiveChild = useMemo(
    () => item.subItems?.some(sub => sub.path === location.pathname) ?? false,
    [item.subItems, location.pathname]
  );

  // Auto-expand if child is active, unless manually collapsed
  const isExpanded = manualExpanded ?? hasActiveChild;

  const isActive = item.path === "/"
    ? location.pathname === "/"
    : location.pathname.startsWith(item.path || "");

  const hasSubItems = item.subItems && item.subItems.length > 0;

  const handleClick = (e: React.MouseEvent) => {
    if (hasSubItems) {
      e.preventDefault();
      setManualExpanded(!isExpanded);
    }
  };

  // Show active style if: directly active, or has subItems and is expanded
  const showActiveStyle = isActive || (hasSubItems && isExpanded);

  const content = (
    <div className={`w-full flex items-center h-[48px] rounded-[6px] text-[16px] leading-[24px] tracking-[0.0875px] transition-colors ${
        showActiveStyle
        ? "bg-[#eae9e5] text-[#37352f] font-medium"
        : "text-[rgba(55,53,47,0.7)] font-normal hover:bg-[#eae9e5]/50"
    } ${collapsed ? "justify-center px-0" : "gap-[12px] px-[12px]"} ${level > 0 && !collapsed ? "pl-8" : ""}`}>
        <div className={`size-4 shrink-0 ${level > 0 && !item.icon ? "invisible" : ""}`}>{item.icon}</div>
        {!collapsed && (
            <>
                <span className="truncate flex-1 text-left">{item.label}</span>
                {hasSubItems && (
                    <ChevronDown className={`size-3 transition-transform ${isExpanded ? "rotate-180" : ""}`} />
                )}
            </>
        )}
    </div>
  );

  return (
    <div className="w-full">
        {hasSubItems ? (
            <button onClick={handleClick} className="w-full block" title={collapsed ? item.label : undefined}>
                {content}
            </button>
        ) : (
            <Link to={item.path || "#"} className="w-full block" title={collapsed ? item.label : undefined}>
                {content}
            </Link>
        )}

        {hasSubItems && isExpanded && !collapsed && (
            <div className="space-y-0.5 mt-0.5 mb-1">
                {item.subItems?.map(sub => (
                    <SidebarItem
                        key={sub.label}
                        item={sub}
                        collapsed={collapsed}
                        level={level + 1}
                    />
                ))}
            </div>
        )}
    </div>
  );
};

export function Sidebar() {
  const { clinicInfo } = useClinicInfo();
  const { collapsed, toggle } = useSidebarStore();

  const menuItems: MenuItem[] = [
    { icon: <LayoutDashboard className="size-4" />, label: "当日の受付", path: "/" },
    { icon: <Users className="size-4" />, label: "飼主・ペット", path: "/owners" },
    { icon: <Calendar className="size-4" />, label: "予約管理", path: "/reservations" },
    { icon: <FileText className="size-4" />, label: "カルテ", path: "/medical-records" },
    { icon: <TestTube className="size-4" />, label: "検査管理", path: "/examinations" },
    { icon: <CreditCard className="size-4" />, label: "会計管理", path: "/accounting" },
    { icon: <Bed className="size-4" />, label: "入院・ホテル", path: "/hospitalization" },
    { icon: <Syringe className="size-4" />, label: "予防接種", path: "/vaccinations" },
    { icon: <Package className="size-4" />, label: "在庫管理", path: "/inventory" },
    { icon: <Scissors className="size-4" />, label: "トリミング", path: "/trimming" },
    {
      icon: <Settings className="size-4" />,
      label: "マスタ設定",
      subItems: [
        { icon: <Building2 className="size-4" />, label: "病院情報", path: "/settings/clinic" },
        { icon: <Activity className="size-4" />, label: "予約区分マスタ", path: "/settings/service-type" },
        { icon: <Stethoscope className="size-4" />, label: "診察マスタ", path: "/settings/consultation" },
        { icon: <TestTube className="size-4" />, label: "検査マスタ", path: "/settings/examination" },
        { icon: <Activity className="size-4" />, label: "処置マスタ", path: "/settings/procedure" },
        { icon: <Syringe className="size-4" />, label: "予防接種マスタ", path: "/settings/vaccine" },
        { icon: <Pill className="size-4" />, label: "薬剤マスタ", path: "/settings/medicine" },
        { icon: <Bed className="size-4" />, label: "入院マスタ", path: "/settings/hospitalization" },
        { icon: <Building2 className="size-4" />, label: "ケージマスタ", path: "/settings/cage" },
        { icon: <Users className="size-4" />, label: "スタッフマスタ", path: "/settings/staff" },
        { icon: <ShieldCheck className="size-4" />, label: "保険マスタ", path: "/settings/insurance" },
      ],
    },
  ];

  return (
    <aside className={`h-full border-r border-[rgba(55,53,47,0.16)] bg-[#f7f6f3] flex flex-col transition-all duration-200 ${collapsed ? "w-[56px]" : "w-[200px]"}`}>
      {/* Header: 48px height with bottom border */}
      <div className="h-[48px] border-b border-[rgba(55,53,47,0.16)] px-[12px] pb-px flex items-center">
        {collapsed ? (
          /* Collapsed: show only toggle button */
          <button
            onClick={toggle}
            className="size-10 flex items-center justify-center rounded-[6px] hover:bg-[#eae9e5] mx-auto"
            title="サイドバーを開く"
          >
            <PanelLeft className="size-5 text-[rgba(55,53,47,0.6)]" />
          </button>
        ) : (
          /* Expanded: show clinic info and toggle button */
          <div className="flex-1 flex items-center justify-between">
            <div className="flex flex-col">
              <span className="text-[14px] font-bold leading-[20px] text-[#37352f] tracking-[-0.1504px]">
                {clinicInfo.name}
              </span>
              <span className="text-[14px] font-normal leading-[20px] text-[rgba(55,53,47,0.6)] tracking-[-0.1504px]">
                {clinicInfo.branchName}
              </span>
            </div>
            <button
              onClick={toggle}
              className="size-10 flex items-center justify-center rounded-[6px] hover:bg-[#eae9e5]"
              title="サイドバーを閉じる"
            >
              <PanelLeft className="size-5 text-[rgba(55,53,47,0.6)]" />
            </button>
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 flex flex-col items-start gap-[2px] overflow-y-auto pt-[8px] px-[8px]">
        {menuItems.map((item) => (
          <SidebarItem key={item.label} item={item} collapsed={collapsed} />
        ))}
      </nav>
    </aside>
  );
}
