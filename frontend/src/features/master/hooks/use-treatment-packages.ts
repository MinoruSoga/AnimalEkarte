/**
 * EMR-523: パッケージ master の react-query hooks と編集セッション hook。
 * 契約正本: backend/docs/api.yaml `TreatmentPackage*`（採用 SHA f1ded6ec367459d91baec08414ce780df39e9a3f）。
 * wire 型は api/treatment-packages.ts（手書き zod・fail-closed）を参照。
 * version は canonical 文字列のまま扱い、編集開始時捕捉版を CAS に使う。
 */
import { useCallback, useMemo, useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { handleApiError } from "@/lib/handle-api-error";
import {
  createTreatmentPackage,
  deleteTreatmentPackage,
  getTreatmentPackage,
  listTreatmentPackages,
  reorderTreatmentPackages,
  updateTreatmentPackage,
  type TreatmentPackage,
  type TreatmentPackageCreate,
  type TreatmentPackagePatch,
  type TreatmentPackageReorder,
} from "../api/treatment-packages";
import {
  emptyPackageDraft,
  draftFromPackage,
  moveId,
  requireEditingVersion,
  toCreateBody,
  toPatchBody,
  type ComponentDraft,
  type TreatmentPackageDraft,
} from "../lib/treatment-package-form";

const CATEGORY = "treatment-packages";
const listKey = () => queryKeys.masters.category(CATEGORY);
const detailKey = (id: string) => ["masters", CATEGORY, "detail", id] as const;

/** 一覧。医院切替は query cache の clear+reload で旧応答を破棄する（clinicId はキーに含めない既定）。 */
export const useGetTreatmentPackages = () =>
  useQuery({
    queryKey: listKey(),
    queryFn: listTreatmentPackages,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useGetTreatmentPackage = (packageId: string | null) =>
  useQuery({
    queryKey: detailKey(packageId ?? ""),
    queryFn: () => getTreatmentPackage(packageId as string),
    enabled: packageId !== null,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

const invalidateAll = (qc: ReturnType<typeof useQueryClient>) =>
  qc.invalidateQueries({ queryKey: listKey() });

export const useCreateTreatmentPackage = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: TreatmentPackageCreate) => createTreatmentPackage(req),
    onSuccess: () => invalidateAll(qc),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useUpdateTreatmentPackage = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: TreatmentPackagePatch }) => {
      // 編集開始時捕捉版のみを CAS に使う（後の refetch 版で送らない）。
      requireEditingVersion(req.version);
      return updateTreatmentPackage(id, req);
    },
    onSuccess: (data) => {
      // CAS 用に応答 version を即キャッシュへ書き、stale version を送らない。
      qc.setQueryData(detailKey(data.id), data);
      return invalidateAll(qc);
    },
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useDeleteTreatmentPackage = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, version }: { id: string; version: string }) =>
      deleteTreatmentPackage(id, requireEditingVersion(version)),
    onSuccess: () => invalidateAll(qc),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useReorderTreatmentPackages = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: TreatmentPackageReorder) => {
      requireEditingVersion(req.order_version);
      return reorderTreatmentPackages(req);
    },
    onSuccess: () => invalidateAll(qc),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export function isConflictError(error: unknown): boolean {
  return isAxiosError(error) && error.response?.status === 409;
}

export type EditorTarget = TreatmentPackage | "new" | null;

export type SaveOutcome =
  | { kind: "saved"; pkg: TreatmentPackage }
  | { kind: "noop" }
  | { kind: "conflict" }
  | { kind: "invalid" }
  | { kind: "create-unknown"; error: unknown }
  | { kind: "create-unknown-blocked" }
  | { kind: "error"; error: unknown };

/**
 * 編集セッション。draft はユーザー入力保持が最優先で、競合・失敗でも破棄しない。
 * 409 は「再読込・比較」へ誘導するだけで、無断上書きや false/0 補完をしない。
 */
export function useTreatmentPackageEditor(target: EditorTarget) {
  const baseline = useMemo<TreatmentPackage | null>(
    () => (target !== null && target !== "new" ? target : null),
    [target],
  );
  const baselineRef = useRef<TreatmentPackage | null>(baseline);
  baselineRef.current = baseline;
  const [draft, setDraft] = useState<TreatmentPackageDraft | null>(() =>
    target === null ? null : target === "new" ? emptyPackageDraft() : draftFromPackage(target),
  );
  const [conflict, setConflict] = useState(false);
  const [createUnknown, setCreateUnknown] = useState(false);

  const resetFor = useCallback((next: EditorTarget) => {
    setDraft(next === null ? null : next === "new" ? emptyPackageDraft() : draftFromPackage(next));
    setConflict(false);
    setCreateUnknown(false);
  }, []);

  const updateDraft = useCallback((patch: Partial<TreatmentPackageDraft>) => {
    setDraft((prev) => (prev === null ? prev : { ...prev, ...patch }));
  }, []);

  const updateComponents = useCallback((components: ComponentDraft[]) => {
    setDraft((prev) => (prev === null ? prev : { ...prev, components }));
  }, []);

  const save = useCallback(
    async (ops: {
      create: (body: TreatmentPackageCreate) => Promise<TreatmentPackage>;
      update: (id: string, body: TreatmentPackagePatch) => Promise<TreatmentPackage>;
    }): Promise<SaveOutcome> => {
      if (draft === null) return { kind: "invalid" };
      const base = baselineRef.current;
      // create 結果不明の未解決がある限り同じ POST を再送しない（盲目再送禁止）。
      // 一覧再読込での確認後、allowCreateRetry で明示解除する。
      if (base === null && createUnknown) return { kind: "create-unknown-blocked" };
      try {
        if (base === null) {
          const body = toCreateBody(draft);
          if (body === null) return { kind: "invalid" };
          const pkg = await ops.create(body);
          setCreateUnknown(false);
          return { kind: "saved", pkg };
        }
        const body = toPatchBody(draft, base);
        if (body === null) {
          const componentsValid = toCreateBody(draft) !== null;
          if (!componentsValid) return { kind: "invalid" };
          return { kind: "noop" };
        }
        const pkg = await ops.update(base.id, body);
        setConflict(false);
        return { kind: "saved", pkg };
      } catch (error) {
        if (isConflictError(error)) {
          setConflict(true);
          return { kind: "conflict" };
        }
        if (base === null) {
          // create 応答不明: 盲目再送・同名検索で推定しない。入力保持したまま結果不明を示す。
          setCreateUnknown(true);
          return { kind: "create-unknown", error };
        }
        return { kind: "error", error };
      }
    },
    [draft, createUnknown],
  );

  /** create 結果不明の明示解決。一覧を再読込して結果を確認した後だけ呼ぶ。
   *  作成済みが見つかった場合は resetFor(作成済みpkg) で baseline を取り込み、
   *  未作成が確認できた場合だけこの解除で informed な再送を許す。 */
  const allowCreateRetry = useCallback(() => setCreateUnknown(false), []);

  const isDirty = useMemo(() => {
    if (draft === null) return false;
    if (baseline === null) return true;
    if (baselineRef.current === null && draft === null) return false;
    if (baseline === null) return true;
    const fresh = draftFromPackage(baseline);
    return JSON.stringify(draft) !== JSON.stringify(fresh);
  }, [draft, baseline]);

  return {
    draft,
    baseline,
    conflict,
    createUnknown,
    isDirty,
    resetFor,
    updateDraft,
    updateComponents,
    save,
    allowCreateRetry,
    clearConflict: useCallback(() => setConflict(false), []),
  };
}

export { moveId };
