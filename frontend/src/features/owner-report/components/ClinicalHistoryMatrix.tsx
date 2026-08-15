import type { ReactNode } from "react";

import { C } from "@/lib/design-tokens";

import type {
  ClinicalHistoryKind,
  ClinicalHistoryMatrix as ClinicalHistoryMatrixModel,
} from "../lib/clinical-briefing";

export type HistoryRowState =
  "ready" | "partial-permission" | "no-permission" | "loading" | "error";

interface ClinicalHistoryMatrixProps {
  matrix: ClinicalHistoryMatrixModel;
  rowStates: Record<ClinicalHistoryKind, HistoryRowState>;
}

type HistoryColumn = ClinicalHistoryMatrixModel["columns"][number];
type HistoryEntry = HistoryColumn["entries"][number];

const rowToneClasses: Record<
  ClinicalHistoryKind,
  { dot: string; heading: string }
> = {
  診療: { dot: C.bgBrandDot, heading: C.bgBrandLight50 },
  検査: { dot: C.bgStatusPurpleDot, heading: C.bgStatusPurple },
  "薬・処方": { dot: C.bgStatusAmberDot, heading: C.bgStatusAmber },
  予防接種: { dot: C.bgStatusEmeraldDot, heading: C.bgStatusEmerald },
  処置: { dot: C.bgStatusRedDot, heading: C.bgRedLight },
  ケア: { dot: C.bgStatusBlueDot, heading: C.bgStatusBlueLight },
};

const rowStateLabels: Record<Exclude<HistoryRowState, "ready">, string> = {
  "partial-permission": "一部閲覧権限なし",
  "no-permission": "閲覧権限なし",
  loading: "読み込み中...",
  error: "取得失敗",
};

function historyStateLabel(state: HistoryRowState): string {
  return state === "ready" ? "" : rowStateLabels[state];
}

function EntryCard({ entry }: { entry: HistoryEntry }) {
  return (
    <article
      className={`border-l-2 pl-1.5 ${entry.isAlert ? C.borderDanger : C.borderBrand}`}
    >
      <div className="flex min-w-0 items-center justify-between gap-1">
        <span
          className={`truncate text-2xs font-semibold ${entry.isAlert ? C.danger : C.textBrand}`}
        >
          {entry.source}
        </span>
        {entry.status ? (
          <span
            className={`shrink-0 text-2xs font-semibold ${entry.isAlert ? C.danger : C.text60}`}
          >
            {entry.status}
          </span>
        ) : null}
      </div>
      <strong
        className={`mt-0.5 block text-sm leading-snug font-medium ${C.text}`}
      >
        {entry.title}
      </strong>
      {entry.detail ? (
        <p className={`mt-0.5 text-xs leading-snug ${C.text50}`}>
          {entry.detail}
        </p>
      ) : null}
    </article>
  );
}

interface HistoryCellContentProps {
  canShowEntries: boolean;
  entries: ReadonlyArray<HistoryEntry>;
  index: number;
  state: HistoryRowState;
}

function HistoryCellContent({
  canShowEntries,
  entries,
  index,
  state,
}: HistoryCellContentProps): ReactNode {
  if (!canShowEntries) {
    return index === 0 ? (
      <p
        className={`text-xs ${state === "error" ? C.danger : C.text50}`}
        role={
          state === "error"
            ? "alert"
            : state === "loading"
              ? "status"
              : undefined
        }
      >
        {historyStateLabel(state)}
      </p>
    ) : null;
  }
  const showsPermissionNotice = state === "partial-permission" && index === 0;
  if (entries.length === 0 && !showsPermissionNotice) {
    return <span className={`block text-center text-xs ${C.text25}`}>—</span>;
  }
  return (
    <div className="flex flex-col gap-1.5">
      {state === "partial-permission" && index === 0 ? (
        <p className={`text-2xs ${C.text50}`}>{historyStateLabel(state)}</p>
      ) : null}
      {entries.map((entry) => (
        <EntryCard key={entry.id} entry={entry} />
      ))}
    </div>
  );
}

interface HistoryCellProps {
  column: HistoryColumn;
  index: number;
  kind: ClinicalHistoryKind;
  state: HistoryRowState;
}

function HistoryCell({ column, index, kind, state }: HistoryCellProps) {
  const entries = column.entries.filter((entry) => entry.kind === kind);
  const canShowEntries = state === "ready" || state === "partial-permission";
  return (
    <td className={`border-r border-b px-2 py-1 align-middle ${C.borderLight}`}>
      <HistoryCellContent
        canShowEntries={canShowEntries}
        entries={entries}
        index={index}
        state={state}
      />
    </td>
  );
}

interface HistoryRowProps {
  columns: ReadonlyArray<HistoryColumn>;
  kind: ClinicalHistoryKind;
  count: number;
  state: HistoryRowState;
}

function HistoryRow({ columns, kind, count, state }: HistoryRowProps) {
  const tone = rowToneClasses[kind];
  const canShowEntries = state === "ready" || state === "partial-permission";
  return (
    <tr>
      <th
        scope="row"
        className={`sticky left-0 z-10 border-r border-b px-2 py-1 align-middle ${C.borderLight} ${tone.heading}`}
      >
        <span className="flex items-center gap-1">
          <span
            aria-hidden="true"
            className={`size-1.5 shrink-0 rounded-full ${tone.dot}`}
          />
          <strong className={`text-2xs font-semibold ${C.text}`}>{kind}</strong>
        </span>
        <small
          className={`ml-2.5 block text-2xs font-semibold tabular-nums ${C.text60}`}
        >
          {canShowEntries ? `${count}件` : "—"}
        </small>
      </th>
      {columns.map((column, index) => (
        <HistoryCell
          key={`${kind}-${column.dateKey}`}
          column={column}
          index={index}
          kind={kind}
          state={state}
        />
      ))}
    </tr>
  );
}

function HistoryTableHeader({
  columns,
}: {
  columns: ReadonlyArray<HistoryColumn>;
}) {
  return (
    <thead className="sticky top-0 z-20">
      <tr>
        <th
          scope="col"
          className={`sticky left-0 z-30 border-r border-b px-2 py-1 text-2xs font-semibold ${C.borderLight} ${C.bgPage} ${C.text60}`}
        >
          種類
        </th>
        {columns.map((column) => (
          <th
            key={column.dateKey}
            scope="col"
            className={`border-r border-b px-2 py-1 text-2xs font-semibold tabular-nums ${C.borderLight} ${C.bgPage} ${C.text}`}
          >
            {column.label}
          </th>
        ))}
      </tr>
    </thead>
  );
}

/** 縦=情報種類、横=日付。薬・予防接種・処置を混同しない履歴マトリクス。 */
export function ClinicalHistoryMatrix({
  matrix,
  rowStates,
}: ClinicalHistoryMatrixProps) {
  const columns: ReadonlyArray<HistoryColumn> =
    matrix.columns.length > 0
      ? matrix.columns
      : [{ dateKey: "empty", label: "記録なし", entries: [] }];
  return (
    <table
      aria-label="診療履歴を種類別に分け、日付の新しい順に左から表示"
      className={`owner-report-history-matrix border-separate border-spacing-0 text-left ${C.bgWhite}`}
      style={{ minWidth: `${5.25 + columns.length * 12.5}rem` }}
    >
      <colgroup>
        <col className="owner-report-history-kind-column" />
        {columns.map((column) => (
          <col
            key={column.dateKey}
            className="owner-report-history-record-column"
          />
        ))}
      </colgroup>
      <HistoryTableHeader columns={columns} />
      <tbody>
        {matrix.rows.map((row) => (
          <HistoryRow
            key={row.kind}
            columns={columns}
            kind={row.kind}
            count={row.count}
            state={rowStates[row.kind]}
          />
        ))}
      </tbody>
    </table>
  );
}
