/**
 * EMR-523: パッケージマスタ編集ページ（公開契約 mock 接続）。
 * 契約正本: backend/docs/api.yaml `TreatmentPackage*`（採用 SHA f1ded6ec367459d91baec08414ce780df39e9a3f）。
 * 実 BE 応答の結合確認は NOT_RUN（PR では「実BE未結合確認」と明記する）。
 */
import { useCallback, useState } from "react";
import { toast } from "sonner";
import { ErrorFallback, LoadingFallback, EmptyState } from "@/components/shared/DataStates";
import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { handleApiError } from "@/lib/handle-api-error";
import { MasterPageShell } from "../components/MasterPageShell";
import { TreatmentPackageEditor } from "../components/TreatmentPackageEditor";
import type { TreatmentPackage } from "../api/treatment-packages";
import {
  useCreateTreatmentPackage,
  useDeleteTreatmentPackage,
  useGetTreatmentPackages,
  useReorderTreatmentPackages,
  useTreatmentPackageEditor,
  useUpdateTreatmentPackage,
  isConflictError,
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

  const packages = list.data?.items ?? [];
  const orderVersion = list.data?.order_version;

  const openTarget = useCallback(
    (next: EditorTarget) => {
      dirty.runWithDiscardCheck(() => {
        setTarget(next);
        editor.resetFor(next);
        setFormErrors([]);
        setComponentErrors(new Map());
      });
    },
    [dirty, editor],
  );

  const handleSave = useCallback(async () => {
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
      setTarget(outcome.pkg);
      editor.resetFor(outcome.pkg);
      dirty.markClean();
      return;
    }
    if (outcome.kind === "noop") {
      toast.success("変更はありません");
      return;
    }
    // conflict / create-unknown / error: 入力は保持。各アラートはエディタが表示。
  }, [editor, createMut, updateMut, dirty]);

  const handleDeleteConfirm = useCallback(() => {
    if (pendingDelete === null) return;
    if (canDelete !== true) return;
    deleteMut.mutate(
      { id: pendingDelete.id, version: pendingDelete.version },
      {
        onSuccess: () => {
          toast.success("パッケージを削除しました");
          setPendingDelete(null);
          setTarget(null);
          editor.resetFor(null);
        },
        onError: (error) => {
          if (isConflictError(error)) {
            handleApiError(error, "パッケージの削除（版が最新ではありません。再読込してください）");
          }
        },
      },
    );
  }, [pendingDelete, canDelete, deleteMut, editor]);

  const handleMove = useCallback(
    (id: string, direction: -1 | 1) => {
      if (orderVersion === undefined || canEdit !== true) return;
      const ids = packages.map((p) => p.id);
      const next = moveId(ids, id, direction);
      if (next.join(",") === ids.join(",")) return;
      reorderMut.mutate(buildReorderBody(orderVersion, next), {
        onError: (error) => {
          if (isConflictError(error)) {
            handleApiError(error, "並べ替え（版が最新ではありません。再読込してください）");
          }
        },
      });
    },
    [orderVersion, canEdit, packages, reorderMut],
  );

  const sidePanel = (
    <div className={`w-[420px] shrink-0 border-l ${C.borderLight} overflow-y-auto`}>
      {editor.draft !== null && target !== null ? (
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
          onChange={editor.updateDraft}
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
                          disabled={reorderMut.isPending}
                          onClick={() => handleMove(p.id, -1)}
                        >
                          ↑
                        </button>
                        <button
                          type="button"
                          aria-label={`${p.name}を下へ`}
                          className="min-h-[44px] min-w-[44px]"
                          disabled={reorderMut.isPending}
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
