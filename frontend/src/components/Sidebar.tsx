import { useState, useMemo } from "react";
import { Link, useLocation } from "react-router-dom";
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
  Stethoscope,
  Activity,
  Menu,
} from "lucide-react";
import { Button } from "./ui/button";
import { Sheet, SheetContent, SheetTrigger, SheetHeader, SheetTitle } from "./ui/sheet";
import { useClinicInfo } from "../features/clinic/hooks/useClinicInfo";
import { useDevice } from "./ui/use-mobile";
import type { MenuItem } from "../types";

interface SidebarItemProps {
  item: MenuItem;
  collapsed?: boolean;
  level?: number;
  onItemClick?: () => void; // ドロワーを閉じるためのコールバック
}

const SidebarItem = ({ item, collapsed = false, level = 0, onItemClick }: SidebarItemProps) => {
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

  const handleLinkClick = () => {
    // リンククリック時にドロワーを閉じる
    onItemClick?.();
  };

  // タッチ最適化: タップ領域を44px以上に拡大
  const content = (
    <div className={`w-full flex items-center gap-3 px-4 py-3.5 rounded-md text-base transition-colors min-h-[44px] ${
        isActive
        ? "bg-[#EAE9E5] text-[#37352F] font-medium"
        : "text-[#37352F]/70 hover:bg-[#EAE9E5]/50 hover:text-[#37352F]"
    } ${collapsed ? "justify-center" : ""} ${level > 0 ? "pl-10" : ""}`}>
        <div className={`size-5 flex-shrink-0 ${level > 0 && !item.icon ? "invisible" : ""}`}>{item.icon}</div>
        {!collapsed && (
            <>
                <span className="truncate flex-1 text-left tracking-wide">{item.label}</span>
                {hasSubItems && (
                    <ChevronDown className={`size-4 transition-transform ${isExpanded ? "rotate-180" : ""}`} />
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
            <Link
              to={item.path || "#"}
              className="w-full block"
              title={collapsed ? item.label : undefined}
              onClick={handleLinkClick}
            >
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
                        onItemClick={onItemClick}
                    />
                ))}
            </div>
        )}
    </div>
  );
};

// メニューアイテム定義（アイコンサイズを統一: size-5）
const createMenuItems = (): MenuItem[] => [
  { icon: <LayoutDashboard className="size-5" />, label: "当日の受付", path: "/" },
  { icon: <Users className="size-5" />, label: "飼主・ペット", path: "/owners" },
  { icon: <Calendar className="size-5" />, label: "予約管理", path: "/reservations" },
  { icon: <FileText className="size-5" />, label: "カルテ", path: "/medical-records" },
  { icon: <TestTube className="size-5" />, label: "検査管理", path: "/examinations" },
  { icon: <CreditCard className="size-5" />, label: "会計管理", path: "/accounting" },
  { icon: <Bed className="size-5" />, label: "入院・ホテル", path: "/hospitalization" },
  { icon: <Syringe className="size-5" />, label: "予防接種", path: "/vaccinations" },
  { icon: <Scissors className="size-5" />, label: "トリミング", path: "/trimming" },
  {
    icon: <Settings className="size-5" />,
    label: "マスタ設定",
    subItems: [
      { icon: <Building2 className="size-5" />, label: "病院情報", path: "/settings/clinic" },
      { icon: <Activity className="size-5" />, label: "予約区分マスタ", path: "/settings/service-type" },
      { icon: <Stethoscope className="size-5" />, label: "診察マスタ", path: "/settings/consultation" },
      { icon: <TestTube className="size-5" />, label: "検査マスタ", path: "/settings/examination" },
      { icon: <Activity className="size-5" />, label: "処置マスタ", path: "/settings/procedure" },
      { icon: <Syringe className="size-5" />, label: "予防接種マスタ", path: "/settings/vaccine" },
      { icon: <Pill className="size-5" />, label: "薬剤マスタ", path: "/settings/medicine" },
      { icon: <Bed className="size-5" />, label: "入院マスタ", path: "/settings/hospitalization" },
      { icon: <Building2 className="size-5" />, label: "ケージマスタ", path: "/settings/cage" },
      { icon: <Users className="size-5" />, label: "スタッフマスタ", path: "/settings/staff" },
      { icon: <ShieldCheck className="size-5" />, label: "保険マスタ", path: "/settings/insurance" },
    ]
  },
];

/**
 * ナビゲーションコンテンツ（デスクトップ/ドロワー共通）
 */
interface NavigationContentProps {
  menuItems: MenuItem[];
  collapsed?: boolean;
  onItemClick?: () => void;
}

const NavigationContent = ({ menuItems, collapsed = false, onItemClick }: NavigationContentProps) => (
  <nav className="flex-1 p-2 space-y-0.5 overflow-y-auto">
    {menuItems.map((item) => (
      <SidebarItem
        key={item.label}
        item={item}
        collapsed={collapsed}
        onItemClick={onItemClick}
      />
    ))}
  </nav>
);

/**
 * デスクトップ用固定サイドバー
 */
interface DesktopSidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  clinicInfo: { name: string; branchName: string };
  menuItems: MenuItem[];
}

const DesktopSidebar = ({ collapsed, onToggle, clinicInfo, menuItems }: DesktopSidebarProps) => (
  <div className={`h-full bg-[#F7F6F3] border-r border-[rgba(55,53,47,0.16)] flex flex-col transition-all duration-300 ${collapsed ? "w-[50px]" : "w-[220px]"}`}>
    {/* Header */}
    <div className="h-14 flex items-center px-3 border-b border-[rgba(55,53,47,0.16)]">
      {!collapsed ? (
        <div className="flex items-center justify-between w-full">
          <div className="flex flex-col">
            <span className="text-sm font-bold text-[#37352F]">{clinicInfo.name}</span>
            <span className="text-sm text-[#37352F]/60">{clinicInfo.branchName}</span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="size-11 p-0 text-[#37352F]/40 hover:text-[#37352F] hover:bg-transparent"
            onClick={onToggle}
          >
            <PanelLeftClose className="size-5" />
          </Button>
        </div>
      ) : (
        <Button
          variant="ghost"
          size="sm"
          className="w-full justify-center p-0 size-11 text-[#37352F]/60 hover:bg-[rgba(55,53,47,0.06)]"
          onClick={onToggle}
        >
          <PanelLeft className="size-5" />
        </Button>
      )}
    </div>

    <NavigationContent menuItems={menuItems} collapsed={collapsed} />
  </div>
);

/**
 * タブレット/モバイル用ドロワーサイドバー
 */
interface DrawerSidebarProps {
  clinicInfo: { name: string; branchName: string };
  menuItems: MenuItem[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const DrawerSidebar = ({ clinicInfo, menuItems, open, onOpenChange }: DrawerSidebarProps) => (
  <Sheet open={open} onOpenChange={onOpenChange}>
    <SheetTrigger asChild>
      <Button
        variant="ghost"
        size="sm"
        className="size-11 p-0 text-[#37352F]/70 hover:text-[#37352F] hover:bg-[rgba(55,53,47,0.06)]"
        aria-label="メニューを開く"
      >
        <Menu className="size-6" />
      </Button>
    </SheetTrigger>
    <SheetContent
      side="left"
      className="w-[280px] p-0 bg-[#F7F6F3] flex flex-col"
    >
      <SheetHeader className="h-14 px-4 border-b border-[rgba(55,53,47,0.16)] flex-row items-center justify-start gap-0">
        <SheetTitle className="flex flex-col items-start">
          <span className="text-sm font-bold text-[#37352F]">{clinicInfo.name}</span>
          <span className="text-sm text-[#37352F]/60 font-normal">{clinicInfo.branchName}</span>
        </SheetTitle>
      </SheetHeader>
      <NavigationContent
        menuItems={menuItems}
        onItemClick={() => onOpenChange(false)}
      />
    </SheetContent>
  </Sheet>
);

/**
 * サイドバー - デスクトップでは固定、タブレット/モバイルではドロワー
 */
export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { clinicInfo } = useClinicInfo();
  const { isDesktop } = useDevice();

  const menuItems = useMemo(() => createMenuItems(), []);

  // デスクトップ (>= 1024px): 固定サイドバー
  if (isDesktop) {
    return (
      <DesktopSidebar
        collapsed={collapsed}
        onToggle={() => setCollapsed(!collapsed)}
        clinicInfo={clinicInfo}
        menuItems={menuItems}
      />
    );
  }

  // タブレット/モバイル (< 1024px): ドロワー
  // ドロワートリガーはヘッダー内に配置するため、ここではnullを返す
  return null;
}

/**
 * ドロワー用のトリガーボタン（タブレット/モバイル用）
 * Layout コンポーネントで使用
 */
export function SidebarDrawerTrigger() {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { clinicInfo } = useClinicInfo();
  const { isDesktop } = useDevice();

  const menuItems = useMemo(() => createMenuItems(), []);

  // デスクトップでは表示しない
  if (isDesktop) {
    return null;
  }

  return (
    <DrawerSidebar
      clinicInfo={clinicInfo}
      menuItems={menuItems}
      open={drawerOpen}
      onOpenChange={setDrawerOpen}
    />
  );
}
