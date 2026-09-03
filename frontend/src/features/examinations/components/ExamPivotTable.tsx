import { useDeferredValue, useMemo, useState } from "react";
import { useQueries } from "@tanstack/react-query";

import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { TableCell, TableHead } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import type { SortOrder } from "@/types";

import { examinationItemsQueryOptions } from "../api/get-examination-items";
import type { ExaminationRecord, ExamResult } from "../api/transforms";
import { EXAM_PIVOT_RECENT_LIMIT } from "../constants";
import { ExamStatusBadge } from "./ExamStatusBadge";

interface ExamPivotTableProps {
  examinations: ExaminationRecord[];
  sortOrder: SortOrder;
}

interface PivotRow {
  key: string;
  label: string;
  unit: string;
  reference: string;
  isLegacy: boolean;
  sortOrder: number;
  cells: Readonly<Record<string, ExamResult>>;
}

function normalizeLegacyValue(value: string): string {
  return value.normalize("NFKC").trim();
}

function getPivotRowKey(
  examination: ExaminationRecord,
  item: ExamResult,
): string {
  if (item.examTypeFieldId !== undefined) {
    return `field:${item.examTypeFieldId}`;
  }

  return `legacy:${JSON.stringify([
    examination.testTypeId,
    normalizeLegacyValue(item.name),
    normalizeLegacyValue(item.unit),
  ])}`;
}

function formatStoredReference(item: ExamResult): string {
  if (item.referenceValue.trim()) {
    return item.referenceValue;
  }

  // 既知の制約: 保存済み refMin/refMax はマスタ解決済み基準値ではなく保存時点の値をそのまま表示する。
  if (item.refMin !== undefined || item.refMax !== undefined) {
    return `${item.refMin ?? ""}-${item.refMax ?? ""}`;
  }

  // 保存済み定性 bounds のみ表示。token を合成・補完しない（open end は数値と同型: '(-)-' / '-(+)'）。
  if (item.qualitativeMin !== undefined || item.qualitativeMax !== undefined) {
    return `${item.qualitativeMin ?? ""}-${item.qualitativeMax ?? ""}`;
  }

  return "-";
}

function buildPivotRows(
  examinations: ExaminationRecord[],
  itemsByExaminationId: Readonly<Record<string, readonly ExamResult[]>>,
): PivotRow[] {
  const rowsByKey = examinations.reduce<Record<string, PivotRow>>(
    (rows, examination) =>
      (itemsByExaminationId[examination.id] ?? []).reduce<Record<string, PivotRow>>(
        (currentRows, item) => {
          if (!item.inspectionValue.trim()) {
            return currentRows;
          }

          const baseKey = getPivotRowKey(examination, item);
          const occupiedKeys = Object.keys(currentRows).filter(
            (key) =>
              (key === baseKey || key.startsWith(`${baseKey}:duplicate:`)) &&
              currentRows[key]?.cells[examination.id] !== undefined,
          );
          const key =
            occupiedKeys.length === 0
              ? baseKey
              : `${baseKey}:duplicate:${occupiedKeys.length + 1}`;
          const existing = currentRows[key];
          const cells = {
            ...(existing?.cells ?? {}),
            [examination.id]: item,
          };

          if (existing) {
            return {
              ...currentRows,
              [key]: { ...existing, cells },
            };
          }

          return {
            ...currentRows,
            [key]: {
              key,
              label:
                item.examTypeFieldId === undefined
                  ? normalizeLegacyValue(item.name)
                  : item.name.trim(),
              unit: item.unit.trim() || "-",
              reference: formatStoredReference(item),
              isLegacy: item.examTypeFieldId === undefined,
              sortOrder: item.sortOrder,
              cells,
            },
          };
        },
        rows,
      ),
    {},
  );

  return Object.values(rowsByKey).sort(
    (left, right) =>
      left.sortOrder - right.sortOrder ||
      left.label.localeCompare(right.label, "ja"),
  );
}

function PivotValueCell({ item }: { item: ExamResult | undefined }) {
  if (!item) {
    return (
      <TableCell className={`min-w-24 border-l text-center ${C.borderMedium}`}>
        -
      </TableCell>
    );
  }

  const statusClass =
    item.status === "high"
      ? C.bgDanger8
      : item.status === "low"
        ? C.bgStatusBlueLight
        : C.bgWhite;

  return (
    <TableCell
      data-status={item.status}
      className={`min-w-24 border-l text-center ${C.borderMedium} ${statusClass}`}
    >
      <div className="flex min-h-8 items-center justify-center gap-2">
        <span className="font-mono">{item.inspectionValue}</span>
        <ExamStatusBadge status={item.status} isAssessed={item.isAssessed} compact />
      </div>
    </TableCell>
  );
}

function getExaminationColumnLabel(
  examination: ExaminationRecord,
  displayedExaminations: ExaminationRecord[],
): string {
  const sameDayExaminations = displayedExaminations.filter(
    (candidate) => candidate.date === examination.date,
  );
  if (sameDayExaminations.length === 1) {
    return examination.date;
  }

  const sameDayIndex = sameDayExaminations.findIndex(
    (candidate) => candidate.id === examination.id,
  );
  return `${examination.date} ${sameDayIndex + 1}件目`;
}

export function ExamPivotTable({
  examinations,
  sortOrder,
}: ExamPivotTableProps) {
  const [itemFilter, setItemFilter] = useState("");
  const deferredItemFilter = useDeferredValue(itemFilter);
  const recentExaminations = useMemo(
    () =>
      [...examinations]
        .sort((left, right) => right.date.localeCompare(left.date))
        .slice(0, EXAM_PIVOT_RECENT_LIMIT),
    [examinations],
  );
  const displayedExaminations = useMemo(
    () =>
      sortOrder === "asc"
        ? [...recentExaminations].reverse()
        : recentExaminations,
    [recentExaminations, sortOrder],
  );

  // 既知の制約: この N+1 fan-out は直近 EXAM_PIVOT_RECENT_LIMIT 件に意図的に絞っている。
  // 件数が増える場合は一括取得エンドポイントへの置換を検討する。
  const itemQueries = useQueries({
    queries: recentExaminations.map((examination) =>
      examinationItemsQueryOptions(examination.id),
    ),
  });
  const itemsByExaminationId = useMemo(
    () =>
      Object.fromEntries(
        recentExaminations.map((examination, index) => [
          examination.id,
          itemQueries[index]?.data ?? [],
        ]),
      ),
    [itemQueries, recentExaminations],
  );
  const rows = useMemo(
    () => buildPivotRows(recentExaminations, itemsByExaminationId),
    [itemsByExaminationId, recentExaminations],
  );
  const normalizedFilter = deferredItemFilter
    .normalize("NFKC")
    .trim()
    .toLocaleLowerCase();
  const visibleRows = useMemo(
    () =>
      normalizedFilter
        ? rows.filter((row) =>
            `${row.label} ${row.unit}`
              .normalize("NFKC")
              .toLocaleLowerCase()
              .includes(normalizedFilter),
          )
        : rows,
    [normalizedFilter, rows],
  );

  if (recentExaminations.length === 0) {
    return (
      <p className={`py-6 text-center text-sm ${C.text45}`}>
        検査記録がありません
      </p>
    );
  }

  if (itemQueries.some((query) => query.isError)) {
    return (
      <p
        role="alert"
        aria-label="検査項目の取得に失敗しました"
        className={`py-6 text-center text-sm ${C.danger}`}
      >
        検査項目の取得に失敗しました
      </p>
    );
  }

  if (itemQueries.some((query) => query.isPending)) {
    return (
      <p role="status" className={`py-6 text-center text-sm ${C.text45}`}>
        読み込み中...
      </p>
    );
  }

  return (
    <div className="space-y-3">
      <div className="space-y-1.5">
        <Label htmlFor="exam-pivot-item-filter" className={`text-sm ${C.text60}`}>
          検査項目
        </Label>
        <Input
          id="exam-pivot-item-filter"
          type="search"
          value={itemFilter}
          onChange={(event) => setItemFilter(event.target.value)}
          placeholder="項目名・単位で絞り込み"
          className={`h-11 ${C.bgWhite} ${C.borderMedium}`}
        />
      </div>

      {visibleRows.length > 0 ? (
        <div
          className={`overflow-x-auto rounded-lg border ${C.borderMedium} ${C.bgWhite}`}
        >
          <table className={`min-w-full border-collapse text-sm ${C.text}`}>
            <thead>
              <tr className={`${C.bgPage} ${STYLE.sectionLabel}`}>
                <TableHead className="min-w-36">
                  項目名
                </TableHead>
                <TableHead className={`min-w-20 border-l text-center ${C.borderMedium}`}>
                  単位
                </TableHead>
                <TableHead className={`min-w-24 border-l text-center ${C.borderMedium}`}>
                  基準値
                </TableHead>
                {displayedExaminations.map((examination) => (
                  <TableHead
                    key={examination.id}
                    aria-label={getExaminationColumnLabel(
                      examination,
                      displayedExaminations,
                    )}
                    className={`min-w-24 border-l text-center ${C.borderMedium}`}
                  >
                    {examination.date}
                  </TableHead>
                ))}
              </tr>
            </thead>
            <tbody>
              {visibleRows.map((row) => (
                <tr key={row.key} className={`border-t ${C.borderMedium}`}>
                  <TableHead scope="row">
                    <span>{row.label}</span>
                    {row.isLegacy ? (
                      <Badge
                        variant="outline"
                        className={`ml-2 ${C.textWarning} ${C.borderWarning20} ${C.bgWarning50}`}
                      >
                        未マッピング
                      </Badge>
                    ) : null}
                  </TableHead>
                  <TableCell className={`border-l text-center ${C.borderMedium} ${C.text60}`}>
                    {row.unit}
                  </TableCell>
                  <TableCell className={`border-l text-center ${C.borderMedium} ${C.text60}`}>
                    {row.reference}
                  </TableCell>
                  {displayedExaminations.map((examination) => (
                    <PivotValueCell
                      key={examination.id}
                      item={row.cells[examination.id]}
                    />
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <p className={`py-6 text-center text-sm ${C.text45}`}>
          {normalizedFilter
            ? "該当する検査項目がありません"
            : "表示できる実施済み項目がありません"}
        </p>
      )}
    </div>
  );
}
