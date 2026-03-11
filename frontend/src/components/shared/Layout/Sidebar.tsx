import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router";
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
  PanelLeftClose,
  Check,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useClinicInfo } from "@/hooks/use-clinic-info";
import type { MenuItem } from "@/types";

interface SidebarItemProps {
  item: MenuItem;
  collapsed?: boolean;
  level?: number;
}

const SidebarItem = ({ item, collapsed = false, level = 0 }: SidebarItemProps) => {
  const location = useLocation();
  const navigate = useNavigate();
  const [manualExpanded, setManualExpanded] = useState<boolean | null>(null);

  const isActive = item.path === "/"
    ? location.pathname === "/"
    : location.pathname.startsWith(item.path || "");

  const hasSubItems = item.subItems && item.subItems.length > 0;

  const hasActiveChild = hasSubItems && item.subItems
    ? item.subItems.some(sub => sub.path === location.pathname)
    : false;

  // Auto-expand if child is active, unless manually collapsed
  const isExpanded = manualExpanded ?? hasActiveChild;

  const handleClick = (e: React.MouseEvent) => {
    if (hasSubItems) {
      e.preventDefault();
      setManualExpanded(!isExpanded);
      if (item.path) {
        navigate(item.path);
      }
    }
  };

  const handleChevronClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setManualExpanded(!isExpanded);
  };

  const content = (
    <div className={`w-full flex items-center gap-3 px-3 py-3 rounded-md text-base transition-colors ${
        isActive
        ? "bg-[#EAE9E5] text-[#37352F] font-medium"
        : "text-[#37352F]/70 hover:bg-[#EAE9E5]/50 hover:text-[#37352F]"
    } ${collapsed ? "justify-center" : ""} ${level > 0 ? "pl-8" : ""}`}>
      <div className={`size-4 flex-shrink-0 ${level > 0 && !item.icon ? "invisible" : ""}`}>{item.icon}</div>
      {!collapsed && (
        <>
          <span className="truncate flex-1 text-left tracking-wide">{item.label}</span>
          {hasSubItems && (
            <span
              role="button"
              onClick={handleChevronClick}
              className="p-0.5 rounded hover:bg-[rgba(55,53,47,0.08)] transition-colors"
            >
              <ChevronDown className={`size-3 transition-transform ${isExpanded ? "rotate-180" : ""}`} />
            </span>
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
  const [collapsed, setCollapsed] = useState(false);
  const { clinicInfo } = useClinicInfo();

  const BRANCHES = ["八王子院", "新宿院", "渋谷院", "池袋院", "横浜院"];

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
      path: "/settings",
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
        { icon: <Scissors className="size-4" />, label: "トリミングコースマスタ", path: "/settings/trimming-course" },
        { icon: <Scissors className="size-4" />, label: "トリミングオプションマスタ", path: "/settings/trimming-option" },
        { icon: <Stethoscope className="size-4" />, label: "診断カテゴリマスタ", path: "/settings/diagnosis-category" },
        { icon: <Stethoscope className="size-4" />, label: "診断名マスタ", path: "/settings/diagnosis-name" },
      ],
    },
  ];

  return (
    <div className={`h-full bg-[#F7F6F3] border-r border-[rgba(55,53,47,0.16)] flex flex-col transition-all duration-300 ${collapsed ? "w-[50px]" : "w-[200px]"}`}>
      {/* Header */}
      <div className="h-12 flex items-center px-3 border-b border-[rgba(55,53,47,0.16)]">
        {!collapsed ? (
          <div className="flex items-center justify-between w-full">
            <div className="flex flex-col items-start">
              <span className="text-sm font-bold text-[#37352F] px-1">{clinicInfo.name}</span>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button className="flex items-center gap-1 text-sm text-[#37352F]/60 hover:text-[#37352F] transition-colors outline-none px-1 rounded-sm hover:bg-[rgba(55,53,47,0.06)]">
                    {clinicInfo.branchName}
                    <ChevronDown className="size-3 opacity-50" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-[160px]">
                  {BRANCHES.map((branch) => (
                    <DropdownMenuItem
                      key={branch}
                      className="justify-between"
                    >
                      {branch}
                      {clinicInfo.branchName === branch && <Check className="size-4 text-[#37352F]" />}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
            <button
              className="h-10 w-10 p-0 flex items-center justify-center text-[#37352F]/40 hover:text-[#37352F] hover:bg-transparent transition-colors"
              onClick={() => setCollapsed(!collapsed)}
            >
              <PanelLeftClose className="size-5" />
            </button>
          </div>
        ) : (
          <button
            className="w-full justify-center p-0 h-10 text-[#37352F]/60 hover:bg-[rgba(55,53,47,0.06)] hover:text-[#37352F] transition-colors flex items-center justify-center"
            onClick={() => setCollapsed(!collapsed)}
          >
            <PanelLeft className="size-5" />
          </button>
        )}
      </div>

      {/* Navigation Items */}
      <nav className="flex-1 p-2 space-y-0.5 overflow-y-auto">
        {menuItems.map((item) => (
          <SidebarItem
            key={item.label}
            item={item}
            collapsed={collapsed}
          />
        ))}
      </nav>
    </div>
  );
}
