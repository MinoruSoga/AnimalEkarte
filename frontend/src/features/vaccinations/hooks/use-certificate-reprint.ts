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
        else if (kind === "forbidden") setState({ phase: "blocked", notice: "forbidden" });
        else if (kind === "unauthenticated") setState({ phase: "blocked", notice: "unauthenticated" });
        else if (kind === "operation-in-progress") setState({ phase: "unknown", notice: "busy" });
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
          setState({ phase: "stale", notice: decision.notice });
          onRefetch();
        } else if (decision.type === "blocked") {
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

  const reset = useCallback(() => {
    pendingRef.current = null;
    setState({ phase: "idle" });
  }, []);

  return { state, start, recheck, retrySameKey, reset };
}
