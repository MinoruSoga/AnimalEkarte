import axios, { type InternalAxiosRequestConfig } from "axios";
import { z } from "zod";
import type {
  LiffSettings,
  LiffProfile,
  Course,
  Staff,
  AvailableDate,
  AvailableTime,
  Reservation,
  CreateReservationBody,
  CreateReservationResponse,
  TrimmingCourse,
  TrimmingOption,
  RevisitIntakeLIFFRead,
  RevisitIntakeSubmitRequest,
  RevisitIntakeResultResponse,
  RevisitIntakeLookupResult,
} from "../types/models";
import { API_BASE_URL } from "../lib/liff-config";
import { sanitizeNullBytes } from "@/shared-liff/sanitize";
import { devError } from "@/shared-liff/dev-log";
import {
  liffSettingsSchema,
  liffProfileSchema,
  courseSchema,
  trimmingCourseSchema,
  trimmingOptionSchema,
  staffSchema,
  availableDatesResponseSchema,
  availableTimeSchema,
  reservationSchema,
  createReservationResponseSchema,
  revisitIntakeLIFFReadSchema,
  revisitIntakeResultResponseSchema,
  revisitIntakeLookupNotFoundSchema,
} from "./schemas";

const httpClient = axios.create({
  baseURL: API_BASE_URL,
});

// R-F20: BUG-067 と同一クラスの障害を防ぐため、POST/PATCH/PUT のリクエストボディから
// NULL バイトを除去する（メインアプリの axios.ts と同じサニタイズロジックを共有利用）。
function sanitizeRequestBody(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig {
  const method = config.method?.toLowerCase();
  if ((method === "post" || method === "patch" || method === "put") && config.data) {
    config.data = sanitizeNullBytes(config.data);
  }
  return config;
}

httpClient.interceptors.request.use(sanitizeRequestBody);

function authHeaders(idToken: string) {
  return { Authorization: `Bearer ${idToken}` };
}

/**
 * FE-RC-079: clinicId は URL パスセグメントとして使われるため encodeURIComponent が必須。
 * liff/src/api/liff-api.ts と同じ規約（各呼び出しで encode）だが、こちらは呼び出し箇所が
 * 多いため共通ヘルパーに集約し、encode 漏れを構造的に防ぐ。
 */
function clinicPath(clinicId: string): string {
  return `/api/liff/${encodeURIComponent(clinicId)}`;
}

// FE5-18: 成功パスのレスポンスを実行時検証する。失敗時はどの API 関数で壊れたかを
// メッセージに含めて throw する（liff/src/api/liff-api.ts の safeParse パターンを踏襲）。
function parseOrThrow<T>(schema: z.ZodType<T>, data: unknown, fnName: string): T {
  const parsed = schema.safeParse(data);
  if (!parsed.success) {
    devError(`[liffApi.${fnName}] invalid response shape:`, parsed.error);
    throw new Error(`${fnName} response validation failed`);
  }
  return parsed.data;
}

/**
 * EMR-474: 1 appointment = 1 pet の現在 target 固定を検査する（fail-closed）。要求した予約以外の
 * target・履歴、または現在 target と pet / 世代 / target_id が違う current 履歴を含む応答は契約エラーとして投げる。
 */
function assertRevisitIntakeReadMatchesAppointment(
  read: RevisitIntakeLIFFRead,
  appointmentId: string,
): void {
  const current = read.current_target;
  const mismatch =
    (current !== null && current.ref.appointment_id !== appointmentId) ||
    read.history.some(
      (h) =>
        h.target.appointment_id !== appointmentId ||
        (h.target_relation === "current" &&
          (current === null ||
            h.target.pet_id !== current.ref.pet_id ||
            current.ref.target_id === null ||
            h.target.target_id !== current.ref.target_id ||
            h.target.target_generation !== current.ref.target_generation)),
    );
  if (mismatch) {
    devError("[liffApi.getRevisitIntake] response does not match the requested appointment");
    throw new Error("getRevisitIntake response does not match the requested appointment");
  }
}

export const liffApi = {
  getSettings: async (clinicId: string): Promise<LiffSettings> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/settings`);
    return parseOrThrow(liffSettingsSchema, res.data, "getSettings");
  },

  getProfile: async (clinicId: string, idToken: string): Promise<LiffProfile> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/profile`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(liffProfileSchema, res.data, "getProfile");
  },

  getCourses: async (clinicId: string, idToken: string): Promise<Course[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/courses`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(z.array(courseSchema), res.data, "getCourses");
  },

  getTrimmingCourses: async (clinicId: string, idToken: string): Promise<TrimmingCourse[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/trimming-courses`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(z.array(trimmingCourseSchema), res.data, "getTrimmingCourses");
  },

  getTrimmingOptions: async (clinicId: string, idToken: string): Promise<TrimmingOption[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/trimming-options`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(z.array(trimmingOptionSchema), res.data, "getTrimmingOptions");
  },

  getStaffs: async (clinicId: string, courseId: number, idToken: string): Promise<Staff[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/staffs`, {
      headers: authHeaders(idToken),
      params: { courseId },
    });
    return parseOrThrow(z.array(staffSchema), res.data, "getStaffs");
  },

  getAvailableDates: async (
    clinicId: string,
    courseId: number,
    staffId: number,
    idToken: string,
  ): Promise<AvailableDate[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/available-dates`, {
      headers: authHeaders(idToken),
      params: { courseId, staffId },
    });
    return parseOrThrow(availableDatesResponseSchema, res.data, "getAvailableDates").dates;
  },

  getAvailableTimes: async (
    clinicId: string,
    courseId: number,
    staffId: number,
    date: string,
    idToken: string,
  ): Promise<AvailableTime[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/available-times`, {
      headers: authHeaders(idToken),
      params: { courseId, staffId, date },
    });
    return parseOrThrow(z.array(availableTimeSchema), res.data, "getAvailableTimes");
  },

  createReservation: async (
    clinicId: string,
    body: CreateReservationBody,
    idToken: string,
  ): Promise<CreateReservationResponse> => {
    const res = await httpClient.post<unknown>(`${clinicPath(clinicId)}/reservations`, body, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(createReservationResponseSchema, res.data, "createReservation");
  },

  getMyReservations: async (clinicId: string, idToken: string): Promise<Reservation[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/my-reservations`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(z.array(reservationSchema), res.data, "getMyReservations");
  },

  cancelReservation: async (clinicId: string, id: number, idToken: string): Promise<void> => {
    await httpClient.delete(`${clinicPath(clinicId)}/my-reservations/${id}`, {
      headers: authHeaders(idToken),
    });
  },

  /**
   * EMR-474: GET /my-reservations/{appointmentId}/revisit-intake（RevisitIntakeLIFFRead）。
   * 未知 query は 400 になる契約のため cursor / limit 以外を送らない。
   */
  getRevisitIntake: async (
    clinicId: string,
    appointmentId: string,
    idToken: string,
    page: { cursor?: string; limit?: number } = {},
  ): Promise<RevisitIntakeLIFFRead> => {
    const params: Record<string, string | number> = {};
    if (page.cursor !== undefined) params.cursor = page.cursor;
    if (page.limit !== undefined) params.limit = page.limit;
    const res = await httpClient.get<unknown>(
      `${clinicPath(clinicId)}/my-reservations/${encodeURIComponent(appointmentId)}/revisit-intake`,
      { headers: authHeaders(idToken), params },
    );
    const read = parseOrThrow(revisitIntakeLIFFReadSchema, res.data, "getRevisitIntake");
    assertRevisitIntakeReadMatchesAppointment(read, appointmentId);
    return read;
  },

  /**
   * EMR-474: POST …/revisit-intake/submissions（Idempotency-Key 必須）。
   * 4xx/5xx は AxiosError のまま投げ、呼び出し側が HTTP status と code で分岐する。
   */
  submitRevisitIntake: async (
    clinicId: string,
    appointmentId: string,
    body: RevisitIntakeSubmitRequest,
    operationKey: string,
    idToken: string,
  ): Promise<RevisitIntakeResultResponse> => {
    const res = await httpClient.post<unknown>(
      `${clinicPath(clinicId)}/my-reservations/${encodeURIComponent(appointmentId)}/revisit-intake/submissions`,
      body,
      {
        headers: {
          ...authHeaders(idToken),
          "Content-Type": "application/json",
          "Idempotency-Key": operationKey,
        },
      },
    );
    return parseOrThrow(revisitIntakeResultResponseSchema, res.data, "submitRevisitIntake");
  },

  /**
   * EMR-474: GET /revisit-intake/operations/{operationKey}。
   * 404 RevisitIntakeLookupNotFound は失敗確定ではないため state=not_found の結果として返す。
   * それ以外の失敗（401/429/5xx/通信断/形不正）は投げる。
   */
  getRevisitIntakeOperation: async (
    clinicId: string,
    operationKey: string,
    idToken: string,
  ): Promise<RevisitIntakeLookupResult> => {
    try {
      const res = await httpClient.get<unknown>(
        `${clinicPath(clinicId)}/revisit-intake/operations/${encodeURIComponent(operationKey)}`,
        { headers: authHeaders(idToken) },
      );
      return parseOrThrow(revisitIntakeResultResponseSchema, res.data, "getRevisitIntakeOperation");
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 404) {
        return parseOrThrow(
          revisitIntakeLookupNotFoundSchema,
          err.response.data,
          "getRevisitIntakeOperation",
        );
      }
      throw err;
    }
  },
};
