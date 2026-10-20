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

/** Blob.text() が無い環境(jsdom 等)でも読めるよう FileReader へフォールバックする。 */
async function readBlobText(blob: Blob): Promise<string> {
  if (typeof blob.text === "function") {
    try {
      return await blob.text();
    } catch {
      // FileReader へフォールバック
    }
  }
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result ?? ""));
    reader.onerror = () => reject(reader.error ?? new Error("blob read failed"));
    reader.readAsText(blob);
  });
}

/** blob 要求の失敗 body は Blob で届く。text → JSON に戻してから分類する。 */
async function readErrorBody(data: unknown): Promise<ErrorBody | null> {
  if (typeof Blob !== "undefined" && data instanceof Blob) {
    try {
      return asErrorBody(JSON.parse(await readBlobText(data)));
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
    return classifyCertificateError(
      error.response.status,
      await readErrorBody(error.response.data),
    );
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
        axios.get<unknown>(`${BASE}/vaccination-certificate-editions/${seg(editionId)}`, {
          signal,
        }),
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
        axios.post<unknown>(
          `${BASE}/vaccination-certificate-editions/${seg(editionId)}/reprints`,
          body,
        ),
      );
      return parse(certificateResultSchema, data);
    },
    async getArtifactContent(artifactId, signal) {
      // 共通 axios（認証 refresh・CSRF・医院 header）を通したまま、この操作だけ Accept を明示する（EMR-525）。
      const { data, headers } = await call(() =>
        axios.get<Blob>(`${BASE}/vaccination-certificate-artifacts/${seg(artifactId)}/content`, {
          headers: { Accept: "application/pdf" },
          responseType: "blob",
          signal,
        }),
      );
      if (!(data instanceof Blob) || data.size === 0) {
        throw new CertificateApiError("malformed", { message: "PDF の取得結果が空です" });
      }
      // 200 でも HTML/JSON（プロキシ・認証エッジ・障害ページ）が届きうる。Content-Type が PDF で、
      // かつ本体が %PDF- で始まるものだけを表示・印刷の対象にする（EMR-525: Content-Type 不一致は契約不正）。
      const contentType = String(headers["content-type"] ?? data.type ?? "");
      if (!isPdfMediaType(contentType) || !(await hasPdfMagic(data))) {
        throw new CertificateApiError("malformed", { message: "PDF ではない応答を受け取りました" });
      }
      return isPdfMediaType(data.type) ? data : new Blob([data], { type: "application/pdf" });
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

function isPdfMediaType(type: string): boolean {
  return type.split(";")[0].trim().toLowerCase() === "application/pdf";
}

/** 先頭が %PDF- の Blob だけを PDF とみなす（Content-Type の検査と併用する）。 */
async function hasPdfMagic(blob: Blob): Promise<boolean> {
  const head = await readHead(blob.slice(0, PDF_MAGIC.length));
  return PDF_MAGIC.every((byte, i) => head[i] === byte);
}

const PDF_MAGIC = [0x25, 0x50, 0x44, 0x46, 0x2d]; // "%PDF-"

function readHead(part: Blob): Promise<Uint8Array> {
  if (typeof part.arrayBuffer === "function") {
    return part.arrayBuffer().then((buffer) => new Uint8Array(buffer));
  }
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(new Uint8Array(reader.result as ArrayBuffer));
    reader.onerror = () => reject(reader.error);
    reader.readAsArrayBuffer(part);
  });
}

/** 実 API の既定インスタンス。 */
export const httpCertificateApi: CertificateApi = createHttpCertificateApi();
