import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ShiftTemplate } from "../types";
import { isShiftType } from "../types";

function normalizeTime(t: string | undefined | null): string {
  if (!t) return "";
  const parts = t.split(":");
  if (parts.length >= 2) {
    return `${parts[0].padStart(2, "0")}:${parts[1].padStart(2, "0")}`;
  }
  return t;
}

interface RawBreak {
  id?: number;
  shift_template_id?: number;
  break_start?: string;
  break_end?: string;
}

interface RawTemplate {
  id?: number;
  clinic_id?: number;
  name?: string;
  shift_type?: string;
  start_time?: string | null;
  end_time?: string | null;
  notes?: string;
  sort_order?: number;
  is_active?: boolean;
  breaks?: RawBreak[];
  created_at?: string;
  updated_at?: string;
}

function transformTemplate(d: RawTemplate): ShiftTemplate {
  return {
    id: String(d.id ?? 0),
    clinic_id: String(d.clinic_id ?? 0),
    name: d.name ?? "",
    shift_type: isShiftType(d.shift_type) ? d.shift_type : "full",
    start_time: normalizeTime(d.start_time),
    end_time: normalizeTime(d.end_time),
    notes: d.notes ?? "",
    sort_order: d.sort_order ?? 0,
    is_active: d.is_active ?? true,
    breaks: (d.breaks ?? []).map((b) => ({
      id: String(b.id ?? 0),
      shift_template_id: String(b.shift_template_id ?? 0),
      break_start: normalizeTime(b.break_start),
      break_end: normalizeTime(b.break_end),
    })),
    created_at: d.created_at ?? "",
    updated_at: d.updated_at ?? "",
  };
}

async function getShiftTemplates(): Promise<ShiftTemplate[]> {
  const { data } = await axios.get<RawTemplate[]>("/v1/shift-templates");
  return (data ?? []).map(transformTemplate);
}

export function useGetShiftTemplates() {
  return useQuery({
    queryKey: queryKeys.shiftTemplates.all(),
    queryFn: getShiftTemplates,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
