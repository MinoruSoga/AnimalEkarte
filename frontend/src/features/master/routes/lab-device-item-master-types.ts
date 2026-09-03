/**
 * lab-device-item-master-settings-model の共有型・定数。
 * ラベル表示 / draft 変換 / バリデーションの各モジュールから参照する。
 */
import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";

export const LAB_DEVICE_UNMAPPED_FIELD = "__unmapped__";
export const LAB_DEVICE_EXAM_SELECT_UNSET = "__unset__";
export const LAB_DEVICE_EXAM_UNSET = "未設定";
export const LAB_DEVICE_EXAM_MIXED = "複数の検査";

export const LAB_DEVICE_SOURCE_ORDER = [
  "fuji_nx600",
  "fuji_au10v",
  "arkray_pu4010",
  "idexx_vetlab",
] as const;

export const LAB_DEVICE_SOURCE_LABELS: Record<string, string> = {
  fuji_nx600: "NX600",
  fuji_au10v: "AU10V",
  arkray_pu4010: "尿（PU-4010）",
  idexx_vetlab: "IDEXX VetLab",
};

export const LAB_DEVICE_VALUE_SHAPE_LABELS: Record<string, string> = {
  numeric: "数値",
  inequality: "不等号",
  qual_and_num: "定性+数値",
  dash: "ダッシュ",
  text: "テキスト",
};

export type LabDeviceExamFieldOption = {
  id: string;
  label: string;
};

export type LabDeviceItemMasterGroup = {
  sourceType: string;
  label: string;
  items: LabDeviceItemMaster[];
};

export type LabDeviceRow = {
  id: string;
  sourceType: string;
  name: string;
  examTypeId: string | null;
  examLabel: string;
  isActive: boolean;
  sortOrder: number;
  itemCount: number;
  unmappedCount: number;
};

export type LabDeviceFormData = {
  name: string;
  sourceType: string;
  examTypeId: string | null;
  isActive: boolean;
  sortOrder: number;
};

export type LabDeviceItemDraft = {
  id: string;
  examTypeFieldId: string | null;
  isActive: boolean;
};
