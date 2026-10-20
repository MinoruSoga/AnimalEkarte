// EMR-474: GET …/revisit-intake の取得・再取得・履歴の追加読込。
// 再取得中も前回の表示を保持し（入力 draft を消さない）、refetching で読込中を示す。
import { useCallback, useEffect, useState, useTransition } from "react";
import { liffApi } from "../api/liff-api";
import { resolveFetchError, type FetchErrorInfo } from "@/shared-liff/handle-fetch-error";
import type { RevisitIntakeHistoryEntry, RevisitIntakeLIFFRead } from "../types/models";

const READ_CONTEXT = "問診の取得";

interface ReadState {
  read: RevisitIntakeLIFFRead | null;
  error: FetchErrorInfo | null;
  loadedToken: number;
}

export interface UseRevisitIntakeReadResult {
  read: RevisitIntakeLIFFRead | null;
  error: FetchErrorInfo | null;
  loading: boolean;
  refetching: boolean;
  refetch: () => void;
  history: RevisitIntakeHistoryEntry[];
  nextCursor: string | null;
  loadMore: () => void;
  loadingMore: boolean;
  loadMoreError: string | null;
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

  useEffect(() => {
    let cancelled = false;
    liffApi
      .getRevisitIntake(clinicId, appointmentId, idToken)
      .then((read) => {
        if (cancelled) return;
        setState({ read, error: null, loadedToken: requestToken });
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setState((prev) => ({
          read: prev.read,
          error: resolveFetchError(err, READ_CONTEXT),
          loadedToken: requestToken,
        }));
      });
    return () => {
      cancelled = true;
    };
  }, [clinicId, appointmentId, idToken, requestToken]);

  const refetch = useCallback(() => {
    setState((prev) => ({ ...prev, error: null }));
    setRequestToken((t) => t + 1);
  }, []);

  const read = state.read;
  // 追加ページは取得元の read と対応しているときだけ使う（再取得で1ページ目から読み直す）。
  const extra = more.forRead === read ? more : null;
  const history = read ? [...read.history, ...(extra?.entries ?? [])] : [];
  const nextCursor = extra ? extra.nextCursor : (read?.next_cursor ?? null);

  const loadMore = useCallback(() => {
    if (!read || nextCursor === null) return;
    const baseRead = read;
    const cursor = nextCursor;
    startLoadMore(async () => {
      try {
        const page = await liffApi.getRevisitIntake(clinicId, appointmentId, idToken, { cursor });
        setMore((prev) => ({
          forRead: baseRead,
          entries: [...(prev.forRead === baseRead ? prev.entries : []), ...page.history],
          nextCursor: page.next_cursor,
          error: null,
        }));
      } catch (err: unknown) {
        const info = resolveFetchError(err, "履歴の取得");
        setMore((prev) => ({
          ...(prev.forRead === baseRead
            ? prev
            : { forRead: baseRead, entries: [], nextCursor: cursor }),
          error: info.message,
        }));
      }
    });
  }, [read, nextCursor, clinicId, appointmentId, idToken]);

  return {
    read,
    error: state.error,
    loading: state.read === null && state.error === null,
    refetching: state.loadedToken !== requestToken && state.read !== null,
    refetch,
    history,
    nextCursor,
    loadMore,
    loadingMore,
    loadMoreError: extra?.error ?? null,
  };
}

/** 403 / 404 は ErrorPage で扱う（receipt・回答の有無を推測しない）。 */
export function isRevisitIntakeAccessDenied(error: FetchErrorInfo | null): boolean {
  return error?.status === 403 || error?.status === 404;
}
