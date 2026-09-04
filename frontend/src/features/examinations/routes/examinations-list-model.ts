import { Calendar, CircleDot, FlaskConical, User } from "lucide-react";
import { paths } from "@/config/paths";
import type {
  ActiveFilter,
  FilterProperty,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import {
  CONDITIONS_NO_EMPTY,
  CONDITIONS_WITH_EMPTY,
} from "@/components/shared/PropertyFilter/types";
import { uniqueSortedOptions } from "@/lib/unique-sorted-options";
import type { ExaminationRecord } from "../api/transforms";

const EXAMINATION_LIST_CHART_TAB = "検査";
const EXAMINATION_LIST_EXAM_ID_PARAM = "examId";

export const EXAMINATIONS_PAGE_SIZE = 20;

const CLIENT_ONLY_FILTER_KEYS = ["status", "testType", "doctor"];

export const EXAMINATION_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "日時" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "testType", label: "検査種別" },
  { key: "doctor", label: "担当医" },
  { key: "status", label: "ステータス" },
];

const STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    // exams.status DEFAULT 'pending' — 空値は存在しない
    // transform で EN→JA 変換済のためフィルタ値も日本語で指定
    // SD-12: backend ExaminationStatus 全5値（pending/in_progress/result_entered/completed/confirmed）に追従
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "依頼中", label: "依頼中" },
      { value: "検査中", label: "検査中" },
      { value: "結果入力済み", label: "結果入力済み" },
      { value: "完了", label: "完了" },
      { value: "確定", label: "確定" },
    ],
  },
];

export function examinationDateFilters(activeFilters: ActiveFilter[]): {
  startDate?: string;
  endDate?: string;
} {
  const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
    { from?: string; to?: string } | undefined;
  return {
    startDate: dateFilter?.from,
    endDate: dateFilter?.to,
  };
}

export function buildExaminationFilterProperties(
  allExaminations: ExaminationRecord[],
): FilterProperty[] {
  const testTypeOptions = uniqueSortedOptions(allExaminations, (r) => r.testType);
  const doctorOptions = uniqueSortedOptions(allExaminations, (r) => r.doctor);
  return [
    ...STATIC_FILTER_PROPERTIES,
    {
      key: "testType",
      label: "検査種別",
      type: "select" as const,
      icon: FlaskConical,
      conditions: CONDITIONS_WITH_EMPTY,
      options: testTypeOptions,
    },
    {
      key: "doctor",
      label: "担当医",
      type: "select" as const,
      icon: User,
      conditions: CONDITIONS_WITH_EMPTY,
      options: doctorOptions,
    },
  ];
}

export function hasExaminationPageScopedFilter(
  deferredSearch: string,
  activeFilters: ActiveFilter[],
): boolean {
  return (
    deferredSearch !== "" || activeFilters.some((f) => CLIENT_ONLY_FILTER_KEYS.includes(f.key))
  );
}

export interface ServerPagePagination<T> {
  paginatedData: T[];
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  currentPage: number;
}

export function buildServerPagePagination<T>(input: {
  rows: T[];
  total: number;
  page: number;
  limit: number;
}): ServerPagePagination<T> {
  const totalPages = Math.max(1, Math.ceil(input.total / input.limit));
  return {
    paginatedData: input.rows,
    totalPages,
    totalCount: input.total,
    startIndex: input.total === 0 ? 0 : (input.page - 1) * input.limit + 1,
    endIndex: Math.min(input.page * input.limit, input.total),
    currentPage: input.page,
  };
}

export function nextListSearchParamsWithPage(prev: URLSearchParams, page: number): URLSearchParams {
  const next = new URLSearchParams(prev);
  if (page === 1) {
    next.delete("page");
  } else {
    next.set("page", String(page));
  }
  return next;
}

export function nextListSearchParamsWithoutPage(prev: URLSearchParams): URLSearchParams {
  const next = new URLSearchParams(prev);
  next.delete("page");
  return next;
}

export function examinationListDetailHref(input: { id: string; medicalRecordId?: string }): string {
  if (input.medicalRecordId) {
    const params = new URLSearchParams({
      tab: EXAMINATION_LIST_CHART_TAB,
      [EXAMINATION_LIST_EXAM_ID_PARAM]: input.id,
    });
    return `${paths.medicalRecords.detail.getHref(input.medicalRecordId)}?${params.toString()}`;
  }
  return paths.examinations.detail.getHref(input.id);
}

export function examinationCreateHref(petId: string): string {
  const params = new URLSearchParams({ petId });
  return `${paths.examinations.new.getHref()}?${params.toString()}`;
}
