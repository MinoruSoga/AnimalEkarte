// React/Framework
import {
  useState,
  useRef,
  useMemo,
  useCallback,
  useDeferredValue,
  useTransition,
  useLayoutEffect,
} from "react";
import { normalizeKana } from "@/lib/normalize-kana";

// External
import { toast } from "sonner";

// Types
import type { UseMutationResult } from "@tanstack/react-query";
import type { ActiveFilter, ActiveSort } from "@/components/shared/PropertyFilter/types";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

/** Minimum shape required for entities managed by useMasterCRUD */
interface MasterEntity {
  id: string;
}

interface MasterCRUDPermissions {
  canDelete: boolean;
}

interface UseMasterCRUDOptions<T extends MasterEntity> {
  /** Raw data from useQuery */
  data: T[] | undefined;

  /** Delete mutation hook result */
  deleteMutation: UseMutationResult<void, Error, string>;

  /** Entity label for toast messages (e.g. "ケージ", "スタッフ") */
  entityLabel: string;

  /**
   * Custom search filter. Defaults to name match.
   * Return true if item matches the search term.
   */
  searchFilter?: (item: T, term: string) => boolean;

  /**
   * Custom filter application for PropertyFilter activeFilters.
   * Return true if item matches all filters. Defaults to isActive status filter.
   */
  activeFilterApply?: (item: T, filters: ActiveFilter[]) => boolean;

  /**
   * BUG-380: サイドパネル編集中の未保存変更を管理するガード。
   * 指定された場合、別行クリック・パネル閉じ・新規作成時に確認ダイアログを出す。
   * 本番は runWithDiscardCheck（継続処理を保持）。confirmDiscard はテスト mock 互換。
   */
  dirtyGuard?: {
    confirmDiscard?: () => boolean;
    runWithDiscardCheck?: (fn: () => void) => void;
  };

  /** Delete permission enforced at the mutation boundary. */
  permissions: MasterCRUDPermissions;
}

export interface UseMasterCRUDReturn<T extends MasterEntity> {
  // ── State ──
  editTarget: T | "new" | null;
  setEditTarget: (target: T | "new" | null) => void;
  searchTerm: string;
  setSearchTerm: (term: string) => void;
  pendingDelete: T | null;
  setPendingDelete: (item: T | null) => void;
  isSavePending: boolean;

  // ── Filter / Sort state ──
  activeFilters: ActiveFilter[];
  setActiveFilters: (filters: ActiveFilter[]) => void;
  activeSorts: ActiveSort[];
  setActiveSorts: (sorts: ActiveSort[]) => void;

  // ── Derived ──
  filteredItems: T[];
  panelItem: T | null;
  isEditing: boolean;

  // ── Handlers ──
  handleClose: () => void;
  handleNew: () => void;
  handleEdit: (item: T) => void;
  handleDeleteRequest: (item: T) => void;
  handleDeleteConfirm: () => void;
  handleDeleteCancel: () => void;
  handleSortChange: (sorts: ActiveSort[]) => void;

  // ── Save transition (pages use this for custom save logic) ──
  startSaveTransition: React.TransitionStartFunction;
}

// ─────────────────────────────────────────────────
// Default search filter
// ─────────────────────────────────────────────────

export function defaultSearchFilter<T extends MasterEntity>(item: T, term: string): boolean {
  if ("name" in item && typeof item.name === "string") {
    return normalizeKana(item.name).toLowerCase().includes(term);
  }
  return false;
}

// ─────────────────────────────────────────────────
// Default active filter application (isActive status)
// ─────────────────────────────────────────────────

export function defaultActiveFilterApply<T extends MasterEntity>(
  item: T,
  filters: ActiveFilter[],
): boolean {
  const record = item as Record<string, unknown>;
  for (const filter of filters) {
    if (filter.key === "status" && typeof filter.value === "string") {
      const hasIsActive = "isActive" in item && typeof record.isActive === "boolean";
      if (!hasIsActive) continue;
      const isActive = record.isActive as boolean;
      const filterActive = filter.value === "active";
      switch (filter.condition) {
        case "is":
          if (isActive !== filterActive) return false;
          break;
        case "is_not":
          if (isActive === filterActive) return false;
          break;
        default:
          break;
      }
    } else if (filter.condition === "is_empty") {
      // 汎用的な空チェック: null / undefined / "" は空とみなす
      const v = record[filter.key];
      if (v != null && v !== "") return false;
    } else if (filter.condition === "is_not_empty") {
      const v = record[filter.key];
      if (v == null || v === "") return false;
    }
  }
  return true;
}

// ─────────────────────────────────────────────────
// Default sort comparator
// ─────────────────────────────────────────────────

export function applySorts<T extends MasterEntity>(items: T[], sorts: ActiveSort[]): T[] {
  if (sorts.length === 0) return items;
  const sorted = [...items];
  sorted.sort((a, b) => {
    for (const sort of sorts) {
      const aVal = String((a as Record<string, unknown>)[sort.key] ?? "");
      const bVal = String((b as Record<string, unknown>)[sort.key] ?? "");
      const cmp = aVal.localeCompare(bVal, "ja");
      if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
    }
    return 0;
  });
  return sorted;
}

// ─────────────────────────────────────────────────
// Hook
// ─────────────────────────────────────────────────

export function useMasterCRUD<T extends MasterEntity>({
  data,
  deleteMutation,
  entityLabel,
  searchFilter = defaultSearchFilter,
  activeFilterApply = defaultActiveFilterApply,
  dirtyGuard,
  permissions,
}: UseMasterCRUDOptions<T>): UseMasterCRUDReturn<T> {
  // ── State ──
  const [editTarget, setEditTarget] = useState<T | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<T | null>(null);
  // rerender-dependencies: pendingDelete オブジェクトを ref 経由で参照し handleDeleteConfirm deps から除外
  const pendingDeleteRef = useRef<T | null>(null);
  useLayoutEffect(() => {
    pendingDeleteRef.current = pendingDelete;
  }, [pendingDelete]);
  const canDelete = permissions.canDelete;
  const permissionsRef = useRef<MasterCRUDPermissions>({
    canDelete: canDelete === true,
  });
  useLayoutEffect(() => {
    permissionsRef.current = { canDelete: canDelete === true };
  }, [canDelete]);
  const [isSavePending, startSaveTransition] = useTransition();
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);

  // ── Search filter (rerender-transitions) ──
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    let items = data ?? [];

    // PropertyFilter activeFilters 適用
    if (activeFilters.length > 0) {
      items = items.filter((item) => activeFilterApply(item, activeFilters));
    }

    // テキスト検索（カタカナ・ひらがな非区別）
    if (deferredSearch) {
      const normalizedTerm = normalizeKana(deferredSearch).toLowerCase();
      items = items.filter((item) => searchFilter(item, normalizedTerm));
    }

    // ソート
    items = applySorts(items, activeSorts);

    return items;
  }, [data, activeFilters, deferredSearch, activeSorts, searchFilter, activeFilterApply]);

  // ── Derived values ──
  const panelItem = editTarget !== null && editTarget !== "new" ? editTarget : null;
  const isEditing = editTarget !== null;

  // ── Handlers ──
  // BUG-380: dirtyGuard 指定時は未保存変更の破棄確認を挟む。
  const withDirtyGuard = useCallback(
    (fn: () => void) => {
      if (dirtyGuard?.runWithDiscardCheck) {
        dirtyGuard.runWithDiscardCheck(fn);
        return;
      }
      if (dirtyGuard?.confirmDiscard && !dirtyGuard.confirmDiscard()) return;
      fn();
    },
    [dirtyGuard],
  );

  const handleClose = useCallback(() => {
    withDirtyGuard(() => setEditTarget(null));
  }, [withDirtyGuard]);

  const handleNew = useCallback(() => {
    withDirtyGuard(() => setEditTarget("new"));
  }, [withDirtyGuard]);

  const handleEdit = useCallback(
    (item: T) => {
      withDirtyGuard(() => setEditTarget(item));
    },
    [withDirtyGuard],
  );

  const handleDeleteRequest = useCallback((item: T) => setPendingDelete(item), []);

  const handleDeleteCancel = useCallback(() => setPendingDelete(null), []);

  const { mutate: deleteMasterFn } = deleteMutation;
  const handleDeleteConfirm = useCallback(() => {
    const target = pendingDeleteRef.current;
    if (!target) return;
    const currentPermissions = permissionsRef.current;
    if (currentPermissions.canDelete !== true) return;
    // onError は deleteMutation 側（各 master/api/*.ts の useDeleteXxx）で handleApiError 済み。
    deleteMasterFn(target.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setEditTarget(null);
        toast.success(`${entityLabel}を削除しました`);
      },
    });
  }, [deleteMasterFn, entityLabel]);

  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  return {
    // State
    editTarget,
    setEditTarget,
    searchTerm,
    setSearchTerm,
    pendingDelete,
    setPendingDelete,
    isSavePending,

    // Filter / Sort
    activeFilters,
    setActiveFilters,
    activeSorts,
    setActiveSorts,

    // Derived
    filteredItems,
    panelItem,
    isEditing,

    // Handlers
    handleClose,
    handleNew,
    handleEdit,
    handleDeleteRequest,
    handleDeleteConfirm,
    handleDeleteCancel,
    handleSortChange,

    // Save transition
    startSaveTransition,
  };
}
