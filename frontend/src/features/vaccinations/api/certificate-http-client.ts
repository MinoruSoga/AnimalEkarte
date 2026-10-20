import { isAxiosError } from "axios";
import type { z } from "zod";
import { axios } from "@/lib/axios";
import type { CertificateApi } from "./certificate-api";
import {
  CertificateApiError,
  classifyCertificateError,
  type ErrorBody,
} from "./certificate-errors";
import {
  certificateContextSchema,
  certificateEditionPageSchema,
  certificateEditionSchema,
  certificatePreviewResponseSchema,
  certificateResultSchema,
  certificateSavedPreviewSchema,
} from "./certificate-schemas";

/**
 * EMR-487: 実 API クライアント（`/api/v1` 配下。axios baseURL は `/api`）。
 * 契約は selected / runtime 未実装。ここにモックへのフォールバックは無い。
 * 失敗は CertificateApiError に分類して投げ、呼び出し側の状態遷移に委ねる。
 */

const BASE = "/v1";

function asErrorBody(value: unknown): ErrorBody | null {
  if (typeof value !== "object" || value === null) return null;
  const rec = value as Record<string, unknown>;
  const pick = (k: string) => (typeof rec[k] === "string" ? (rec[k] as string) : undefined);
  return {
    error: pick("error"),
    code: pick("code"),
    status: pick("status"),
    operation_key: pick("operation_key"),
  };
}

/** blob 要求の失敗 body は Blob で届く。text → JSON に戻してから分類する。 */
async function readErrorBody(data: unknown): Promise<ErrorBody | null> {
  if (typeof Blob !== "undefined" && data instanceof Blob) {
    try {
      return asErrorBody(JSON.parse(await data.text()));
    } catch {
      return null;
    }
  }
  return asErrorBody(data);
}

async function toApiError(error: unknown): Promise<CertificateApiError> {
  if (error instanceof CertificateApiError) return error;
  if (isAxiosError(error)) {
    if (error.response === undefined) return classifyCertificateError(undefined, null);
    return classifyCertificateError(error.response.status, await readErrorBody(error.response.data));
  }
  return new CertificateApiError("network");
}

function parse<S extends z.ZodType>(schema: S, data: unknown): z.infer<S> {
  const result = schema.safeParse(data);
  if (!result.success) {
    // 契約外の応答は原文を出さず fail-closed（PII を例外メッセージへ載せない）。
    throw new CertificateApiError("malformed", { message: "証明書 API の応答形式が不正です" });
  }
  return result.data;
}

async function call<T>(run: () => Promise<T>): Promise<T> {
  try {
    return await run();
  } catch (error) {
    throw await toApiError(error);
  }
}

const seg = encodeURIComponent;

export function createHttpCertificateApi(): CertificateApi {
  return {
    async getContext(vaccinationId, signal) {
      const { data } = await call(() =>
        axios.get<unknown>(`${BASE}/vaccinations/${vaccinationId}/certificate-context`, { signal }),
      );
      return parse(certificateContextSchema, data);
    },
    async previewForIssue(vaccinationId, body, signal) {
      const { data } = await call(() =>
        axios.post<unknown>(`${BASE}/vaccinations/${vaccinationId}/certificate-preview`, body, {
          signal,
        }),
      );
      return parse(certificatePreviewResponseSchema, data);
    },
    async issue(vaccinationId, body) {
      const { data } = await call(() =>
        axios.post<unknown>(`${BASE}/vaccinations/${vaccinationId}/certificate-editions`, body),
      );
      return parse(certificateResultSchema, data);
    },
    async getEdition(editionId, signal) {
      const { data } = await call(() =>
        axios.get<unknown>(`${BASE}/vaccination-certificate-editions/${seg(editionId)}`, { signal }),
      );
      return parse(certificateEditionSchema, data);
    },
    async listEditions(vaccinationId, query, signal) {
      const { data } = await call(() =>
        axios.get<unknown>(`${BASE}/vaccinations/${vaccinationId}/certificate-editions`, {
          params: { kind: query.kind, page: query.page ?? 1, limit: query.limit ?? 20 },
          signal,
        }),
      );
      return parse(certificateEditionPageSchema, data);
    },
    async previewSaved(editionId, layout, signal) {
      const { data } = await call(() =>
        axios.post<unknown>(
          `${BASE}/vaccination-certificate-editions/${seg(editionId)}/preview`,
          { layout },
          { signal },
        ),
      );
      return parse(certificateSavedPreviewSchema, data);
    },
    async reprint(editionId, body) {
      const { data } = await call(() =>
        axios.post<unknown>(`${BASE}/vaccination-certificate-editions/${seg(editionId)}/reprints`, body),
      );
      return parse(certificateResultSchema, data);
    },
    async getArtifactContent(artifactId, signal) {
      // 共有 axios は Accept を application/json に固定する。PDF 応答は BE が Content-Type で返す前提
      // （Accept 上書きの要否は BE 実装時の確認事項）。
      const { data } = await call(() =>
        axios.get<Blob>(`${BASE}/vaccination-certificate-artifacts/${seg(artifactId)}/content`, {
          responseType: "blob",
          signal,
        }),
      );
      if (!(data instanceof Blob) || data.size === 0) {
        throw new CertificateApiError("malformed", { message: "PDF の取得結果が空です" });
      }
      return data;
    },
    async lookupOperation(operationKey, intent, signal) {
      const { data } = await call(() =>
        axios.get<unknown>(`${BASE}/vaccination-certificate-operations/${seg(operationKey)}`, {
          params: { intent },
          signal,
        }),
      );
      return parse(certificateResultSchema, data);
    },
  };
}

/** 実 API の既定インスタンス。 */
export const httpCertificateApi: CertificateApi = createHttpCertificateApi();
