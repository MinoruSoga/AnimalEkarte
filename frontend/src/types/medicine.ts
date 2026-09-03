import type { Medicine as BackendMedicine } from "@/types/generated/models";
import type { ReorderRequest } from "@/types/form";

/**
 * サーバー側で自動生成されるフィールド（リクエストに含めない）
 */
type ServerFields = "id" | "clinic_id" | "created_at" | "updated_at" | "inventory";

/**
 * リクエストで送信可能な Medicine フィールド
 * Goモデル変更 → make codegen → models.ts 更新で自動追従
 */
type MedicineWritable = Omit<BackendMedicine, ServerFields>;

/**
 * 薬剤作成リクエスト（バックエンドAPI）
 * name のみ必須、残りはoptional
 */
export type CreateMedicineRequest = Pick<MedicineWritable, "name"> &
  Partial<Omit<MedicineWritable, "name">>;

/**
 * 薬剤更新リクエスト（バックエンドAPI）
 * PATCH: 全フィールドoptional
 * clear_parent_id: true を送ると parent_id を NULL にクリアする
 */
export type UpdateMedicineRequest = Partial<MedicineWritable> & {
  clear_parent_id?: boolean;
};

/**
 * 薬剤並び替えリクエスト
 * models.ts に対応フィールドなし → 手書き
 */
export type ReorderMedicinesRequest = ReorderRequest;
