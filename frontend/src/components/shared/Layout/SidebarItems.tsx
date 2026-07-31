import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { MenuItem } from "@/types";
import { ChevronDown } from "lucide-react";
import { memo, useState, type MouseEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { useAuth } from "@/hooks/use-auth";
import type { AuthContextValue } from "@/types/auth";

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

/** Parent resource is not a hard gate: show group when any descendant is viewable (LINE residual FINAL R-06/R-07). */
function isMenuItemVisible(
  item: MenuItem,
  hasPermission: AuthContextValue["hasPermission"],
): boolean {
  const selfOk =
    item.resource === undefined || hasPermission(item.resource, "view");
  if (item.subItems?.length) {
    const anyChild = item.subItems.some((sub) => isMenuItemVisible(sub, hasPermission));
    return selfOk || anyChild;
  }
  return selfOk;
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
    // Only navigate when path is present (callers strip path when unauthorized).
    if (item.path) navigate(item.path);
  };

  const handleChevronClick = (event: MouseEvent) => {
    event.preventDefault();
    event.stopPropagation();
    setManualExpanded(!isExpanded);
  };

  const contentBaseClassName = [
    // rounded-xxs: コードベース全体112箇所の既存compact-control標準値(全面改修は範囲外)
    "w-full flex items-center gap-3 h-12 rounded-xxs text-base transition-colors",
    isActive ? STYLE.sidebarItemActive : STYLE.sidebarItemIdle,
    collapsed ? "justify-center px-0" : "px-3",
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
            className={`${collapsed ? "min-w-11 justify-center" : "min-w-0"} flex-1 min-h-11 h-full flex items-center gap-3 text-left`}
            title={collapsed ? item.label : undefined}
            aria-label={collapsed ? item.label : undefined}
            aria-expanded={collapsed ? undefined : isExpanded}
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
              aria-label={isExpanded ? `${item.label}を折りたたむ` : `${item.label}を展開`}
              className={`min-h-11 min-w-11 flex items-center justify-center rounded ${C.hoverBgMedium} transition-colors`}
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
  const { hasPermission } = useAuth();

  if (!isMenuItemVisible(item, hasPermission)) {
    return null;
  }

  const selfOk =
    item.resource === undefined || hasPermission(item.resource, "view");
  // Parent failed its own resource but a child is visible: expand-only shell (no default navigate).
  const renderItem: MenuItem =
    !selfOk && item.subItems?.length
      ? { ...item, resource: undefined, path: undefined }
      : item;

  return <SidebarItem item={renderItem} collapsed={collapsed} level={level} />;
});
