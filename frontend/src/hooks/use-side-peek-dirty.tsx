import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";

import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";

import { useUnsavedChanges } from "./use-unsaved-changes";

const DISCARD_CONFIRM_TITLE = "未保存の変更があります";
const DISCARD_CONFIRM_DESCRIPTION = "破棄してよろしいですか?";

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
 *  dirty.runWithDiscardCheck(() => { handleEdit(row); });
 *  // ページ JSX
 *  {dirty.discardDialog}
 *  ```
 */
export function useSidePeekDirty() {
  const { isDirty, markDirty: setDirty, markClean: setClean } = useUnsavedChanges();
  const isDirtyRef = useRef(false);
  const pendingFnRef = useRef<(() => void) | null>(null);
  const [discardOpen, setDiscardOpen] = useState(false);

  useEffect(() => {
    isDirtyRef.current = isDirty;
  }, [isDirty]);

  const markDirty = useCallback(() => {
    isDirtyRef.current = true;
    setDirty();
  }, [setDirty]);

  const markClean = useCallback(() => {
    isDirtyRef.current = false;
    setClean();
  }, [setClean]);

  /**
   * 未保存変更の破棄確認。dirty でなければ常に true。
   * dirty であれば ConfirmDialog を開き false を返す（継続処理は走らせない）。
   * 本番の破棄経路は runWithDiscardCheck を使う。本関数はテスト mock 互換用。
   */
  const confirmDiscard = useCallback((): boolean => {
    if (!isDirtyRef.current) return true;
    pendingFnRef.current = null;
    setDiscardOpen(true);
    return false;
  }, []);

  /**
   * dirty でなければ fn を直ちに実行する。
   * dirty なら fn を保持して ConfirmDialog を開き、確認後に markClean + fn を実行する。
   */
  const runWithDiscardCheck = useCallback((fn: () => void): void => {
    if (!isDirtyRef.current) {
      fn();
      return;
    }
    pendingFnRef.current = fn;
    setDiscardOpen(true);
  }, []);

  const handleDiscardConfirm = useCallback(() => {
    const fn = pendingFnRef.current;
    pendingFnRef.current = null;
    setDiscardOpen(false);
    markClean();
    fn?.();
  }, [markClean]);

  const handleDiscardClose = useCallback(() => {
    pendingFnRef.current = null;
    setDiscardOpen(false);
  }, []);

  const discardDialog: ReactNode = useMemo(
    () => (
      <ConfirmDialog
        open={discardOpen}
        onClose={handleDiscardClose}
        onConfirm={handleDiscardConfirm}
        title={DISCARD_CONFIRM_TITLE}
        description={DISCARD_CONFIRM_DESCRIPTION}
      />
    ),
    [discardOpen, handleDiscardClose, handleDiscardConfirm],
  );

  return {
    isDirty,
    isDirtyRef,
    markDirty,
    markClean,
    confirmDiscard,
    runWithDiscardCheck,
    discardDialog,
  };
}
