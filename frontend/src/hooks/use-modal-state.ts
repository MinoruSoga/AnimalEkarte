import { useCallback, useState } from "react";

/**
 * モーダル・ダイアログの開閉状態を管理するフック。
 * `useState<T | null>` パターンを共通化し、命名・構造を統一する。
 *
 * @example
 * const deleteModal = useModalState<{ id: string; label: string }>();
 * deleteModal.open({ id: "1", label: "削除対象" });
 * deleteModal.isOpen // true
 * deleteModal.item   // { id: "1", label: "削除対象" }
 * deleteModal.close();
 */
export function useModalState<T>(initialValue: T | null = null) {
  const [item, setItem] = useState<T | null>(initialValue);

  const open = useCallback((value: T) => setItem(value), []);
  const close = useCallback(() => setItem(null), []);

  return {
    item,
    isOpen: item !== null,
    open,
    close,
  };
}
