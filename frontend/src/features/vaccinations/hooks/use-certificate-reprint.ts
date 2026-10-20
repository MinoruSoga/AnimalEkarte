import { useCallback, useSyncExternalStore } from "react";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import { newOperationKey } from "../api/certificate-operation-key";
import {
  readSucceededRefs,
  type CertificateEdition,
  type CertificateLayout,
  type CertificateResult,
} from "../api/certificate-schemas";
import { decideIssueFailure } from "../lib/certificate-issue-reducer";
import type { ReprintEntry, ReprintPending, ReprintState, ReprintStore } from "../lib/certificate-reprint-store";

export type { ReprintState };

/**
 * 保存版の再印刷（POST /reprints）。issue と同じ operation_key 規則:
 * 結果不明では同 key で照会し、新しい key は「成功なし」確定後の新しい操作でだけ作る。
 * key と進行状態は行の外の store（画面所有）に置く。行が unmount されても未確定の key を失わない。
 */
/** 結果が確定していない間（同じ key で照会・再試行するだけの状態）。 */
export function isReprintUnresolved(state: ReprintState): boolean {
  return state.phase === "running" || state.phase === "unknown" || state.phase === "not-issued";
}

export function useCertificateReprint(args: {
  api: CertificateApi;
  edition: CertificateEdition;
  sourceVersion: number;
  onRefetch: () => void;
  /** mutation 直前に最新の edit 権限を返す（UI の disabled だけを最終防壁にしない）。 */
  getCanEdit: () => boolean;
  store: ReprintStore;
}) {
  const { api, edition, sourceVersion, onRefetch, getCanEdit, store } = args;
  const editionKey = edition.id;
  const entry = useSyncExternalStore(
    store.subscribe,
    () => store.get(editionKey),
    () => store.get(editionKey),
  );
  const state = entry.state;
  // 実行中・結果待ち・完了の間は、実際に使った向きを返す（行が作り直されても表示がずれない）。
  const activeLayout = isReprintUnresolved(state) || state.phase === "done" ? entry.layout : null;

  const commit = useCallback(
    (patch: Partial<ReprintEntry>) => store.set(editionKey, { ...store.get(editionKey), ...patch }),
    [store, editionKey],
  );

  const apply = useCallback(
    (result: CertificateResult) => {
      if (result.status === "not-succeeded") {
        commit({ state: { phase: "not-issued", notice: "not-succeeded" } });
        return;
      }
      const refs = readSucceededRefs(result);
      if (refs === null) {
        commit({ state: { phase: "unknown", notice: "contract-violation" } });
        return;
      }
      commit({ pending: null, state: { phase: "done", ...refs, replayed: result.status === "replayed" } });
    },
    [commit],
  );

  const lookup = useCallback(
    async (pending: ReprintPending) => {
      try {
        apply(await api.lookupOperation(pending.request.operation_key, "reprint"));
      } catch (error) {
        const kind = error instanceof CertificateApiError ? error.kind : "network";
        if (kind === "scope-not-found") commit({ state: { phase: "not-issued", notice: "not-claimed" } });
        else if (kind === "forbidden") commit({ pending: null, state: { phase: "blocked", notice: "forbidden" } });
        else if (kind === "unauthenticated") {
          commit({ pending: null, state: { phase: "blocked", notice: "unauthenticated" } });
        } else if (kind === "operation-in-progress") commit({ state: { phase: "unknown", notice: "busy" } });
        else commit({ state: { phase: "unknown", notice: "unavailable" } });
      }
    },
    [api, apply, commit],
  );

  const execute = useCallback(
    async (pending: ReprintPending) => {
      // POST の直前に最新の権限を確認する。ここまで POST は送っていない（呼び出し元は未送信か
      // 「成功なし」確定後の再試行だけ）ため、拒否時は key を捨てて新しい操作に戻れる。
      if (!getCanEdit()) {
        commit({ pending: null, state: { phase: "blocked", notice: "forbidden" } });
        return;
      }
      commit({ pending, layout: pending.request.layout, state: { phase: "running" } });
      try {
        const result = await api.reprint(pending.editionId, pending.request);
        if (result.status === "not-succeeded") {
          commit({ state: { phase: "unknown", notice: "unavailable" } });
          await lookup(pending);
          return;
        }
        apply(result);
      } catch (error) {
        const decision = decideIssueFailure(error);
        if (decision.type === "stale") {
          commit({ pending: null, state: { phase: "stale", notice: decision.notice } });
          onRefetch();
        } else if (decision.type === "blocked") {
          commit({ pending: null, state: { phase: "blocked", notice: decision.notice } });
        } else {
          commit({ state: { phase: "unknown", notice: decision.notice } });
          await lookup(pending);
        }
      }
    },
    [api, apply, commit, getCanEdit, lookup, onRefetch],
  );

  const start = useCallback(
    async (layout: CertificateLayout) => {
      // 結果不明の操作が残っている間は新しい key を作らない（同 key の照会・再試行だけ）。
      if (store.get(editionKey).pending !== null) return;
      await execute({
        editionId: String(editionKey),
        request: {
          operation_key: newOperationKey(),
          layout,
          preconditions: { source_version: sourceVersion, latest_edition_id: editionKey },
        },
      });
    },
    [store, editionKey, sourceVersion, execute],
  );

  const recheck = useCallback(async () => {
    const pending = store.get(editionKey).pending;
    if (pending) await lookup(pending);
  }, [store, editionKey, lookup]);

  const retrySameKey = useCallback(async () => {
    const pending = store.get(editionKey).pending;
    if (pending) await execute(pending);
  }, [store, editionKey, execute]);

  /** 「成功なし」が確定した操作を職員の明示操作で破棄する（向き変更・別の再印刷へ進める）。 */
  const discard = useCallback((): boolean => {
    if (store.get(editionKey).state.phase !== "not-issued") return false;
    commit({ pending: null, state: { phase: "idle" } });
    return true;
  }, [store, editionKey, commit]);

  /** 画面上の表示を初期化する。未確定の key があるときは拒否して false を返す。 */
  const reset = useCallback((): boolean => {
    if (store.get(editionKey).pending !== null) return false;
    commit({ state: { phase: "idle" } });
    return true;
  }, [store, editionKey, commit]);

  return { state, activeLayout, start, recheck, retrySameKey, discard, reset };
}
