import { addHours } from "date-fns";

import {
  buildJSTWallDateTime,
  formatJSTWallDate,
  isPastJSTDate,
  jstWallDateToISOString,
  todayJSTISO,
} from "@/lib/jst-date";
import type { Reservation } from "@/types";

/**
 * EMR-308: POST /v1/reservations/:id/confirm-tentative のリクエスト body。
 * tentative（次回予定・未予約）行を同一行のまま実予約へ昇格させる契約。
 * generated 型にはこのエンドポイント専用の DTO が無いため、共有モーダル側で
 * フォームデータ（Partial<Reservation>）→ request body の変換を一元管理する。
 * （feature 同士は相互 import できないため medical-records / reservations 双方が
 * ここを経由する。）
 */
export interface ConfirmTentativeRequestBody {
  /** 予約開始日時（ISO 8601。JST wall-clock Date → +09:00 変換済み） */
  start_time: string;
  /** 予約終了日時（ISO 8601） */
  end_time: string;
  reservation_type_id?: number;
  doctor_id?: number;
  is_designated?: boolean;
  /** "pending" | "confirmed" のみ許容（BE 契約） */
  status?: "pending" | "confirmed";
  notes?: string;
}

/**
 * tentative 確定フォームの入力値 → confirm-tentative リクエスト body。
 * start/end が未設定なら null（モーダルのバリデーションで弾かれる想定だが、
 * 呼び出し側の安全側境界として残す）。
 */
export function buildConfirmTentativeBody(
  data: Partial<Reservation>,
): ConfirmTentativeRequestBody | null {
  if (!data.start || !data.end) return null;

  const body: ConfirmTentativeRequestBody = {
    start_time: jstWallDateToISOString(data.start),
    end_time: jstWallDateToISOString(data.end),
  };

  const typeId = Number(data.type);
  if (Number.isSafeInteger(typeId) && typeId > 0) {
    body.reservation_type_id = typeId;
  }

  const doctorId = Number(data.doctor);
  if (Number.isSafeInteger(doctorId) && doctorId > 0) {
    body.doctor_id = doctorId;
  }

  if (data.isDesignated !== undefined) {
    body.is_designated = data.isDesignated;
  }

  // BE 契約は "pending" | "confirmed" のみ。他の status は送らない。
  body.status = data.status === "pending" ? "pending" : "confirmed";

  if (data.notes !== undefined && data.notes !== "") {
    body.notes = data.notes;
  }

  return body;
}

/**
 * 確定モーダルの初期値を組み立てるための tentative 行の最小形状。
 * tentative の start は意図日の JST 00:00 sentinel。
 * transformReservation の結果（Reservation）は追加フィールドを持つため
 * キャストなしで代入できる。
 */
export interface TentativeConfirmSource {
  start: Date;
  visitType?: "first" | "revisit";
  reservationTypeId?: string;
  doctorId?: string;
  notes?: string;
}

/**
 * 確定モーダル用 initialData を組み立てる。
 * - 日付は tentative の意図日。意図日が過去（期限超過 tentative）なら今日に倒す
 *   （新規予約と同じく過去日はバリデーション NG のため、過去日を初期値にしない）。
 * - 時刻は 10:00–11:00（新規予約作成のデフォルトと同一）。
 */
export function buildTentativeConfirmInitialData(
  tentative: TentativeConfirmSource,
  petId: string,
): Partial<Reservation> {
  const intendedDate = formatJSTWallDate(tentative.start);
  const date = isPastJSTDate(intendedDate) ? todayJSTISO() : intendedDate;
  const start = buildJSTWallDateTime(date, "10:00");
  return {
    petId,
    start,
    end: addHours(start, 1),
    visitType: tentative.visitType ?? "revisit",
    type: tentative.reservationTypeId ?? "",
    doctor: tentative.doctorId ?? "",
    isDesignated: false,
    status: "confirmed",
    notes: tentative.notes,
  };
}
