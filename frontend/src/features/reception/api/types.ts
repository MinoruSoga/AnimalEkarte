/**
 * Reception API types
 * ReceptionAppointment と ReceptionColumn は transforms.ts の ReturnType から導出。
 * このファイルは後方互換のため re-export を維持する。
 */
export type { ReceptionAppointment, ReceptionColumn } from "./transforms";
