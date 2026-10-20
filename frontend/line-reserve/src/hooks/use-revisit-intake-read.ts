// EMR-474: GET …/revisit-intake の取得・再取得・履歴の追加読込。
// 再取得中も前回の表示を保持し（入力 draft を消さない）、refetching で読込中を示す。
import { useCallback, useEffect, useState, useTransition } from "react";
import { liffApi } from "../api/liff-api";
import { resolveFetchError, type FetchErrorInfo } from "@/shared-liff/handle-fetch-error";
import type { RevisitIntakeHistoryEntry, RevisitIntakeLIFFRead } from "../types/models";
import { revisitIntakeRetryAfterSeconds } from "../lib/revisit-intake-submit";

/** 読込エラー。429 / Retry-After 付き 503 では再試行可能になる期限を持つ。 */
export interface RevisitIntakeReadError extends FetchErrorInfo {
  retryAfterSeconds: number | null;
  retryAvailableAt: number | null;
}

const READ_CONTEXT = "問診の取得";

interface ReadBackoff {
  until: number;
  seconds: number;
}

// route（clinic + 予約）別の Retry-After 期限。画面を閉じて開き直しても期限前に GET を出さないよう、
// hook ではなく module の memory に持つ（localStorage 等へは保存しない。LINE 終了で消える）。
const readBackoffs = new Map<string, ReadBackoff>();

function backoffKey(clinicId: string, appointmentId: string): string {
  return `${clinicId}|${appointmentId}`;
}

function activeBackoff(key: string): ReadBackoff | null {
  const stored = readBackoffs.get(key);
  if (stored === undefined) return null;
  if (Date.now() >= stored.until) {
    readBackoffs.delete(key);
    return null;
  }
  return stored;
}

/** テスト用: module memory の期限をすべて破棄する。 */
export function clearRevisitIntakeReadBackoffs(): void {
  readBackoffs.clear();
}

interface ReadState {
  read: RevisitIntakeLIFFRead | null;
  error: RevisitIntakeReadError | null;
  loadedToken: number;
}

export interface UseRevisitIntakeReadResult {
  read: RevisitIntakeLIFFRead | null;
  error: RevisitIntakeReadError | null;
  /** Retry-After の期限前は「再試行」を無効にする */
  retryBlocked: boolean;
  loading: boolean;
  refetching: boolean;
  refetch: () => void;
  history: RevisitIntakeHistoryEntry[];
  nextCursor: string | null;
  loadMore: () => void;
  loadingMore: boolean;
  loadMoreError: string | null;
  /** 「さらに表示」の 429 / Retry-After 付き 503 の待ち秒（期限中だけ非 null） */
  loadMoreRetryAfterSeconds: number | null;
}

function sameTargetIdentity(a: RevisitIntakeLIFFRead, b: RevisitIntakeLIFFRead): boolean {
  const x = a.current_target?.ref ?? null;
  const y = b.current_target?.ref ?? null;
  if (x === null || y === null) return x === y && a.availability === b.availability;
  return (
    x.appointment_id === y.appointment_id &&
    x.pet_id === y.pet_id &&
    x.target_generation === y.target_generation
  );
}

export function useRevisitIntakeRead(
  clinicId: string,
  appointmentId: string,
  idToken: string,
): UseRevisitIntakeReadResult {
  const [requestToken, setRequestToken] = useState(0);
  const [state, setState] = useState<ReadState>({ read: null, error: null, loadedToken: -1 });
  const [more, setMore] = useState<{
    forRead: RevisitIntakeLIFFRead | null;
    entries: RevisitIntakeHistoryEntry[];
    nextCursor: string | null;
    error: string | null;
  }>({ forRead: null, entries: [], nextCursor: null, error: null });
  const [loadingMore, startLoadMore] = useTransition();
  // route 共通の Retry-After 期限。読込の再試行と「さらに表示」の両方をこの期限まで止める。
  const routeKey = backoffKey(clinicId, appointmentId);
  const [backoff, setBackoffState] = useState<ReadBackoff | null>(() => activeBackoff(routeKey));
  const setBackoff = useCallback(
    (next: ReadBackoff | null) => {
      if (next === null) readBackoffs.delete(routeKey);
      else readBackoffs.set(routeKey, next);
      setBackoffState(next);
    },
    [routeKey],
  );

  // 期限中に要求された再取得（送信成功後・INTAKE_CLOSED 後など）は捨てずに保留し、期限が来たら実行する。
  // 保留中は refetching として扱い、editor は古い snapshot のまま送信できない。
  const [refetchQueued, setRefetchQueued] = useState(false);

  useEffect(() => {
    if (backoff === null) return;
    const timer = setTimeout(
      () => {
        setBackoff(null);
        if (refetchQueued) {
          setRefetchQueued(false);
          setState((prev) => ({ ...prev, error: null }));
          setRequestToken((t) => t + 1);
        }
      },
      Math.max(0, backoff.until - Date.now()),
    );
    return () => clearTimeout(timer);
  }, [backoff, refetchQueued, setBackoff]);

  useEffect(() => {
    let cancelled = false;
    // 開き直し直後でも、保存された期限前には GET を出さない（期限後に取得する）。
    const waitMs = Math.max(0, (activeBackoff(routeKey)?.until ?? 0) - Date.now());
    const request = () => {
      liffApi
        .getRevisitIntake(clinicId, appointmentId, idToken)
        .then((read) => {
          if (cancelled) return;
          setState({ read, error: null, loadedToken: requestToken });
        })
        .catch((err: unknown) => {
          if (cancelled) return;
          const retryAfterSeconds = revisitIntakeRetryAfterSeconds(err);
          const retryAvailableAt =
            retryAfterSeconds !== null ? Date.now() + retryAfterSeconds * 1000 : null;
          const info = resolveFetchError(err, READ_CONTEXT);
          // 401 では保持中の本文・履歴を出し続けない（403/404 と同じく保護された snapshot を捨てる）。
          setState((prev) => ({
            read: info.status === 401 ? null : prev.read,
            error: { ...info, retryAfterSeconds, retryAvailableAt },
            loadedToken: requestToken,
          }));
          if (retryAfterSeconds !== null && retryAvailableAt !== null) {
            setBackoff({ until: retryAvailableAt, seconds: retryAfterSeconds });
          }
        });
    };
    const timer = waitMs > 0 ? setTimeout(request, waitMs) : null;
    if (timer === null) request();
    return () => {
      cancelled = true;
      if (timer !== null) clearTimeout(timer);
    };
  }, [clinicId, appointmentId, idToken, requestToken, routeKey, setBackoff]);

  const refetch = useCallback(() => {
    if (backoff !== null && Date.now() < backoff.until) {
      setRefetchQueued(true);
      return;
    }
    setState((prev) => ({ ...prev, error: null }));
    setRequestToken((t) => t + 1);
  }, [backoff]);

  const read = state.read;
  // 追加ページは取得元の read と対応しているときだけ使う（再取得で1ページ目から読み直す）。
  const extra = more.forRead === read ? more : null;
  const history = read ? [...read.history, ...(extra?.entries ?? [])] : [];
  const nextCursor = extra ? extra.nextCursor : (read?.next_cursor ?? null);

  const loadMore = useCallback(() => {
    if (!read || nextCursor === null) return;
    if (backoff !== null && Date.now() < backoff.until) return;
    const baseRead = read;
    const cursor = nextCursor;
    startLoadMore(async () => {
      try {
        const page = await liffApi.getRevisitIntake(clinicId, appointmentId, idToken, { cursor });
        if (!sameTargetIdentity(baseRead, page)) {
          // 最初の読込の後に target が変わった。旧 read へ追記せず捨て、1ページ目から読み直す
          // （再取得が走るまで refetching になり editor は送信できない）。
          setState((prev) => ({ ...prev, error: null }));
          setRequestToken((t) => t + 1);
          return;
        }
        setMore((prev) => ({
          forRead: baseRead,
          entries: [...(prev.forRead === baseRead ? prev.entries : []), ...page.history],
          nextCursor: page.next_cursor,
          error: null,
        }));
      } catch (err: unknown) {
        const info = resolveFetchError(err, "履歴の取得");
        const retryAfter = revisitIntakeRetryAfterSeconds(err);
        if (retryAfter !== null) {
          setBackoff({ until: Date.now() + retryAfter * 1000, seconds: retryAfter });
        }
        // 401 / 403 / 404 では保持中の本文・履歴を捨てる（403/404 はページが ErrorPage へ替わる）。
        if (info.status === 401 || info.status === 403 || info.status === 404) {
          setState((prev) => ({
            read: null,
            error: {
              ...resolveFetchError(err, READ_CONTEXT),
              retryAfterSeconds: null,
              retryAvailableAt: null,
            },
            loadedToken: prev.loadedToken,
          }));
          return;
        }
        setMore((prev) => ({
          ...(prev.forRead === baseRead
            ? prev
            : { forRead: baseRead, entries: [], nextCursor: cursor }),
          error: info.message,
        }));
      }
    });
  }, [read, nextCursor, backoff, clinicId, appointmentId, idToken, setBackoff]);

  return {
    read,
    error: state.error,
    retryBlocked: backoff !== null,
    loading: state.read === null && state.error === null,
    refetching: (state.loadedToken !== requestToken || refetchQueued) && state.read !== null,
    refetch,
    history,
    nextCursor,
    loadMore,
    loadingMore,
    loadMoreError: extra?.error ?? null,
    loadMoreRetryAfterSeconds: backoff?.seconds ?? null,
  };
}

/** 403 / 404 は ErrorPage で扱う（receipt・回答の有無を推測しない）。 */
export function isRevisitIntakeAccessDenied(error: FetchErrorInfo | null): boolean {
  return error?.status === 403 || error?.status === 404;
}
