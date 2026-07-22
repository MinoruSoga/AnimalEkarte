// React/Framework
import { memo } from "react";

// Internal
import { C } from "@/lib/design-tokens";

// Relative
import type { Treatment } from "../types";

interface ClinicInfo {
  name?: string;
  address?: string;
  phone?: string;
}

interface PetInfo {
  name: string;
  species?: string;
  ownerName?: string;
}

interface MedicalRecordPrintViewProps {
  recordNo?: string;
  date?: string;
  doctorName?: string;
  pet: PetInfo;
  clinic: ClinicInfo;
  chiefComplaint?: string;
  treatmentPolicy?: string;
  physicalExam?: string;
  diagnosisDetails?: string;
  treatments?: ReadonlyArray<Treatment>;
}

const EMPTY_TREATMENTS: ReadonlyArray<Treatment> = [];

/**
 * カルテ印刷レイアウト。
 * `window.print()` 呼び出し時に表示する。
 * 親要素で `hidden print:block` を付与してスクリーン表示を制御する。
 */
export const MedicalRecordPrintView = memo(function MedicalRecordPrintView({
  recordNo,
  date,
  doctorName,
  pet,
  clinic,
  chiefComplaint,
  treatmentPolicy,
  physicalExam,
  diagnosisDetails,
  treatments = EMPTY_TREATMENTS,
}: MedicalRecordPrintViewProps) {
  const medicineTreatments = treatments.filter((t) => t.item_type === "medicine" && t.is_selected);

  return (
    <div className="font-sans text-[12pt] text-black w-full max-w-[210mm] mx-auto p-8">
      {/* ── ヘッダー: 病院情報 ── */}
      <div className="flex justify-between items-start mb-6 pb-4 border-b border-black">
        <div>
          <h1 className="text-[16pt] font-bold">{clinic.name ?? "動物病院"}</h1>
          {clinic.address ? <p className="text-[10pt] mt-0.5">{clinic.address}</p> : null}
          {clinic.phone ? <p className="text-[10pt]">TEL: {clinic.phone}</p> : null}
        </div>
        <div className="text-right">
          <h2 className="text-[14pt] font-bold">診療カルテ</h2>
          {date ? <p className="text-[10pt] mt-1">診療日: {date}</p> : null}
          {recordNo ? <p className="text-[10pt]">カルテ番号: {recordNo}</p> : null}
        </div>
      </div>

      {/* ── 患者情報 ── */}
      <div className="mb-6 grid grid-cols-2 gap-4">
        <div className={`border ${C.borderMedium} rounded p-3`}>
          <h3 className="text-[11pt] font-semibold mb-2">患者情報</h3>
          <table className="w-full text-[10pt]">
            <tbody>
              <tr>
                <td className={`w-24 ${C.text60}`}>ペット名</td>
                <td className="font-medium">{pet.name}</td>
              </tr>
              {pet.species ? (
                <tr>
                  <td className={C.text60}>種別</td>
                  <td>{pet.species}</td>
                </tr>
              ) : null}
              {pet.ownerName ? (
                <tr>
                  <td className={C.text60}>飼主名</td>
                  <td>{pet.ownerName}</td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
        <div className={`border ${C.borderMedium} rounded p-3`}>
          <h3 className="text-[11pt] font-semibold mb-2">担当医</h3>
          <p className="text-[10pt] font-medium">{doctorName ?? "—"}</p>
        </div>
      </div>

      {/* ── 主訴 ── */}
      {chiefComplaint ? (
        <section className="mb-5">
          <h3 className={`text-[11pt] font-semibold border-b ${C.borderLight} pb-1 mb-2`}>主訴</h3>
          <p className="text-[10pt] whitespace-pre-wrap">{chiefComplaint}</p>
        </section>
      ) : null}

      {/* ── 身体検査所見 ── */}
      {physicalExam ? (
        <section className="mb-5">
          <h3 className={`text-[11pt] font-semibold border-b ${C.borderLight} pb-1 mb-2`}>身体検査所見</h3>
          <p className="text-[10pt] whitespace-pre-wrap">{physicalExam}</p>
        </section>
      ) : null}

      {/* ── 診断 ── */}
      {diagnosisDetails ? (
        <section className="mb-5">
          <h3 className={`text-[11pt] font-semibold border-b ${C.borderLight} pb-1 mb-2`}>診断</h3>
          <p className="text-[10pt] whitespace-pre-wrap">{diagnosisDetails}</p>
        </section>
      ) : null}

      {/* ── 治療方針 ── */}
      {treatmentPolicy ? (
        <section className="mb-5">
          <h3 className={`text-[11pt] font-semibold border-b ${C.borderLight} pb-1 mb-2`}>治療方針</h3>
          <p className="text-[10pt] whitespace-pre-wrap">{treatmentPolicy}</p>
        </section>
      ) : null}

      {/* ── 処方箋 ── */}
      {medicineTreatments.length > 0 ? (
        <section className="mb-5">
          <h3 className={`text-[11pt] font-semibold border-b ${C.borderLight} pb-1 mb-2`}>処方薬</h3>
          <table className="w-full text-[10pt] border-collapse">
            <thead>
              <tr className={`border-b ${C.borderMedium}`}>
                <th className="text-left py-1 pr-4 font-semibold">薬品名</th>
                <th className="text-right py-1 pr-4 font-semibold w-20">数量</th>
                <th className="text-left py-1 font-semibold">備考 (投与方法等)</th>
              </tr>
            </thead>
            <tbody>
              {medicineTreatments.map((t) => (
                <tr key={t.id} className={`border-b ${C.borderLight}`}>
                  <td className="py-1 pr-4">{t.content}</td>
                  <td className="py-1 pr-4 text-right">{t.quantity}</td>
                  <td className="py-1">{t.memo ?? ""}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      ) : null}

      {/* ── フッター ── */}
      <div className={`mt-12 pt-4 border-t ${C.borderLight} text-[9pt] ${C.text50} flex justify-between`}>
        <span>{clinic.name}</span>
        <span>獣医師: {doctorName ?? "—"}</span>
      </div>
    </div>
  );
});
