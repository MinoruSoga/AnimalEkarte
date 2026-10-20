import { useCallback, useEffect, useState } from "react";
import type { CertificateApi } from "../api/certificate-api";
import { CertificateApiError, type CertificateErrorKind } from "../api/certificate-errors";

export type ArtifactContentState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ready"; blob: Blob }
  | { status: "error"; kind: CertificateErrorKind };

/**
 * 印刷に使う唯一の PDF（GET /vaccination-certificate-artifacts/{id}/content）を取得する。
 * PDF の Blob は react-query のキャッシュに載せず、このコンポーネントの state にだけ持つ。
 * 失敗しても別の PDF（preview など）へフォールバックしない。
 */
export function useCertificateArtifactContent(
  api: CertificateApi,
  artifactId: string | null,
  enabled: boolean,
) {
  const [state, setState] = useState<ArtifactContentState>({ status: "idle" });
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    if (!enabled || artifactId === null) {
      setState({ status: "idle" });
      return;
    }
    const controller = new AbortController();
    setState({ status: "loading" });
    api
      .getArtifactContent(artifactId, controller.signal)
      .then((blob) => {
        if (!controller.signal.aborted) setState({ status: "ready", blob });
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted) return;
        setState({
          status: "error",
          kind: error instanceof CertificateApiError ? error.kind : "network",
        });
      });
    return () => controller.abort();
  }, [api, artifactId, enabled, attempt]);

  const retry = useCallback(() => setAttempt((n) => n + 1), []);
  return { state, retry };
}
