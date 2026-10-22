/**
 * EMR-523: パッケージマスタ編集ページ（公開契約 mock 接続）。
 * 契約正本: backend/docs/api.yaml `TreatmentPackage*`（採用 SHA f1ded6ec367459d91baec08414ce780df39e9a3f）。
 * 実 BE 応答の結合確認は NOT_RUN（PR では「実BE未結合確認」と明記する）。
 */
import { useCallback, useRef, useState } from "react";
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
  canSendReorder,
  moveId,
  reorderBlockReason,
  treatmentPackageKindViewResource,
  validateDraft,
  isMasterItemKind,
  type ComponentDraft,
  type KindViewResource,
} from "../lib/treatment-package-form";

export function TreatmentPackageSettings() {
  // TASK-444-S1: 新規ファイルは @/types/generated/models を import しない。
  // "master-medical" リソースリテラルは Resource union のメンバなので型検査を通る。
  const { canView, canCreate, canEdit, canDelete } = usePermission("master-medical");
  // Q3: 型別閲覧権限の再確認用。resource 種別は固定 3 種なのでトップレベルで呼べる。
  const checkupsPerm = usePermission("checkups");
  const merchandisePerm = usePermission("master-merchandise");
  const kindViewCanView: Record<KindViewResource, boolean> = {
    "master-medical": canView,
    checkups: checkupsPerm.canView,
    "master-merchandise": merchandisePerm.canView,
  };
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
  // パネルを権限エラー表示へ差し替える。Q3: 一覧取得成功だけでは latch を解除しない。
  // 解除は「権限を再確認」の明示操作（対象 fresh GET + 保持 draft 各型の現在 view 権限
  // 確認）のみ。5xx・通信失敗は失権扱いしない（403 の後に 5xx が続いても latch は維持）。
  // lostAt は失権を観測した時点の dataUpdatedAt（記録用）。
  // useEffect 同期は set-state-in-effect になるため、レンダー中の比較で同期する。
  const listForbidden = isForbiddenError(list.error);
  const [lostAt, setLostAt] = useState<number | null>(() =>
    listForbidden ? list.dataUpdatedAt : null,
  );
  const [prevListForbidden, setPrevListForbidden] = useState(listForbidden);
  if (prevListForbidden !== listForbidden) {
    setPrevListForbidden(listForbidden);
    if (listForbidden) {
      setLostAt((prev) => prev ?? list.dataUpdatedAt);
    }
    // 一覧成功だけでは解除しない（Q3）。
  }
  const authorizationLost = listForbidden || lostAt !== null;
  const markAuthorizationLost = useCallback(() => {
    setLostAt(list.dataUpdatedAt);
    void list.refetch();
  }, [list]);

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

  // A1-01: 編集 session 識別子。openTarget（開く/切替/閉鎖）と S08 明示選択・削除成功で
  // 増やす。保存開始時に捕捉し、await 後に一致確認して遅延応答を破棄する。
  const editSessionRef = useRef(0);
  const bumpEditSession = useCallback(() => {
    editSessionRef.current += 1;
  }, []);

  // Q3: 失権からの再表示は明示的な「権限を再確認」のみ。対象 fresh GET（既存対象のみ）
  // + 保持 draft 各型（manual_other 除く）+ master-medical の現在 view 権限を確認し、
  // 全て通ったときだけ latch を解除する。target/baseline/draft は変更しない。
  // A2-02: 解除判定は GET 応答後のレンダーで「現在の」権限 boolean で行う
  //（クリック時の snapshot を使わない）。GET 成功は get-ok として state に記録し、
  // 下段のレンダー中比較（prev-compare パターン）で判定する。
  type RecheckState =
    | { kind: "idle" }
    | { kind: "checking" }
    | { kind: "get-ok" }
    | { kind: "denied-get-forbidden" }
    | { kind: "denied-get-deleted" }
    | { kind: "denied-view" }
    | { kind: "failed" };
  const [recheck, setRecheck] = useState<RecheckState>({ kind: "idle" });
  const recheckTokenRef = useRef(0);
  // A2-03: 依存は primitive/安定値のみ（毎レンダー生成の kindViewCanView は依存に入れない）。
  const handleRecheckPermission = useCallback(async () => {
    if (target === null) return;
    const token = ++recheckTokenRef.current;
    const sessionTarget = target;
    setRecheck({ kind: "checking" });
    if (sessionTarget !== "new") {
      try {
        await getTreatmentPackage(sessionTarget.id);
      } catch (error) {
        if (recheckTokenRef.current !== token) return;
        if (isNotFoundError(error)) {
          setRecheck({ kind: "denied-get-deleted" });
        } else if (isForbiddenError(error)) {
          setRecheck({ kind: "denied-get-forbidden" });
        } else {
          // 5xx・通信失敗は失権と区別する（latch は維持・追加失権扱いしない）。
          setRecheck({ kind: "failed" });
        }
        return;
      }
    }
    if (recheckTokenRef.current !== token) return;
    // 一覧の再取得は event handler 内の副作用として掛け、判定自体はレンダーで行う。
    void list.refetch();
    setRecheck({ kind: "get-ok" });
  }, [target, list]);

  // A2-02: get-ok の解除判定。レンダー中に現在の draft 各型（manual_other 除く）+
  // master-medical の view 権限を確認し、全て通る場合だけ latch を解除する。
  // 権限不足は denied-view へ遷移させる（再確認のやり直しは利用者の再クリック）。
  if (recheck.kind === "get-ok" && editor.draft !== null) {
    const needed = new Set<KindViewResource>(["master-medical"]);
    if (editor.draft !== null) {
      for (const c of editor.draft.components) {
        if (isMasterItemKind(c.kind)) needed.add(treatmentPackageKindViewResource(c.kind));
      }
    }
    const lacking = [...needed].filter((r) => !kindViewCanView[r]);
    if (lacking.length > 0) {
      setRecheck({ kind: "denied-view" });
    } else {
      setLostAt(null);
      setRecheck({ kind: "idle" });
    }
  }

  const openTarget = useCallback(
    (next: EditorTarget) => {
      dirty.runWithDiscardCheck(() => {
        clearConflictFetch();
        bumpEditSession();
        recheckTokenRef.current += 1;
        setRecheck({ kind: "idle" });
        setTarget(next);
        editor.resetFor(next);
        setFormErrors([]);
        setComponentErrors(new Map());
      });
    },
    [dirty, editor, clearConflictFetch, bumpEditSession],
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
    const sessionAtStart = editSessionRef.current;
    const outcome = await editor.save(
      {
        create: (body) => createMut.mutateAsync(body),
        update: (id, body) => updateMut.mutateAsync({ id, req: body }),
      },
      () => editSessionRef.current !== sessionAtStart,
    );
    // A1-01: 保存中に対象が切替・閉鎖された場合は遅延応答（409/成功/403/error）を
    // 新 session へ適用しない。
    if (editSessionRef.current !== sessionAtStart) return;
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
    bumpEditSession();
    setTarget(latest);
    editor.resetFor(latest);
    dirty.markClean();
  }, [conflictFetch, editor, dirty, clearConflictFetch, bumpEditSession]);

  // S08 明示選択 (b): 入力を保持し、CAS 基準版だけを最新へ取り直す。
  // resetFor は呼ばない（draft 不変・dirty 継続）。保存は利用者の再操作のみ。
  const handleKeepEditingOnLatest = useCallback(() => {
    if (conflictFetch?.kind !== "ready") return;
    const latest = conflictFetch.latest;
    clearConflictFetch();
    bumpEditSession();
    setTarget(latest);
    editor.clearConflict();
  }, [conflictFetch, editor, clearConflictFetch, bumpEditSession]);

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
          bumpEditSession();
          recheckTokenRef.current += 1;
          setRecheck({ kind: "idle" });
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

  // Q1: 並替の事前抑止。raw（filter 前）items が閾値以上・pending・error・
  // order_version なしではボタンと handler の双方で送信しない。
  const reorderBlock = reorderBlockReason({
    isPending: list.isPending,
    isError: list.isError,
    orderVersion,
    rawItemCount: list.data?.items.length ?? 0,
  });

  const handleMove = useCallback(
    (id: string, direction: -1 | 1) => {
      // Q1: 事前抑止は純粋判定 canSendReorder に集約（handler と disabled の双方で使用）。
      if (
        !canSendReorder({
          authorizationLost,
          block: reorderBlock,
          orderVersion,
          canEdit: canEdit === true,
        })
      ) {
        return;
      }
      const ids = packages.map((p) => p.id);
      // canSendReorder は boolean 判定のみなので、buildReorderBody 前に narrowing する。
      if (orderVersion === undefined) return;
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
    [
      orderVersion,
      canEdit,
      packages,
      reorderMut,
      authorizationLost,
      reorderBlock,
      markAuthorizationLost,
    ],
  );

  const sidePanel = (
    <div className={`w-[420px] shrink-0 border-l ${C.borderLight} overflow-y-auto`}>
      {authorizationLost && target !== null ? (
        // S07: 失権中は editor を権限エラーへ差し替える。draft は hook 内に保持（非表示）。
        // Q3: 再表示は「権限を再確認」の明示操作のみ（対象 fresh GET + draft 各型 view 確認）。
        <div className="p-4">
          <ErrorFallback message="この構成を表示する権限がありません。権限を確認してから再読込してください。" />
          <div className="mt-3">
            <button
              type="button"
              className="min-h-[44px] min-w-[44px] text-sm underline"
              disabled={recheck.kind === "checking"}
              onClick={() => void handleRecheckPermission()}
            >
              権限を再確認
            </button>
            {recheck.kind === "checking" ? (
              <p className={`mt-2 text-sm ${C.text60}`}>権限を確認しています…</p>
            ) : recheck.kind === "denied-get-forbidden" ? (
              <p role="alert" className={`mt-2 text-sm ${C.danger}`}>
                対象パッケージへのアクセス権がありません。
              </p>
            ) : recheck.kind === "denied-get-deleted" ? (
              <p role="alert" className={`mt-2 text-sm ${C.danger}`}>
                対象パッケージは削除されています。
              </p>
            ) : recheck.kind === "denied-view" ? (
              <p role="alert" className={`mt-2 text-sm ${C.danger}`}>
                構成に含まれる型の閲覧権限がありません。
              </p>
            ) : recheck.kind === "failed" ? (
              <p role="alert" className={`mt-2 text-sm ${C.danger}`}>
                再確認中に取得エラーが発生しました。権限状態は不明です。
              </p>
            ) : null}
          </div>
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
            canSave={
              (target === "new" ? canCreate : canEdit) &&
              !editor.createUnknown &&
              conflictFetch?.kind !== "deleted"
            }
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
          <>
            {reorderBlock === "count-guard" ? (
              // Q1: 「切詰めが確定した」とは表示しない。保守的な事前抑止の告知のみ。
              <p role="alert" className={`px-3 py-2 text-sm ${C.danger}`}>
                全件取得を確認できないため並替できません
              </p>
            ) : null}
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
                            disabled={
                              reorderMut.isPending || authorizationLost || reorderBlock !== null
                            }
                            onClick={() => handleMove(p.id, -1)}
                          >
                            ↑
                          </button>
                          <button
                            type="button"
                            aria-label={`${p.name}を下へ`}
                            className="min-h-[44px] min-w-[44px]"
                            disabled={
                              reorderMut.isPending || authorizationLost || reorderBlock !== null
                            }
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
          </>
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
