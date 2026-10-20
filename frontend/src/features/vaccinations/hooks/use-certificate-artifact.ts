import { useCallback, useEffect, useState } from "react";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError, type CertificateErrorKind } from "../api/certificate-errors";

export type ArtifactContentState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ready"; blob: Blob }
  | { status: "error"; kind: CertificateErrorKind };

interface Settled {
  /** どの要求（api・artifact・試行回）の結果か。現在の要求と一致するときだけ採用する。 */
  readonly api: CertificateApi;
  readonly artifactId: string;
  readonly attempt: number;
  readonly outcome: { status: "ready"; blob: Blob } | { status: "error"; kind: CertificateErrorKind };
}

/**
 * 印刷に使う唯一の PDF（GET /vaccination-certificate-artifacts/{id}/content）を取得する。
 * PDF の Blob は react-query のキャッシュに載せず、このコンポーネントの state にだけ持つ。
 * 失敗しても別の PDF（preview など）へフォールバックしない。
 * 状態は「現在の要求」と「確定した結果」から導出し、effect 本体で同期的に setState しない。
 */
export function useCertificateArtifactContent(
  api: CertificateApi,
  artifactId: string | null,
  enabled: boolean,
) {
  const [settled, setSettled] = useState<Settled | null>(null);
  const [attempt, setAttempt] = useState(0);
  const active = enabled && artifactId !== null;

  useEffect(() => {
    if (!enabled || artifactId === null) return;
    const controller = new AbortController();
    api
      .getArtifactContent(artifactId, controller.signal)
      .then((blob) => {
        if (!controller.signal.aborted) {
          setSettled({ api, artifactId, attempt, outcome: { status: "ready", blob } });
        }
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted) return;
        setSettled({
          api,
          artifactId,
          attempt,
          outcome: { status: "error", kind: error instanceof CertificateApiError ? error.kind : "network" },
        });
      });
    return () => controller.abort();
  }, [api, enabled, artifactId, attempt]);

  let state: ArtifactContentState;
  if (!active) state = { status: "idle" };
  else if (settled && settled.api === api && settled.artifactId === artifactId && settled.attempt === attempt) {
    state = settled.outcome;
  } else state = { status: "loading" };

  const retry = useCallback(() => setAttempt((n) => n + 1), []);
  return { state, retry };
}
