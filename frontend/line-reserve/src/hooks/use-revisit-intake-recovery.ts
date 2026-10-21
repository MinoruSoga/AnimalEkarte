// EMR-474 operation-recovery Slice A: mount 時（と明示の再確認時）に未確定操作を発見する hook。
// 状態は React state（memory）だけ。localStorage / sessionStorage / IndexedDB / URL へ何も書かない。
//
// - target scope（現在 target の ref が分かるとき）。404 SCOPE_NOT_FOUND / 403 NOT_AUTHORIZED なら
//   self scope へ切り替える（失権した対象は存在非開示のため）。target が無い・読込失敗のときは最初から self。
// - 発見が終わるまで・失敗したときは新しい送信を止める（newSubmissionBlocked=true）。
//   0 件・complete・経過時間から「送信してよい」を導かない。許可を表す field はこの hook にない。
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  RevisitIntakeRecoveryQueryError,
  type RevisitIntakeRecoveryErrorKind,
} from "../api/revisit-intake-recovery";
import {
  createRecoveryPageFetcher,
  discoverRevisitIntakeRecovery,
  summarizeRevisitIntakeRecovery,
  type RecoveryPageFetcher,
  type RevisitIntakeRecoveryDiscovery,
  type RevisitIntakeRecoveryScope,
  type RevisitIntakeRecoverySummary,
} from "../lib/revisit-intake-recovery-discovery";

interface RevisitIntakeRecoveryTargetInput {
  appointmentId: string;
  petId: string;
  targetGeneration: string;
}

interface UseRevisitIntakeRecoveryArgs {
  clinicId: string;
  idToken: string;
  /** 現在 target の ref。未取得なら null（このとき selfScope=true でなければ発見を始めない）。 */
  target: RevisitIntakeRecoveryTargetInput | null;
  /** target を使えない（読込拒否・失敗・target 不明）ので self scope で発見する。 */
  selfScope: boolean;
  /** テスト用の注入口。既定は実 adapter（mock へのフォールバックはしない）。 */
  fetchPage?: RecoveryPageFetcher;
}

type RevisitIntakeRecoveryStatus = "idle" | "checking" | "done" | "error";

export interface UseRevisitIntakeRecoveryResult {
  status: RevisitIntakeRecoveryStatus;
  discovery: RevisitIntakeRecoveryDiscovery | null;
  summary: RevisitIntakeRecoverySummary | null;
  /** target scope が使えず self scope へ切り替えた。 */
  fellBackToSelf: boolean;
  errorKind: RevisitIntakeRecoveryErrorKind | null;
  /** 新しい送信を始めてはならない。idle / checking / error / 保留理由あり のとき true。 */
  newSubmissionBlocked: boolean;
  /** 同じ scope で発見をやり直す（結果の再確認後などに使う）。 */
  refresh: () => void;
}

interface Settled {
  runKey: string;
  scopeKey: string;
  outcome:
    | { ok: true; discovery: RevisitIntakeRecoveryDiscovery; fellBack: boolean }
    | { ok: false; kind: RevisitIntakeRecoveryErrorKind };
}

const FALLBACK_KINDS: ReadonlySet<RevisitIntakeRecoveryErrorKind> = new Set([
  "scope_not_found",
  "not_authorized",
]);

function toKind(err: unknown): RevisitIntakeRecoveryErrorKind {
  return err instanceof RevisitIntakeRecoveryQueryError ? err.kind : "unclassified";
}

export function useRevisitIntakeRecovery(
  args: UseRevisitIntakeRecoveryArgs,
): UseRevisitIntakeRecoveryResult {
  const { clinicId, idToken, target, selfScope, fetchPage } = args;
  const appointmentId = target?.appointmentId ?? null;
  const petId = target?.petId ?? null;
  const targetGeneration = target?.targetGeneration ?? null;
  const [refreshToken, setRefreshToken] = useState(0);
  const [settled, setSettled] = useState<Settled | null>(null);

  // 発見の対象。target が分かればそれ、なければ self（読込失敗・拒否・target 不明）、どちらも無ければ待機。
  const scope = useMemo<RevisitIntakeRecoveryScope | null>(() => {
    if (appointmentId !== null && petId !== null && targetGeneration !== null) {
      return {
        scope: "target",
        targetRef: {
          appointment_id: appointmentId,
          pet_id: petId,
          target_generation: targetGeneration,
        },
      };
    }
    return selfScope ? { scope: "self" } : null;
  }, [appointmentId, petId, targetGeneration, selfScope]);

  const scopeKey =
    scope === null
      ? null
      : scope.scope === "target"
        ? `${clinicId}|target|${scope.targetRef.appointment_id}|${scope.targetRef.pet_id}|${scope.targetRef.target_generation}`
        : `${clinicId}|self`;
  const runKey = scopeKey === null ? null : `${scopeKey}#${refreshToken}`;

  useEffect(() => {
    if (scope === null || scopeKey === null || runKey === null) return;
    const controller = new AbortController();
    const fetcher = fetchPage ?? createRecoveryPageFetcher(clinicId, idToken);
    const run = async (): Promise<Settled["outcome"]> => {
      try {
        return {
          ok: true,
          fellBack: false,
          discovery: await discoverRevisitIntakeRecovery(fetcher, scope, {
            signal: controller.signal,
          }),
        };
      } catch (err: unknown) {
        const kind = toKind(err);
        if (scope.scope === "target" && FALLBACK_KINDS.has(kind)) {
          try {
            return {
              ok: true,
              fellBack: true,
              discovery: await discoverRevisitIntakeRecovery(fetcher, { scope: "self" }, {
                signal: controller.signal,
              }),
            };
          } catch (selfErr: unknown) {
            return { ok: false, kind: toKind(selfErr) };
          }
        }
        return { ok: false, kind };
      }
    };
    void run().then((outcome) => {
      if (controller.signal.aborted) return;
      setSettled({ runKey, scopeKey, outcome });
    });
    return () => controller.abort();
  }, [scope, scopeKey, runKey, clinicId, idToken, fetchPage]);

  const refresh = useCallback(() => setRefreshToken((n) => n + 1), []);

  let status: RevisitIntakeRecoveryStatus = "idle";
  let discovery: RevisitIntakeRecoveryDiscovery | null = null;
  let fellBackToSelf = false;
  let errorKind: RevisitIntakeRecoveryErrorKind | null = null;
  if (runKey !== null) {
    if (settled !== null && settled.runKey === runKey) {
      if (settled.outcome.ok) {
        status = "done";
        discovery = settled.outcome.discovery;
        fellBackToSelf = settled.outcome.fellBack;
      } else {
        status = "error";
        errorKind = settled.outcome.kind;
      }
    } else {
      status = "checking";
      // 再確認中は同じ scope の前回結果を表示だけ残す（保留は解除しない）。
      if (settled !== null && settled.scopeKey === scopeKey && settled.outcome.ok) {
        discovery = settled.outcome.discovery;
        fellBackToSelf = settled.outcome.fellBack;
      }
    }
  }
  const summary = discovery !== null ? summarizeRevisitIntakeRecovery(discovery) : null;
  const newSubmissionBlocked =
    status !== "done" || summary === null || summary.newSubmissionBlocked;
  return {
    status,
    discovery,
    summary,
    fellBackToSelf,
    errorKind,
    newSubmissionBlocked,
    refresh,
  };
}
