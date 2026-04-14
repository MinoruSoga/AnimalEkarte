import { useCallback, useEffect, useRef, useState } from "react";

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
  const [isDirty, setIsDirty] = useState(false);
  const isDirtyRef = useRef(false);

  useEffect(() => {
    isDirtyRef.current = isDirty;
  }, [isDirty]);

  // beforeunload: タブクローズ・リロード防止
  useEffect(() => {
    if (!isDirty) return;
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = "";
    };
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [isDirty]);

  const markDirty = useCallback(() => setIsDirty(true), []);
  const markClean = useCallback(() => setIsDirty(false), []);

  /**
   * 未保存変更の破棄確認。dirty でなければ常に true。
   * dirty であれば window.confirm を出して、OK で markClean + true、Cancel で false。
   */
  const confirmDiscard = useCallback((): boolean => {
    if (!isDirtyRef.current) return true;
    const ok = window.confirm("未保存の変更があります。破棄してよろしいですか?");
    if (ok) {
      setIsDirty(false);
    }
    return ok;
  }, []);

  return {
    isDirty,
    isDirtyRef,
    markDirty,
    markClean,
    confirmDiscard,
  };
}
