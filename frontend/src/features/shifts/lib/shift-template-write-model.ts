import type { CreateShiftTemplateInput, UpdateShiftTemplateInput } from "../types";
import { isShiftTemplateTimeHidden } from "./shift-template-form-utils";
import type { TemplateFormData } from "./shift-template-form-model";

export function toShiftTemplateCreateInput(formData: TemplateFormData): CreateShiftTemplateInput {
  const breaks = formData.breaks.filter((b) => b.break_start && b.break_end);
  const isTimeHidden = isShiftTemplateTimeHidden(formData.shift_type);
  return {
    name: formData.name,
    shift_type: formData.shift_type,
    start_time: isTimeHidden ? undefined : formData.start_time || undefined,
    end_time: isTimeHidden ? undefined : formData.end_time || undefined,
    notes: formData.notes,
    is_active: formData.is_active,
    breaks: isTimeHidden ? [] : breaks,
  };
}

export function toShiftTemplateUpdateInput(formData: TemplateFormData): UpdateShiftTemplateInput {
  const breaks = formData.breaks.filter((b) => b.break_start && b.break_end);
  const isTimeHidden = isShiftTemplateTimeHidden(formData.shift_type);
  return {
    name: formData.name,
    shift_type: formData.shift_type,
    start_time: isTimeHidden ? null : formData.start_time || undefined,
    end_time: isTimeHidden ? null : formData.end_time || undefined,
    notes: formData.notes,
    is_active: formData.is_active,
    breaks: isTimeHidden ? [] : breaks,
  };
}
