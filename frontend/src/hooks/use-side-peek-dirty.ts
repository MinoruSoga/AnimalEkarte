import { useCallback, useEffect, useRef } from "react";

import { useUnsavedChanges } from "./use-unsaved-changes";

/**
 * BUG-380: サイドパネル編集中の未保存変更を追跡し、破棄操作前に確認ダイアログを出すための共通フック。
 *
 * 対象シナリオ:
 *  - 別行の編集アイコンクリック → パネル差し替え前に確認
 *  - マスタ内タブ切替 → パネルクローズ前に確認
 *  - ブラウザタブクローズ / リロード → beforeunload ダイアログ
 *
 * 使い方:
 *  ```tsx
 *  const dirty = useSidePeekDirty();
 *  // 入力変更時
 *  onChange={(e) => { setValue(e.target.value); dirty.markDirty(); }}
 *  // 保存成功時
 *  onSuccess={() => dirty.markClean()}
 *  // 別行クリック時
 *  if (!dirty.confirmDiscard()) return;
 *  handleEdit(row);
 *  ```
 */
export function useSidePeekDirty() {
  const { isDirty, markDirty, markClean } = useUnsavedChanges();
  const isDirtyRef = useRef(false);

  useEffect(() => {
    isDirtyRef.current = isDirty;
  }, [isDirty]);

  /**
   * 未保存変更の破棄確認。dirty でなければ常に true。
   * dirty であれば window.confirm を出して、OK で markClean + true、Cancel で false。
   */
  const confirmDiscard = useCallback((): boolean => {
    if (!isDirtyRef.current) return true;
    const ok = window.confirm("未保存の変更があります。破棄してよろしいですか?");
    if (ok) {
      markClean();
    }
    return ok;
  }, [markClean]);

  return {
    isDirty,
    isDirtyRef,
    markDirty,
    markClean,
    confirmDiscard,
  };
}
