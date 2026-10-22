/**
 * EMR-523: パッケージマスタ編集ページ（公開契約 mock 接続）。
 * 契約正本: backend/docs/api.yaml `TreatmentPackage*`（採用 SHA f1ded6ec367459d91baec08414ce780df39e9a3f）。
 * 実 BE 応答の結合確認は NOT_RUN（PR では「実BE未結合確認」と明記する）。
 */
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { ErrorFallback, LoadingFallback, EmptyState } from "@/components/shared/DataStates";
import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { handleApiError } from "@/lib/handle-api-error";
import { MasterPageShell } from "../components/MasterPageShell";
import { TreatmentPackageEditor } from "../components/TreatmentPackageEditor";
import { TreatmentPackageConflictCompare } from "../components/TreatmentPackageConflictCompare";
import { getTreatmentPackage, type TreatmentPackage } from "../api/treatment-packages";
import {
  useCreateTreatmentPackage,
  useDeleteTreatmentPackage,
  useGetTreatmentPackages,
  useReorderTreatmentPackages,
  useTreatmentPackageEditor,
  useUpdateTreatmentPackage,
  isConflictError,
  isForbiddenError,
  isNotFoundError,
  type EditorTarget,
} from "../hooks/use-treatment-packages";
import {
  buildReorderBody,
  moveId,
  validateDraft,
  type ComponentDraft,
} from "../lib/treatment-package-form";

export function TreatmentPackageSettings() {
  // TASK-444-S1: 新規ファイルは @/types/generated/models を import しない。
  // "master-medical" リテラルは Resource union のメンバなので型検査を通る。
  const { canView, canCreate, canEdit, canDelete } = usePermission("master-medical");
  const list = useGetTreatmentPackages();
  const createMut = useCreateTreatmentPackage();
  const updateMut = useUpdateTreatmentPackage();
  const deleteMut = useDeleteTreatmentPackage();
  const reorderMut = useReorderTreatmentPackages();
  const [target, setTarget] = useState<EditorTarget>(null);
  const [pendingDelete, setPendingDelete] = useState<TreatmentPackage | null>(null);
  const [formErrors, setFormErrors] = useState<readonly string[]>([]);
  const [componentErrors, setComponentErrors] = useState<ReadonlyMap<string, readonly string[]>>(
    new Map(),
  );
  const dirty = useSidePeekDirty();
  const editor = useTreatmentPackageEditor(target);

  // S07: 型別 view 失権（403）は latch する。draft は hook 内に保持したまま
  // パネルを権限エラー表示へ差し替え、後続の一覧取得成功でのみ表示を戻す。
  // 5xx・通信失敗は失権扱いしない。
  const [authorizationLost, setAuthorizationLost] = useState(false);
  const authLostBaseline = useRef(0);
  const markAuthorizationLost = useCallback(() => {
    authLostBaseline.current = list.dataUpdatedAt;
    setAuthorizationLost(true);
    void list.refetch();
  }, [list]);
  useEffect(() => {
    if (isForbiddenError(list.error)) {
      authLostBaseline.current = list.dataUpdatedAt;
      setAuthorizationLost(true);
      return;
    }
    if (authorizationLost && list.dataUpdatedAt > authLostBaseline.current) {
      setAuthorizationLost(false);
    }
  }, [list.error, list.dataUpdatedAt, authorizationLost]);

  // S08: 409 後の最新取得・比較・明示選択。キャッシュを使わず直接 GET し、
  // 遅延応答・対象切替・editor 閉鎖は要求トークンで破棄する。
  type ConflictFetch =
    | { kind: "fetching"; targetId: string }
    | { kind: "ready"; targetId: string; latest: TreatmentPackage }
    | { kind: "failed"; targetId: string }
    | { kind: "deleted"; targetId: string };
  const [conflictFetch, setConflictFetch] = useState<ConflictFetch | null>(null);
  const conflictTokenRef = useRef(0);
  const clearConflictFetch = useCallback(() => {
    conflictTokenRef.current += 1;
    setConflictFetch(null);
  }, []);
  const beginConflictFetch = useCallback(
    (packageId: string) => {
      const token = ++conflictTokenRef.current;
      setConflictFetch({ kind: "fetching", targetId: packageId });
      void getTreatmentPackage(packageId)
        .then((latest) => {
          if (conflictTokenRef.current !== token) return;
          setConflictFetch({ kind: "ready", targetId: packageId, latest });
        })
        .catch((error: unknown) => {
          if (conflictTokenRef.current !== token) return;
          if (isForbiddenError(error)) {
            markAuthorizationLost();
            return;
          }
          if (isNotFoundError(error)) {
            setConflictFetch({ kind: "deleted", targetId: packageId });
            return;
          }
          setConflictFetch({ kind: "failed", targetId: packageId });
        });
    },
    [markAuthorizationLost],
  );

  const packages = list.data?.items ?? [];
  const orderVersion = list.data?.order_version;

  const openTarget = useCallback(
    (next: EditorTarget) => {
      dirty.runWithDiscardCheck(() => {
        clearConflictFetch();
        setTarget(next);
        editor.resetFor(next);
        setFormErrors([]);
        setComponentErrors(new Map());
      });
    },
    [dirty, editor, clearConflictFetch],
  );

  const handleSave = useCallback(async () => {
    if (authorizationLost) return; // 失権中は送信しない（draft は保持）。
    if (editor.draft === null) return;
    const v = validateDraft(editor.draft);
    if (!v.valid) {
      setFormErrors(v.errors);
      setComponentErrors(v.componentErrors);
      return;
    }
    const outcome = await editor.save({
      create: (body) => createMut.mutateAsync(body),
      update: (id, body) => updateMut.mutateAsync({ id, req: body }),
    });
    if (outcome.kind === "saved") {
      toast.success("パッケージを保存しました");
      clearConflictFetch();
      setTarget(outcome.pkg);
      editor.resetFor(outcome.pkg);
      dirty.markClean();
      return;
    }
    if (outcome.kind === "noop") {
      toast.success("変更はありません");
      return;
    }
    if (outcome.kind === "conflict") {
      // S08: edit のみ。draft を保持したまま最新を直接取得して比較へ進む。
      const baseId = editor.baseline?.id;
      if (baseId !== undefined) beginConflictFetch(baseId);
      return;
    }
    if (
      (outcome.kind === "error" || outcome.kind === "create-unknown") &&
      isForbiddenError(outcome.error)
    ) {
      markAuthorizationLost(); // mutation の 403 は一覧へも反映させる。
      return;
    }
    // error / create-unknown: 入力は保持。各アラートはエディタが表示。
  }, [
    editor,
    createMut,
    updateMut,
    dirty,
    authorizationLost,
    beginConflictFetch,
    clearConflictFetch,
    markAuthorizationLost,
  ]);

  // S08 明示選択 (a): 最新を採用し入力を破棄。自動保存はしない。
  const handleAdoptLatest = useCallback(() => {
    if (conflictFetch?.kind !== "ready") return;
    const latest = conflictFetch.latest;
    clearConflictFetch();
    setTarget(latest);
    editor.resetFor(latest);
    dirty.markClean();
  }, [conflictFetch, editor, dirty, clearConflictFetch]);

  // S08 明示選択 (b): 入力を保持し、CAS 基準版だけを最新へ取り直す。
  // resetFor は呼ばない（draft 不変・dirty 継続）。保存は利用者の再操作のみ。
  const handleKeepEditingOnLatest = useCallback(() => {
    if (conflictFetch?.kind !== "ready") return;
    const latest = conflictFetch.latest;
    clearConflictFetch();
    setTarget(latest);
    editor.clearConflict();
  }, [conflictFetch, editor, clearConflictFetch]);

  const handleDeleteConfirm = useCallback(() => {
    if (authorizationLost) return;
    if (pendingDelete === null) return;
    if (canDelete !== true) return;
    deleteMut.mutate(
      { id: pendingDelete.id, version: pendingDelete.version },
      {
        onSuccess: () => {
          toast.success("パッケージを削除しました");
          setPendingDelete(null);
          clearConflictFetch();
          setTarget(null);
          editor.resetFor(null);
        },
        onError: (error) => {
          if (isForbiddenError(error)) {
            markAuthorizationLost();
            return;
          }
          if (isConflictError(error)) {
            handleApiError(error, "パッケージの削除（版が最新ではありません。再読込してください）");
          }
        },
      },
    );
  }, [
    pendingDelete,
    canDelete,
    deleteMut,
    editor,
    authorizationLost,
    clearConflictFetch,
    markAuthorizationLost,
  ]);

  const handleMove = useCallback(
    (id: string, direction: -1 | 1) => {
      if (authorizationLost) return;
      if (orderVersion === undefined || canEdit !== true) return;
      const ids = packages.map((p) => p.id);
      const next = moveId(ids, id, direction);
      if (next.join(",") === ids.join(",")) return;
      reorderMut.mutate(buildReorderBody(orderVersion, next), {
        onError: (error) => {
          if (isForbiddenError(error)) {
            markAuthorizationLost();
            return;
          }
          if (isConflictError(error)) {
            handleApiError(error, "並べ替え（版が最新ではありません。再読込してください）");
          }
        },
      });
    },
    [orderVersion, canEdit, packages, reorderMut, authorizationLost, markAuthorizationLost],
  );

  const sidePanel = (
    <div className={`w-[420px] shrink-0 border-l ${C.borderLight} overflow-y-auto`}>
      {authorizationLost ? (
        // S07: 失権中は editor を権限エラーへ差し替える。draft は hook 内に保持（非表示）。
        <div className="p-4">
          <ErrorFallback message="この構成を表示する権限がありません。権限を確認してから再読込してください。" />
        </div>
      ) : editor.draft !== null && target !== null ? (
        <>
          {conflictFetch !== null ? (
            <div className="p-4 pb-0">
              {conflictFetch.kind === "fetching" ? (
                <p className={`text-sm ${C.text60}`}>最新を取得しています…</p>
              ) : conflictFetch.kind === "deleted" ? (
                <div
                  role="alert"
                  className={`rounded-lg border ${C.borderLight} p-3 text-sm ${C.danger}`}
                >
                  このパッケージは削除されています。入力は保持されています。
                </div>
              ) : conflictFetch.kind === "failed" ? (
                <div
                  role="alert"
                  className={`rounded-lg border ${C.borderLight} p-3 text-sm ${C.danger}`}
                >
                  最新の取得に失敗しました。入力は保持されています。
                  <div className="mt-2">
                    <button
                      type="button"
                      className="min-h-[44px] min-w-[44px] text-sm underline"
                      onClick={() => beginConflictFetch(conflictFetch.targetId)}
                    >
                      再試行
                    </button>
                  </div>
                </div>
              ) : (
                <TreatmentPackageConflictCompare
                  draft={editor.draft}
                  latest={conflictFetch.latest}
                  onAdoptLatest={handleAdoptLatest}
                  onKeepEditing={handleKeepEditingOnLatest}
                />
              )}
            </div>
          ) : null}
          <TreatmentPackageEditor
            mode={target === "new" ? "create" : "edit"}
            draft={editor.draft}
            componentErrors={componentErrors}
            formErrors={formErrors}
            conflict={editor.conflict}
            createUnknown={editor.createUnknown}
            onResolveCreateUnknown={() => {
              // 再取得の成功を待ってから解除する。失敗・遅延中は再送禁止を維持。
              void list.refetch().then((res) => {
                if (!res.isError) editor.allowCreateRetry();
              });
            }}
            canSave={(target === "new" ? canCreate : canEdit) && !editor.createUnknown}
            canDelete={canDelete}
            saving={createMut.isPending || updateMut.isPending}
            deleting={deleteMut.isPending}
            onChange={(patch) => {
              editor.updateDraft(patch);
              dirty.markDirty();
            }}
            onComponentsChange={(components: ComponentDraft[]) => {
              editor.updateComponents(components);
              dirty.markDirty();
            }}
            onSave={() => void handleSave()}
            onDeleteRequest={
              target !== "new" && target !== null ? () => setPendingDelete(target) : undefined
            }
            onClose={() => openTarget(null)}
          />
        </>
      ) : null}
    </div>
  );

  return (
    <>
      <MasterPageShell
        title="パッケージマスタ"
        resource="master-medical"
        onNew={() => openTarget("new")}
        sidePanel={sidePanel}
      >
        {!canView ? (
          <ErrorFallback message="このマスタを表示する権限がありません" />
        ) : list.isError ? (
          <ErrorFallback message="パッケージの取得に失敗しました" />
        ) : list.isPending ? (
          <LoadingFallback />
        ) : packages.length === 0 ? (
          <EmptyState message="登録されたパッケージはありません。新規登録から作成できます。" />
        ) : (
          <table className="w-full text-sm" aria-label="パッケージ一覧">
            <thead>
              <tr className={`text-left ${C.text60}`}>
                <th className="px-3 py-2 font-medium">名称</th>
                <th className="px-3 py-2 font-medium">構成数</th>
                <th className="px-3 py-2 font-medium">状態</th>
                <th className="px-3 py-2 font-medium">操作</th>
              </tr>
            </thead>
            <tbody>
              {packages.map((p) => (
                <tr key={p.id} className={`border-t ${C.borderLight}`}>
                  <td className="px-3 py-2">
                    <button
                      type="button"
                      className="underline text-left"
                      onClick={() => openTarget(p)}
                    >
                      {p.name}
                    </button>
                  </td>
                  <td className="px-3 py-2 tabular-nums">{p.components.length}</td>
                  <td className="px-3 py-2">{p.is_active ? "有効" : "無効"}</td>
                  <td className="px-3 py-2">
                    {canEdit ? (
                      <span className="flex gap-1">
                        <button
                          type="button"
                          aria-label={`${p.name}を上へ`}
                          className="min-h-[44px] min-w-[44px]"
                          disabled={reorderMut.isPending || authorizationLost}
                          onClick={() => handleMove(p.id, -1)}
                        >
                          ↑
                        </button>
                        <button
                          type="button"
                          aria-label={`${p.name}を下へ`}
                          className="min-h-[44px] min-w-[44px]"
                          disabled={reorderMut.isPending || authorizationLost}
                          onClick={() => handleMove(p.id, 1)}
                        >
                          ↓
                        </button>
                      </span>
                    ) : null}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </MasterPageShell>
      {pendingDelete !== null ? (
        <div
          role="alertdialog"
          aria-label="削除確認"
          className="fixed inset-0 flex items-center justify-center"
        >
          <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} p-4`}>
            <p className="text-sm">
              「{pendingDelete.name}」を削除しますか？この操作は取り消せません。
            </p>
            <div className="mt-3 flex gap-3 justify-end">
              <button
                type="button"
                className="text-sm underline"
                onClick={() => setPendingDelete(null)}
              >
                キャンセル
              </button>
              <button
                type="button"
                className="text-sm font-medium underline"
                disabled={deleteMut.isPending}
                onClick={handleDeleteConfirm}
              >
                削除する
              </button>
            </div>
          </div>
        </div>
      ) : null}
      {dirty.discardDialog}
    </>
  );
}
