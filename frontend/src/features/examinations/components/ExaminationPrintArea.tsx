import { PrintPortal } from "@/components/shared/PrintPortal";
import type { ExaminationPrintModel } from "../lib/examination-print-model";

interface ExaminationPrintAreaProps {
  model: ExaminationPrintModel;
}

/**
 * TASK-031: 検査結果の印刷／PDF 出力面。
 * 描画源は GET print-snapshot 由来の pure model のみ（formItems 不可）。
 */
export function ExaminationPrintArea({ model }: ExaminationPrintAreaProps) {
  return (
    <PrintPortal testId="examination-print-area" orientation="landscape">
      <div className="mb-3 text-center relative">
        {model.isDraft && model.watermark ? (
          <p
            data-testid="examination-print-watermark"
            className="absolute inset-x-0 top-0 text-[18pt] font-bold text-red-600 opacity-70 tracking-widest"
          >
            {model.watermark}
          </p>
        ) : null}
        <h1 className="text-[14pt] font-bold pt-6">{model.title}</h1>
        <p className="text-[10pt]">
          {model.examTypeName ? `${model.examTypeName} / ` : null}
          検査日: {model.date}
          {` / v${model.version}`}
        </p>
      </div>

      <table className="w-full text-[9pt] border-collapse mb-3">
        <tbody>
          <tr>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left w-24">
              カルテ番号
            </th>
            <td className="border border-gray-300 px-2 py-1">
              {model.medicalRecordNo || "—"}
            </td>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left w-24">
              飼主
            </th>
            <td className="border border-gray-300 px-2 py-1">
              {model.ownerName || "—"}
            </td>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left w-24">
              ペット
            </th>
            <td className="border border-gray-300 px-2 py-1">
              {model.petName || "—"}
              {model.speciesName ? ` (${model.speciesName})` : null}
            </td>
          </tr>
          <tr>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left">
              担当医
            </th>
            <td className="border border-gray-300 px-2 py-1">
              {model.doctorName || "—"}
            </td>
            <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left">
              機器
            </th>
            <td className="border border-gray-300 px-2 py-1" colSpan={3}>
              {model.machine || "—"}
            </td>
          </tr>
          {model.resultSummary ? (
            <tr>
              <th className="border border-gray-400 bg-gray-100 px-2 py-1 text-left">
                結果要約
              </th>
              <td className="border border-gray-300 px-2 py-1" colSpan={5}>
                {model.resultSummary}
              </td>
            </tr>
          ) : null}
        </tbody>
      </table>

      <p className="font-semibold text-[9pt] mb-1">検査項目</p>
      <table className="w-full text-[9pt] border-collapse">
        <thead>
          <tr className="bg-gray-100">
            <th className="border border-gray-400 px-1 py-0.5 text-left">項目</th>
            <th className="border border-gray-400 px-1 py-0.5 text-right">結果</th>
            <th className="border border-gray-400 px-1 py-0.5 text-left">単位</th>
            <th className="border border-gray-400 px-1 py-0.5 text-left">基準</th>
            <th className="border border-gray-400 px-1 py-0.5 text-center">判定</th>
          </tr>
        </thead>
        <tbody>
          {model.rows.length === 0 ? (
            <tr>
              <td
                className="border border-gray-300 px-1 py-0.5 text-gray-500"
                colSpan={5}
              >
                項目なし
              </td>
            </tr>
          ) : (
            model.rows.map((row) => (
              <tr key={row.id}>
                <td className="border border-gray-300 px-1 py-0.5">{row.name}</td>
                <td
                  className={`border border-gray-300 px-1 py-0.5 text-right${
                    row.isAbnormal ? " font-semibold" : ""
                  }`}
                >
                  {row.inspectionValue}
                </td>
                <td className="border border-gray-300 px-1 py-0.5">{row.unit}</td>
                <td className="border border-gray-300 px-1 py-0.5">
                  {row.referenceValue}
                </td>
                <td className="border border-gray-300 px-1 py-0.5 text-center">
                  {row.statusLabel}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </PrintPortal>
  );
}
