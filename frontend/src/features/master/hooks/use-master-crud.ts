// React/Framework
import { useState, useMemo, useCallback, useDeferredValue, useTransition } from "react";

// External
import { toast } from "sonner";

// Types
import type { UseMutationResult } from "@tanstack/react-query";
import type { ActiveFilter, ActiveSort } from "@/components/shared/NotionFilter/types";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

/** Minimum shape required for entities managed by useMasterCRUD */
interface MasterEntity {
  id: string;
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
   * Custom filter application for NotionFilter activeFilters.
   * Return true if item matches all filters. Defaults to isActive status filter.
   */
  activeFilterApply?: (item: T, filters: ActiveFilter[]) => boolean;
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

function defaultSearchFilter<T extends MasterEntity>(item: T, term: string): boolean {
  if ("name" in item && typeof item.name === "string") {
    return item.name.toLowerCase().includes(term);
  }
  return false;
}

// ─────────────────────────────────────────────────
// Default active filter application (isActive status)
// ─────────────────────────────────────────────────

function defaultActiveFilterApply<T extends MasterEntity>(item: T, filters: ActiveFilter[]): boolean {
  for (const filter of filters) {
    if (filter.key === "status" && typeof filter.value === "string") {
      const hasIsActive = "isActive" in item && typeof (item as Record<string, unknown>).isActive === "boolean";
      if (!hasIsActive) continue;
      const isActive = (item as Record<string, unknown>).isActive as boolean;
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
    }
  }
  return true;
}

// ─────────────────────────────────────────────────
// Default sort comparator
// ─────────────────────────────────────────────────

function applySorts<T extends MasterEntity>(items: T[], sorts: ActiveSort[]): T[] {
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
}: UseMasterCRUDOptions<T>): UseMasterCRUDReturn<T> {
  // ── State ──
  const [editTarget, setEditTarget] = useState<T | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<T | null>(null);
  const [isSavePending, startSaveTransition] = useTransition();
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);

  // ── Search filter (rerender-transitions) ──
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    let items = data ?? [];

    // NotionFilter activeFilters 適用
    if (activeFilters.length > 0) {
      items = items.filter((item) => activeFilterApply(item, activeFilters));
    }

    // テキスト検索
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter((item) => searchFilter(item, lower));
    }

    // ソート
    items = applySorts(items, activeSorts);

    return items;
  }, [data, activeFilters, deferredSearch, activeSorts, searchFilter, activeFilterApply]);

  // ── Derived values ──
  const panelItem = editTarget !== null && editTarget !== "new" ? editTarget : null;
  const isEditing = editTarget !== null;

  // ── Handlers ──
  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleNew = useCallback(() => setEditTarget("new"), []);

  const handleEdit = useCallback((item: T) => setEditTarget(item), []);

  const handleDeleteRequest = useCallback((item: T) => setPendingDelete(item), []);

  const handleDeleteCancel = useCallback(() => setPendingDelete(null), []);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setEditTarget(null);
        toast.success(`${entityLabel}を削除しました`);
      },
      onError: () => toast.error(`${entityLabel}の削除に失敗しました`),
    });
  }, [pendingDelete, deleteMutation, entityLabel]);

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
