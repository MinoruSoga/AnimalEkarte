import type { CreateShiftTemplateInput, UpdateShiftTemplateInput } from "../types";
import type { TemplateFormData } from "./shift-template-form-model";

// EMR-241: 時刻や休憩の送信可否をカテゴリ名で分岐しない。
// 空の時刻はそのままサーバへ送り、必須判定はバックエンドの RequiresTimeSlot 検証に委ねる。
export function toShiftTemplateCreateInput(formData: TemplateFormData): CreateShiftTemplateInput {
  const breaks = formData.breaks.filter((b) => b.break_start && b.break_end);
  return {
    name: formData.name,
    shift_type: formData.shift_type,
    start_time: formData.start_time || undefined,
    end_time: formData.end_time || undefined,
    notes: formData.notes,
    is_active: formData.is_active,
    breaks,
  };
}

export function toShiftTemplateUpdateInput(formData: TemplateFormData): UpdateShiftTemplateInput {
  const breaks = formData.breaks.filter((b) => b.break_start && b.break_end);
  // 更新は PATCH 意味論: 空文字 "" は「クリア」として送る（未指定なら既存値を維持してしまう）。
  return {
    name: formData.name,
    shift_type: formData.shift_type,
    start_time: formData.start_time,
    end_time: formData.end_time,
    notes: formData.notes,
    is_active: formData.is_active,
    breaks,
  };
}
