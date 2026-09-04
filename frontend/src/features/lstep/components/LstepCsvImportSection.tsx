import { useActionState, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { UploadCloud } from "lucide-react";

import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { TableCell, TableHead } from "@/components/ui/table";
import { C, ICON, PALETTE } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { formatJSTDateTimeLocal } from "@/lib/jst-date";
import { queryKeys } from "@/lib/query-keys";

import { useGetLstepCsvImports } from "../api/get-lstep-csv-imports";
import { uploadFriendAttributesCsv } from "../api/upload-friend-attributes-csv";
import { CSV_STATUS_LABELS } from "../lib/lstep-analytics-model";

export function LstepCsvImportSection() {
  return (
    <section aria-labelledby="csv-import-heading" className="space-y-4 mt-8">
      <h2 id="csv-import-heading" className={`text-base font-semibold ${C.text80}`}>
        友だち属性 CSV インポート
      </h2>

      <div className={`border ${C.borderLight} rounded-xs ${C.bgWhite} p-4 space-y-6`}>
        <div>
          <p className={`text-sm font-medium ${C.text80} mb-3`}>新規アップロード</p>
          <CsvUploadSection />
        </div>
        <hr className={C.borderLight} />
        <div>
          <p className={`text-sm font-medium ${C.text80} mb-3`}>インポート履歴</p>
          <CsvImportHistoryTable />
        </div>
      </div>
    </section>
  );
}

type UploadState = { success: true } | { error: string } | null;

function CsvUploadSection() {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedFileName, setSelectedFileName] = useState("");
  const [state, formAction, isPending] = useActionState<UploadState, FormData>(async () => {
    const file = fileInputRef.current?.files?.[0];
    if (!file || file.size === 0) {
      return { error: "CSVファイルを選択してください" };
    }
    try {
      const clinicId = requireStoredClinicId();
      await uploadFriendAttributesCsv(clinicId, file);
      queryClient.invalidateQueries({ queryKey: queryKeys.lstepCsvImports.all() });
      return { success: true };
    } catch (err) {
      handleApiError(err, "CSVアップロード");
      return { error: "アップロードに失敗しました" };
    }
  }, null);

  return (
    <form action={formAction} className="space-y-3">
      <div
        className={`flex flex-col items-stretch gap-3 p-4 border ${C.borderLight} rounded-xs ${C.bgLight} sm:flex-row sm:items-center`}
        data-testid="csv-upload-row"
      >
        <div className="flex min-w-0 items-center gap-3">
          <UploadCloud className={`${ICON.action} ${C.text40} shrink-0`} />
          <div className="min-w-0 flex-1">
            <p className={`text-sm font-medium ${C.text80}`}>友だち属性 CSV</p>
            <p className={`text-xs ${C.text40}`}>
              Lステップ管理画面からエクスポートした CSV をアップロード
            </p>
          </div>
        </div>
        <div
          className="flex w-full min-w-0 items-center gap-2 sm:w-auto sm:shrink-0"
          data-testid="csv-upload-controls"
        >
          <label
            className={`relative inline-flex min-h-11 min-w-11 shrink-0 cursor-pointer items-center justify-center rounded-xs border ${C.borderLight} ${C.bgWhite} px-3 text-xs font-medium ${C.text60} focus-within:ring-2 focus-within:ring-ring`}
          >
            <input
              ref={fileInputRef}
              type="file"
              name="file"
              accept=".csv"
              className="absolute inset-0 size-full cursor-pointer opacity-0"
              aria-label="友だち属性CSVファイルを選択"
              onChange={(event) => setSelectedFileName(event.target.files?.[0]?.name ?? "")}
            />
            CSVファイルを選択
          </label>
          <span
            className={`min-w-0 flex-1 truncate text-xs ${C.text60} sm:max-w-40`}
            aria-live="polite"
            data-testid="csv-selected-file-name"
          >
            {selectedFileName || "未選択"}
          </span>
        </div>
      </div>
      {state && "error" in state ? <p className={`text-sm ${C.danger}`}>{state.error}</p> : null}
      {state && "success" in state ? (
        <p className={`text-sm ${C.textSuccess}`}>アップロードが完了しました</p>
      ) : null}
      <div className="flex justify-end">
        <SubmitButton disabled={isPending}>
          {isPending ? "アップロード中..." : "アップロード"}
        </SubmitButton>
      </div>
    </form>
  );
}

function CsvImportHistoryTable() {
  const { data, isLoading } = useGetLstepCsvImports(20);

  if (isLoading) {
    return <p className={`text-sm ${C.text40} py-4`}>読み込み中...</p>;
  }
  if (!data || data.length === 0) {
    return <p className={`text-sm ${C.text40} py-8 text-center`}>インポート履歴はありません</p>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className={`${C.bgLight} border-b ${C.borderLight}`}>
            <TableHead className={C.text55}>ファイル名</TableHead>
            <TableHead className={`text-right ${C.text55}`}>総行数</TableHead>
            <TableHead className={`text-right ${C.text55}`}>成功</TableHead>
            <TableHead className={`text-right ${C.text55}`}>失敗</TableHead>
            <TableHead className={`text-center ${C.text55}`}>ステータス</TableHead>
            <TableHead className={C.text55}>アップロード日時</TableHead>
          </tr>
        </thead>
        <tbody>
          {data.map((item) => (
            <tr key={item.id} className={`border-b ${C.borderLight}`}>
              <TableCell className={`${C.text80} max-w-[240px] truncate`} title={item.file_name}>
                {item.file_name}
              </TableCell>
              <TableCell className={`text-right ${C.text60} tabular-nums`}>
                {item.row_count.toLocaleString()}
              </TableCell>
              <TableCell
                className="text-right tabular-nums"
                style={{ color: PALETTE.successGreen }}
              >
                {item.success_count.toLocaleString()}
              </TableCell>
              <TableCell
                className="text-right tabular-nums"
                style={{ color: item.error_count > 0 ? PALETTE.danger : undefined }}
              >
                {item.error_count.toLocaleString()}
              </TableCell>
              <TableCell className="text-center">
                <span
                  className={`text-xs px-2 py-0.5 rounded-full ${
                    item.status === "completed"
                      ? `${C.bgSuccess10} ${C.textSuccess}`
                      : item.status === "failed"
                        ? `${C.bgDanger10} ${C.danger}`
                        : `${C.bgLight} ${C.text60}`
                  }`}
                >
                  {CSV_STATUS_LABELS[item.status] ?? item.status}
                </span>
              </TableCell>
              <TableCell className={C.text60}>
                {formatJSTDateTimeLocal(item.created_at).replace("T", " ")}
              </TableCell>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
