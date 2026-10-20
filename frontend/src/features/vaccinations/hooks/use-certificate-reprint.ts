import { useCallback, useRef, useState } from "react";
import type { CertificateApi, CertificateReprintRequest } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import { newOperationKey } from "../api/certificate-operation-key";
import {
  readSucceededRefs,
  type CertificateEdition,
  type CertificateLayout,
  type CertificateResult,
} from "../api/certificate-schemas";
import { decideIssueFailure, type IssueNoticeCode } from "../lib/certificate-issue-reducer";

/**
 * 保存版の再印刷（POST /reprints）。issue と同じ operation_key 規則:
 * 結果不明では同 key で照会し、新しい key は「成功なし」確定後の新しい操作でだけ作る。
 */
/** 結果が確定していない間（同じ key で照会・再試行するだけの状態）。 */
export function isReprintUnresolved(state: ReprintState): boolean {
  return state.phase === "running" || state.phase === "unknown" || state.phase === "not-issued";
}

export type ReprintState =
  | { phase: "idle" }
  | { phase: "running" }
  | { phase: "unknown"; notice: IssueNoticeCode }
  | { phase: "not-issued"; notice: IssueNoticeCode }
  | { phase: "stale"; notice: IssueNoticeCode }
  | { phase: "blocked"; notice: IssueNoticeCode }
  | { phase: "done"; editionId: string; artifactId: string; replayed: boolean };

interface Pending {
  readonly editionId: string;
  readonly request: CertificateReprintRequest;
}

export function useCertificateReprint(args: {
  api: CertificateApi;
  edition: CertificateEdition;
  sourceVersion: number;
  onRefetch: () => void;
}) {
  const { api, edition, sourceVersion, onRefetch } = args;
  const [state, setState] = useState<ReprintState>({ phase: "idle" });
  const pendingRef = useRef<Pending | null>(null);

  const apply = useCallback(
    (result: CertificateResult) => {
      if (result.status === "not-succeeded") {
        setState({ phase: "not-issued", notice: "not-succeeded" });
        return;
      }
      const refs = readSucceededRefs(result);
      if (refs === null) {
        setState({ phase: "unknown", notice: "contract-violation" });
        return;
      }
      pendingRef.current = null;
      setState({ phase: "done", ...refs, replayed: result.status === "replayed" });
    },
    [],
  );

  const lookup = useCallback(
    async (pending: Pending) => {
      try {
        apply(await api.lookupOperation(pending.request.operation_key, "reprint"));
      } catch (error) {
        const kind = error instanceof CertificateApiError ? error.kind : "network";
        if (kind === "scope-not-found") setState({ phase: "not-issued", notice: "not-claimed" });
        else if (kind === "forbidden") {
          pendingRef.current = null;
          setState({ phase: "blocked", notice: "forbidden" });
        } else if (kind === "unauthenticated") {
          pendingRef.current = null;
          setState({ phase: "blocked", notice: "unauthenticated" });
        } else if (kind === "operation-in-progress") setState({ phase: "unknown", notice: "busy" });
        else setState({ phase: "unknown", notice: "unavailable" });
      }
    },
    [api, apply],
  );

  const execute = useCallback(
    async (pending: Pending) => {
      setState({ phase: "running" });
      try {
        const result = await api.reprint(pending.editionId, pending.request);
        if (result.status === "not-succeeded") {
          setState({ phase: "unknown", notice: "unavailable" });
          await lookup(pending);
          return;
        }
        apply(result);
      } catch (error) {
        const decision = decideIssueFailure(error);
        if (decision.type === "stale") {
          pendingRef.current = null;
          setState({ phase: "stale", notice: decision.notice });
          onRefetch();
        } else if (decision.type === "blocked") {
          pendingRef.current = null;
          setState({ phase: "blocked", notice: decision.notice });
        } else {
          setState({ phase: "unknown", notice: decision.notice });
          await lookup(pending);
        }
      }
    },
    [api, apply, lookup, onRefetch],
  );

  const start = useCallback(
    async (layout: CertificateLayout) => {
      // 結果不明の操作が残っている間は新しい key を作らない（同 key の照会・再試行だけ）。
      if (pendingRef.current !== null) return;
      const pending: Pending = {
        editionId: String(edition.id),
        request: {
          operation_key: newOperationKey(),
          layout,
          preconditions: { source_version: sourceVersion, latest_edition_id: edition.id },
        },
      };
      pendingRef.current = pending;
      await execute(pending);
    },
    [edition.id, sourceVersion, execute],
  );

  const recheck = useCallback(async () => {
    if (pendingRef.current) await lookup(pendingRef.current);
  }, [lookup]);

  const retrySameKey = useCallback(async () => {
    if (pendingRef.current) await execute(pendingRef.current);
  }, [execute]);

  /** 画面上の表示を初期化する。未確定の key があるときは拒否して false を返す。 */
  const reset = useCallback((): boolean => {
    if (pendingRef.current !== null) return false;
    setState({ phase: "idle" });
    return true;
  }, []);

  return { state, start, recheck, retrySameKey, reset };
}
