import { useActionState, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { UploadCloud } from "lucide-react";

import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, PALETTE } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { axios } from "@/lib/axios";
import { formatJSTDateTimeLocal } from "@/lib/jst-date";

import { useGetLstepCsvImports } from "../api/get-lstep-csv-imports";
import { CSV_STATUS_LABELS } from "./LstepAnalyticsModel";

export function CsvImportSection() {
  return (
    <section aria-labelledby="csv-import-heading" className="space-y-4 mt-8">
      <h2 id="csv-import-heading" className={`text-base font-semibold ${C.text80}`}>
        友だち属性 CSV インポート
      </h2>

      <div className={`border ${C.borderLight} rounded-[4px] ${C.bgWhite} p-4 space-y-5`}>
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
  const [state, formAction, isPending] = useActionState<UploadState, FormData>(
    async () => {
      const file = fileInputRef.current?.files?.[0];
      if (!file || file.size === 0) {
        return { error: "CSVファイルを選択してください" };
      }
      try {
        const clinicId = localStorage.getItem("auth_current_clinic:v1");
        if (!clinicId) {
          throw new Error("クリニックが選択されていません。ページをリロードしてください。");
        }
        const fd = new FormData();
        fd.append("file", file);
        await axios.post(
          `/v1/clinics/${clinicId}/lstep/csv-imports/friend-attributes`,
          fd,
          { headers: { "Content-Type": "multipart/form-data" } },
        );
        queryClient.invalidateQueries({ queryKey: ["lstep-csv-imports"] });
        return { success: true };
      } catch (err) {
        handleApiError(err, "CSVアップロード");
        return { error: "アップロードに失敗しました" };
      }
    },
    null,
  );

  return (
    <form action={formAction} className="space-y-3">
      <div className={`flex items-center gap-3 p-4 border ${C.borderLight} rounded-[4px] ${C.bgLight}`}>
        <UploadCloud className={`${ICON.action} ${C.text40} shrink-0`} />
        <div className="flex-1 min-w-0">
          <p className={`text-sm font-medium ${C.text80}`}>友だち属性 CSV</p>
          <p className={`text-xs ${C.text40}`}>
            Lステップ管理画面からエクスポートした CSV をアップロード
          </p>
        </div>
        <input
          ref={fileInputRef}
          type="file"
          name="file"
          accept=".csv"
          className={`text-sm ${C.text60} file:mr-3 file:px-3 file:py-1.5 file:rounded file:border-0 file:text-xs file:font-medium file:${C.bgWhite} file:border file:${C.borderLight}`}
          aria-label="友だち属性CSVファイルを選択"
        />
      </div>
      {state && "error" in state ? (
        <p className={`text-sm text-[${PALETTE.danger}]`}>{state.error}</p>
      ) : null}
      {state && "success" in state ? (
        <p className={`text-sm text-[${PALETTE.successGreen}]`}>
          アップロードが完了しました
        </p>
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
    return (
      <p className={`text-sm ${C.text40} py-8 text-center`}>
        インポート履歴はありません
      </p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm border-collapse">
        <thead>
          <tr className={`${C.bgLight} border-b ${C.borderLight}`}>
            <th className={`text-left px-3 py-2 font-medium ${C.text80}`}>ファイル名</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text80}`}>総行数</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text80}`}>成功</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text80}`}>失敗</th>
            <th className={`text-center px-3 py-2 font-medium ${C.text80}`}>ステータス</th>
            <th className={`text-left px-3 py-2 font-medium ${C.text80}`}>アップロード日時</th>
          </tr>
        </thead>
        <tbody>
          {data.map((item) => (
            <tr key={item.id} className={`border-b ${C.borderLight}`}>
              <td className={`px-3 py-2 ${C.text80} max-w-[240px] truncate`} title={item.file_name}>
                {item.file_name}
              </td>
              <td className={`text-right px-3 py-2 ${C.text60} tabular-nums`}>
                {item.row_count.toLocaleString()}
              </td>
              <td className="text-right px-3 py-2 tabular-nums" style={{ color: PALETTE.successGreen }}>
                {item.success_count.toLocaleString()}
              </td>
              <td
                className="text-right px-3 py-2 tabular-nums"
                style={{ color: item.error_count > 0 ? PALETTE.danger : undefined }}
              >
                {item.error_count.toLocaleString()}
              </td>
              <td className="text-center px-3 py-2">
                <span
                  className={`text-xs px-2 py-0.5 rounded-full ${
                    item.status === "completed"
                      ? `bg-[${PALETTE.successGreen}]/10 text-[${PALETTE.successGreen}]`
                      : item.status === "failed"
                        ? `bg-[${PALETTE.danger}]/10 text-[${PALETTE.danger}]`
                        : `${C.bgLight} ${C.text60}`
                  }`}
                >
                  {CSV_STATUS_LABELS[item.status] ?? item.status}
                </span>
              </td>
              <td className={`px-3 py-2 ${C.text60} text-xs`}>
                {formatJSTDateTimeLocal(item.created_at).replace("T", " ")}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
