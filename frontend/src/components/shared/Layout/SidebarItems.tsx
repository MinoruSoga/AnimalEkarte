import { C, ICON, STYLE, Z } from "@/lib/design-tokens";
import type { MenuItem } from "@/types";
import { ChevronDown } from "lucide-react";
import {
  memo,
  useEffect,
  useId,
  useLayoutEffect,
  useRef,
  useState,
  type FocusEvent,
  type MouseEvent,
} from "react";
import { createPortal } from "react-dom";
import { Link, useLocation, useNavigate } from "react-router";
import { useAuth } from "@/hooks/use-auth";
import type { AuthContextValue } from "@/types/auth";

import { sidebarItemDescriptions } from "./SidebarMenu";

interface SidebarItemProps {
  item: MenuItem;
  /** Path used for the purpose sentence. Survives when navigation path is stripped. */
  purposePath?: string;
  collapsed?: boolean;
  level?: number;
}

const TOOLTIP_GAP_PX = 8;
const VIEWPORT_EDGE_PX = 8;

function tooltipBox(
  anchorRight: number,
  viewportWidth: number,
): { left: number; maxWidth: string } {
  const left = anchorRight + TOOLTIP_GAP_PX;
  const remaining = Math.max(0, viewportWidth - left - TOOLTIP_GAP_PX);
  return { left, maxWidth: `min(28rem, ${remaining}px)` };
}

function tooltipTop(anchorTop: number, contentHeight: number, viewportHeight: number): number {
  const available = Math.max(0, viewportHeight - VIEWPORT_EDGE_PX * 2);
  const height = Math.min(Math.max(contentHeight, 0), available);
  const maxTop = VIEWPORT_EDGE_PX + Math.max(0, available - height);
  return Math.min(Math.max(anchorTop, VIEWPORT_EDGE_PX), maxTop);
}

function targetInPurposeTooltip(target: EventTarget | null): boolean {
  return target instanceof Element && target.closest("[data-purpose-tooltip]") !== null;
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
  const selfOk = item.resource === undefined || hasPermission(item.resource, "view");
  if (item.subItems?.length) {
    const anyChild = item.subItems.some((sub) => isMenuItemVisible(sub, hasPermission));
    return selfOk || anyChild;
  }
  return selfOk;
}

function purposeForKey(key: string | undefined): string | undefined {
  if (key === undefined) return undefined;
  return sidebarItemDescriptions[key];
}

function owningSidebarRow(node: Element): Element | null {
  return node.closest("[data-sidebar-row]");
}

function eventTargetInRow(target: EventTarget | null, row: Element): boolean {
  return target instanceof Element && owningSidebarRow(target) === row;
}

function controlsOwnedByRow(row: HTMLElement): Element[] {
  return [...row.querySelectorAll("a, button")].filter(
    (control) => owningSidebarRow(control) === row,
  );
}

function PurposeTooltip({
  id,
  text,
  exposure,
  anchor,
  onPointerEnter,
  onPointerLeave,
}: {
  id: string;
  text: string;
  exposure: string;
  anchor: HTMLElement | null;
  onPointerEnter: () => void;
  onPointerLeave: (next: EventTarget | null) => void;
}) {
  const rect = anchor?.getBoundingClientRect() ?? null;
  const box = tooltipBox(rect?.right ?? 0, window.innerWidth);
  const [node, setNode] = useState<HTMLDivElement | null>(null);
  const [contentHeight, setContentHeight] = useState(0);
  useLayoutEffect(() => {
    if (!node) return;
    const next = node.scrollHeight;
    if (next !== contentHeight) {
      // Height is known only after layout, and the position must update before paint.
      // eslint-disable-next-line react-hooks/set-state-in-effect -- measure tooltip before paint
      setContentHeight(next);
    }
  }, [contentHeight, node, text, box.maxWidth]);
  const available = Math.max(0, window.innerHeight - VIEWPORT_EDGE_PX * 2);
  const top = tooltipTop(rect?.top ?? 0, contentHeight, window.innerHeight);
  const bridgeHeight = Math.max(rect?.height ?? 0, contentHeight, 1);
  const leave = (event: MouseEvent<HTMLDivElement>) => {
    onPointerLeave(event.relatedTarget);
  };
  return createPortal(
    <>
      <div
        data-purpose-tooltip=""
        aria-hidden="true"
        onMouseEnter={onPointerEnter}
        onMouseLeave={leave}
        style={{
          position: "fixed",
          top,
          left: rect?.right ?? 0,
          width: TOOLTIP_GAP_PX,
          height: bridgeHeight,
          zIndex: Z.overlay,
        }}
      />
      <div
        ref={setNode}
        id={id}
        role="tooltip"
        data-purpose-tooltip=""
        data-exposure={exposure}
        onMouseEnter={onPointerEnter}
        onMouseLeave={leave}
        style={{
          position: "fixed",
          top,
          left: box.left,
          zIndex: Z.overlay,
          maxWidth: box.maxWidth,
          maxHeight: `${available}px`,
          overflow: "auto",
        }}
        className={`w-max rounded ${C.bgTooltip} px-2.5 py-1.5 text-xs leading-relaxed ${C.textWhite}`}
      >
        {text}
      </div>
    </>,
    document.body,
  );
}

const SidebarItem = memo(function SidebarItem({
  item,
  purposePath,
  collapsed = false,
  level = 0,
}: SidebarItemProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [manualExpanded, setManualExpanded] = useState<boolean | null>(null);

  const isActive = item.path
    ? item.path === "/"
      ? location.pathname === "/"
      : location.pathname.startsWith(item.path)
    : false;

  const hasSubItems = !!item.subItems?.length;
  const hasActiveChild = hasSubItems
    ? checkAnyChildActive(item.subItems ?? [], location.pathname)
    : false;
  const isExpanded = manualExpanded ?? hasActiveChild;
  const description = purposeForKey(item.purposeKey ?? purposePath ?? item.path);
  const descriptionId = useId();
  const [rowElement, setRowElement] = useState<HTMLDivElement | null>(null);
  const keyboardIntent = useRef(false);
  const hoverRef = useRef(false);
  const focusRef = useRef(false);
  const [hovered, setHovered] = useState(false);
  const [keyboardFocused, setKeyboardFocused] = useState(false);
  const [dismissed, setDismissed] = useState(false);
  const descriptionOpen = description !== undefined && !dismissed && (hovered || keyboardFocused);
  const exposure = [hovered ? "hover" : "", keyboardFocused ? "focus-visible" : ""]
    .filter((part) => part !== "")
    .join(" ");
  const describedBy = descriptionOpen ? descriptionId : undefined;

  useEffect(() => {
    const row = rowElement;
    if (!row || description === undefined) return;
    const controls = controlsOwnedByRow(row);
    // jsdom fires mouseleave in the same turn as mouseenter. Ignore only that pair.
    let ignorePairedLeave = false;
    const relatedTargetOf = (event: Event): EventTarget | null =>
      "relatedTarget" in event
        ? (event as Event & { relatedTarget: EventTarget | null }).relatedTarget
        : null;
    const endHover = () => {
      hoverRef.current = false;
      setHovered(false);
      if (!focusRef.current) setDismissed(false);
    };
    const pointerStillOnRow = (next: EventTarget | null) =>
      targetInPurposeTooltip(next) || eventTargetInRow(next, row);
    const enter = () => {
      ignorePairedLeave = true;
      hoverRef.current = true;
      setHovered(true);
      queueMicrotask(() => {
        ignorePairedLeave = false;
      });
    };
    const leaveUnlessStillHere = (event: Event) => {
      if (ignorePairedLeave) return;
      if (pointerStillOnRow(relatedTargetOf(event))) return;
      endHover();
    };
    // Gaps in the row are not controls. Leaving the row, or entering a nested row, ends hover.
    const leaveForNestedRow = (event: Event) => {
      if (ignorePairedLeave || !hoverRef.current) return;
      const raw = event.target;
      const target = raw instanceof Element ? raw : raw instanceof Node ? raw.parentElement : null;
      if (target === null || owningSidebarRow(target) === row) return;
      endHover();
    };
    const onFocus = (event: Event) => {
      const target = event.target;
      if (!(target instanceof Element) || owningSidebarRow(target) !== row) return;
      if (target.matches(":focus-visible") || keyboardIntent.current) {
        focusRef.current = true;
        setKeyboardFocused(true);
      }
    };
    const onBlur = (event: Event) => {
      const next = event instanceof FocusEvent ? event.relatedTarget : null;
      if (eventTargetInRow(next, row)) return;
      focusRef.current = false;
      setKeyboardFocused(false);
      if (!hoverRef.current) setDismissed(false);
    };
    for (const control of controls) {
      control.addEventListener("mouseenter", enter);
      control.addEventListener("mouseleave", leaveUnlessStillHere);
      control.addEventListener("focus", onFocus);
      control.addEventListener("blur", onBlur);
    }
    row.addEventListener("mouseleave", leaveUnlessStillHere);
    row.addEventListener("mouseover", leaveForNestedRow);
    return () => {
      for (const control of controls) {
        control.removeEventListener("mouseenter", enter);
        control.removeEventListener("mouseleave", leaveUnlessStillHere);
        control.removeEventListener("focus", onFocus);
        control.removeEventListener("blur", onBlur);
      }
      row.removeEventListener("mouseleave", leaveUnlessStillHere);
      row.removeEventListener("mouseover", leaveForNestedRow);
    };
  }, [description, collapsed, isExpanded, hasSubItems, rowElement]);
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Tab") keyboardIntent.current = true;
      if (event.key !== "Escape" || !descriptionOpen) return;
      event.preventDefault();
      setDismissed(true);
    };
    const onPointer = () => {
      keyboardIntent.current = false;
    };
    window.addEventListener("keydown", onKey);
    window.addEventListener("pointerdown", onPointer);
    return () => {
      window.removeEventListener("keydown", onKey);
      window.removeEventListener("pointerdown", onPointer);
    };
  }, [descriptionOpen]);
  const showKeyboardFocus = (event: FocusEvent<HTMLElement>) => {
    const target = event.target;
    if (!(target instanceof Element) || description === undefined) return;
    const row = event.currentTarget.hasAttribute("data-sidebar-row")
      ? event.currentTarget
      : event.currentTarget.closest("[data-sidebar-row]");
    if (row === null || owningSidebarRow(target) !== row) return;
    const focusVisibleTarget = target.matches(":focus-visible") || keyboardIntent.current;
    if (focusVisibleTarget) {
      focusRef.current = true;
      setKeyboardFocused(true);
    }
  };

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
    <div className={contentBaseClassName}>
      <div
        className={`${ICON.navItem} flex items-center justify-center shrink-0${level > 0 && !item.icon ? " invisible" : ""}`}
      >
        {item.icon}
      </div>
      {!collapsed ? <span className="truncate flex-1 text-left">{item.label}</span> : null}
    </div>
  );

  return (
    <div
      ref={setRowElement}
      data-sidebar-row=""
      className="w-full"
      onFocus={showKeyboardFocus}
      onBlur={(event) => {
        if (eventTargetInRow(event.relatedTarget, event.currentTarget)) return;
        setKeyboardFocused(false);
      }}
    >
      {hasSubItems ? (
        <div className={contentBaseClassName}>
          <button
            type="button"
            onClick={handleClick}
            onFocus={showKeyboardFocus}
            className={`${collapsed ? "min-w-11 justify-center" : "min-w-0"} flex-1 min-h-11 h-full flex items-center gap-3 text-left`}
            title={collapsed ? item.label : undefined}
            aria-label={collapsed ? item.label : undefined}
            aria-describedby={describedBy}
            aria-expanded={collapsed ? undefined : isExpanded}
          >
            <div
              className={`${ICON.navItem} flex items-center justify-center shrink-0${level > 0 && !item.icon ? " invisible" : ""}`}
            >
              {item.icon}
            </div>
            {!collapsed ? <span className="truncate flex-1">{item.label}</span> : null}
          </button>
          {!collapsed ? (
            <button
              type="button"
              onClick={handleChevronClick}
              onFocus={showKeyboardFocus}
              aria-label={isExpanded ? `${item.label}を折りたたむ` : `${item.label}を展開`}
              aria-describedby={describedBy}
              className={`min-h-11 min-w-11 flex items-center justify-center rounded ${C.hoverBgMedium} transition-colors`}
            >
              <ChevronDown
                className={`${ICON.xs} transition-transform${isExpanded ? " rotate-180" : ""}`}
              />
            </button>
          ) : null}
        </div>
      ) : (
        <Link
          to={item.path || "#"}
          className="w-full block"
          title={collapsed ? item.label : undefined}
          aria-current={isActive ? "page" : undefined}
          aria-describedby={describedBy}
          onFocus={showKeyboardFocus}
        >
          {content}
        </Link>
      )}

      {hasSubItems && isExpanded && !collapsed ? (
        <div className="space-y-0.5 mt-0.5 mb-1">
          {item.subItems?.map((sub) => (
            <SidebarItemWithPermission
              key={sub.label}
              item={sub}
              collapsed={collapsed}
              level={level + 1}
            />
          ))}
        </div>
      ) : null}
      {description !== undefined && descriptionOpen ? (
        <PurposeTooltip
          id={descriptionId}
          text={description}
          exposure={exposure}
          anchor={rowElement}
          onPointerEnter={() => {
            hoverRef.current = true;
            setHovered(true);
          }}
          onPointerLeave={(next) => {
            if (rowElement && eventTargetInRow(next, rowElement)) return;
            hoverRef.current = false;
            setHovered(false);
            if (!focusRef.current) setDismissed(false);
          }}
        />
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

  const selfOk = item.resource === undefined || hasPermission(item.resource, "view");
  // Parent failed its own resource but a child is visible: expand-only shell (no default navigate).
  const renderItem: MenuItem =
    !selfOk && item.subItems?.length ? { ...item, resource: undefined, path: undefined } : item;

  return (
    <SidebarItem item={renderItem} purposePath={item.path} collapsed={collapsed} level={level} />
  );
});
