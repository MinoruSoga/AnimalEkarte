import type { ShiftTemplate, ShiftType } from "../types";

export interface BreakInput {
  break_start: string;
  break_end: string;
}

export interface TemplateFormData {
  name: string;
  shift_type: ShiftType;
  start_time: string;
  end_time: string;
  notes: string;
  is_active: boolean;
  breaks: BreakInput[];
}

export const DEFAULT_SHIFT_TEMPLATE_FORM: TemplateFormData = {
  name: "",
  shift_type: "full",
  start_time: "",
  end_time: "",
  notes: "",
  is_active: true,
  breaks: [],
};

export function templateToFormData(template: ShiftTemplate): TemplateFormData {
  return {
    name: template.name,
    shift_type: template.shift_type,
    start_time: template.start_time,
    end_time: template.end_time,
    notes: template.notes,
    is_active: template.is_active,
    breaks: template.breaks.map((b) => ({
      break_start: b.break_start,
      break_end: b.break_end,
    })),
  };
}
