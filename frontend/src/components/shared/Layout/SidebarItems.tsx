import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { MenuItem } from "@/types";
import { ChevronDown } from "lucide-react";
import { memo, useState, type MouseEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { usePermission } from "@/hooks/use-permission";
import type { Resource } from "@/types/generated/models";

interface SidebarItemProps {
  item: MenuItem;
  collapsed?: boolean;
  level?: number;
}

function checkAnyChildActive(items: MenuItem[], pathname: string): boolean {
  return items.some(
    (sub) =>
      (sub.path ? pathname.startsWith(sub.path) : false) ||
      (sub.subItems ? checkAnyChildActive(sub.subItems, pathname) : false),
  );
}

const SidebarItem = memo(function SidebarItem({ item, collapsed = false, level = 0 }: SidebarItemProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [manualExpanded, setManualExpanded] = useState<boolean | null>(null);

  const isActive = item.path
    ? item.path === "/"
      ? location.pathname === "/"
      : location.pathname.startsWith(item.path)
    : false;

  const hasSubItems = !!item.subItems?.length;
  const hasActiveChild = hasSubItems ? checkAnyChildActive(item.subItems ?? [], location.pathname) : false;
  const isExpanded = manualExpanded ?? hasActiveChild;

  const handleClick = (event: MouseEvent) => {
    if (!hasSubItems) return;

    event.preventDefault();
    setManualExpanded(!isExpanded);
    if (item.path) navigate(item.path);
  };

  const handleChevronClick = (event: MouseEvent) => {
    event.preventDefault();
    event.stopPropagation();
    setManualExpanded(!isExpanded);
  };

  const contentBaseClassName = [
    // rounded-[3px]: コードベース全体112箇所の既存compact-control標準値(全面改修は範囲外)
    "w-full flex items-center gap-3 px-3 h-12 rounded-[3px] text-base transition-colors",
    isActive ? STYLE.sidebarItemActive : STYLE.sidebarItemIdle,
    collapsed ? "justify-center" : "",
    level === 1 ? "pl-8" : level > 1 ? "pl-14" : "",
  ].join(" ");

  const content = (
    <div
      className={contentBaseClassName}
    >
      <div className={`${ICON.navItem} flex items-center justify-center shrink-0${level > 0 && !item.icon ? " invisible" : ""}`}>
        {item.icon}
      </div>
      {!collapsed ? <span className="truncate flex-1 text-left">{item.label}</span> : null}
    </div>
  );

  return (
    <div className="w-full">
      {hasSubItems ? (
        <div className={contentBaseClassName}>
          <button
            type="button"
            onClick={handleClick}
            className="min-w-0 flex-1 h-full flex items-center gap-3 text-left"
            title={collapsed ? item.label : undefined}
            aria-expanded={isExpanded}
          >
            <div className={`${ICON.navItem} flex items-center justify-center shrink-0${level > 0 && !item.icon ? " invisible" : ""}`}>
              {item.icon}
            </div>
            {!collapsed ? <span className="truncate flex-1">{item.label}</span> : null}
          </button>
          {!collapsed ? (
            <button
              type="button"
              onClick={handleChevronClick}
              aria-label={isExpanded ? `${item.label}を折りたむ` : `${item.label}を展開`}
              className={`p-0.5 rounded ${C.hoverBgMedium} transition-colors`}
            >
              <ChevronDown className={`${ICON.xs} transition-transform${isExpanded ? " rotate-180" : ""}`} />
            </button>
          ) : null}
        </div>
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
          {item.subItems?.map((sub) => (
            <SidebarItemWithPermission key={sub.label} item={sub} collapsed={collapsed} level={level + 1} />
          ))}
        </div>
      ) : null}
    </div>
  );
});

interface SidebarItemWithPermissionProps {
  item: MenuItem;
  collapsed?: boolean;
  level?: number;
}

export const SidebarItemWithPermission = memo(function SidebarItemWithPermission({
  item,
  collapsed = false,
  level = 0,
}: SidebarItemWithPermissionProps) {
  // FE6-2: item.resource は任意（権限不要メニュー項目もある）。React のフック呼び出し順序を
  // 崩さずに常に usePermission を呼ぶための sentinel。下の item.resource !== undefined ガードが
  // 実際の権限判定を制御するため、"" は Resource マップに存在しない安全なダミー値として扱われる。
  const { canView } = usePermission((item.resource ?? "") as Resource);
  if (item.resource !== undefined && !canView) return null;

  return <SidebarItem item={item} collapsed={collapsed} level={level} />;
});
