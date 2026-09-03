import axios, { type InternalAxiosRequestConfig } from 'axios';
import { z } from 'zod';
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
} from '../types/models';
import { API_BASE_URL } from '../lib/liff-config';
import { sanitizeNullBytes } from '@/lib/sanitize';
import { devError } from '@/shared-liff/dev-log';
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
} from './schemas';

const httpClient = axios.create({
  baseURL: API_BASE_URL,
});

// R-F20: BUG-067 と同一クラスの障害を防ぐため、POST/PATCH/PUT のリクエストボディから
// NULL バイトを除去する（メインアプリの axios.ts と同じサニタイズロジックを共有利用）。
function sanitizeRequestBody(config: InternalAxiosRequestConfig): InternalAxiosRequestConfig {
  const method = config.method?.toLowerCase();
  if ((method === 'post' || method === 'patch' || method === 'put') && config.data) {
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

export const liffApi = {
  getSettings: async (clinicId: string): Promise<LiffSettings> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/settings`);
    return parseOrThrow(liffSettingsSchema, res.data, 'getSettings');
  },

  getProfile: async (clinicId: string, idToken: string): Promise<LiffProfile> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/profile`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(liffProfileSchema, res.data, 'getProfile');
  },

  getCourses: async (clinicId: string, idToken: string): Promise<Course[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/courses`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(z.array(courseSchema), res.data, 'getCourses');
  },

  getTrimmingCourses: async (clinicId: string, idToken: string): Promise<TrimmingCourse[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/trimming-courses`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(z.array(trimmingCourseSchema), res.data, 'getTrimmingCourses');
  },

  getTrimmingOptions: async (clinicId: string, idToken: string): Promise<TrimmingOption[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/trimming-options`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(z.array(trimmingOptionSchema), res.data, 'getTrimmingOptions');
  },

  getStaffs: async (clinicId: string, courseId: number, idToken: string): Promise<Staff[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/staffs`, {
      headers: authHeaders(idToken),
      params: { courseId },
    });
    return parseOrThrow(z.array(staffSchema), res.data, 'getStaffs');
  },

  getAvailableDates: async (
    clinicId: string,
    courseId: number,
    staffId: number,
    idToken: string,
  ): Promise<AvailableDate[]> => {
    const res = await httpClient.get<unknown>(
      `${clinicPath(clinicId)}/available-dates`,
      {
        headers: authHeaders(idToken),
        params: { courseId, staffId },
      },
    );
    return parseOrThrow(availableDatesResponseSchema, res.data, 'getAvailableDates').dates;
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
    return parseOrThrow(z.array(availableTimeSchema), res.data, 'getAvailableTimes');
  },

  createReservation: async (
    clinicId: string,
    body: CreateReservationBody,
    idToken: string,
  ): Promise<CreateReservationResponse> => {
    const res = await httpClient.post<unknown>(`${clinicPath(clinicId)}/reservations`, body, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(createReservationResponseSchema, res.data, 'createReservation');
  },

  getMyReservations: async (clinicId: string, idToken: string): Promise<Reservation[]> => {
    const res = await httpClient.get<unknown>(`${clinicPath(clinicId)}/my-reservations`, {
      headers: authHeaders(idToken),
    });
    return parseOrThrow(z.array(reservationSchema), res.data, 'getMyReservations');
  },

  cancelReservation: async (clinicId: string, id: number, idToken: string): Promise<void> => {
    await httpClient.delete(`${clinicPath(clinicId)}/my-reservations/${id}`, {
      headers: authHeaders(idToken),
    });
  },
};
