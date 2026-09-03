/**
 * ReservationFormModal.*.test.tsx 間で共有するテストフィクスチャ。
 * FE-RC-045: 936行の単一テストファイルをトピック別に分割した際に抽出。
 *
 * 注意: `vi.mock("@/components/ui/searchable-select", ...)` はここに置かない。
 * Vitest の hoisting は各テストファイル内の静的な `vi.mock` 呼び出しを前提としており、
 * import 経由の間接呼び出しでは、依存グラフの評価順によってモック登録前に実モジュールが
 * 解決される可能性がある。そのため `vi.mock` は各分割ファイルの先頭に直接記述すること。
 */
import { http, HttpResponse } from "msw";
import { createTestWrapper } from "@/testing/utils";

export function createWrapper() {
  return createTestWrapper({ router: true });
}

/** 空レスポンスを返すハンドラ群。ReservationFormModal 内部のクエリを全て黙らせる */
export const silentApiHandlers = [
  http.get("/api/v1/clinic-holidays", () => HttpResponse.json([])),
  http.get("/api/v1/pets", () => HttpResponse.json({ data: [] })),
  http.get("/api/v1/masters/reservation-types", () => HttpResponse.json([])),
  http.get("/api/v1/masters/staffs", () => HttpResponse.json([])),
  http.get("/api/v1/shifts/on-duty-staffs", () => HttpResponse.json([])),
  http.get("/api/v1/masters/animal-species", () => HttpResponse.json([])),
];

export const noop = () => {};
