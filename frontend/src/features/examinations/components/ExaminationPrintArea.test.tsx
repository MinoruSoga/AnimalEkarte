import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { ExaminationPrintArea } from "./ExaminationPrintArea";
import { buildExaminationPrintModel } from "../lib/examination-print-model";
import type { ExaminationPrintSnapshot } from "../api/get-examination-print-snapshot";

function snapshot(
  overrides: Partial<ExaminationPrintSnapshot> = {},
): ExaminationPrintSnapshot {
  return {
    examinationId: "10",
    clinicId: "1",
    version: 2,
    kind: "official",
    status: "confirmed",
    printBoundary: "official",
    watermark: "",
    date: "2026-08-04",
    resultSummary: "CBC normal",
    machine: "Sysmex",
    examTypeId: "3",
    medicalRecordId: "5",
    petId: "7",
    doctorId: "9",
    display: {
      medicalRecordNo: "MR-1",
      petName: "ポチ",
      medicalRecordOwnerName: "田中",
      petOwnerName: "田中太郎",
      speciesName: "犬",
      examTypeName: "CBC",
      doctorName: "山田",
    },
    items: [
      {
        id: "1",
        examTypeFieldId: "11",
        name: "WBC",
        inspectionValue: "12.5",
        normalValue: "",
        result: "",
        unit: "10^3/uL",
        referenceValue: "6-17",
        refMin: 6,
        refMax: 17,
        qualitativeMin: null,
        qualitativeMax: null,
        isAssessed: true,
        isAbnormal: false,
        status: "normal",
        sortOrder: 1,
      },
    ],
    ...overrides,
  };
}

describe("ExaminationPrintArea (TASK-031)", () => {
  it("renders snapshot values without form/unsaved tokens or danger_reason", () => {
    const model = buildExaminationPrintModel(snapshot());
    render(<ExaminationPrintArea model={model} />);
    const area = screen.getByTestId("examination-print-area");
    expect(within(area).getByText("検査結果")).toBeInTheDocument();
    expect(within(area).getByText(/ポチ/)).toBeInTheDocument();
    expect(within(area).getByText("WBC")).toBeInTheDocument();
    expect(within(area).getByText("12.5")).toBeInTheDocument();
    expect(within(area).getByText("6-17")).toBeInTheDocument();
    expect(area.textContent).not.toContain("danger_reason");
    expect(area.textContent).not.toContain("formItems");
    expect(area.textContent).not.toContain("unsaved");
  });

  it("shows draft watermark for non-confirmed print boundary", () => {
    const model = buildExaminationPrintModel(
      snapshot({
        kind: "working",
        status: "completed",
        printBoundary: "draft",
        watermark: "DRAFT / 未確定",
      }),
    );
    render(<ExaminationPrintArea model={model} />);
    expect(screen.getByTestId("examination-print-watermark")).toHaveTextContent(
      "DRAFT / 未確定",
    );
  });

  it("uses stored is_assessed status labels without FE recalculation", () => {
    const model = buildExaminationPrintModel(
      snapshot({
        items: [
          {
            id: "2",
            examTypeFieldId: null,
            name: "ALT",
            inspectionValue: "999",
            normalValue: "",
            result: "",
            unit: "U/L",
            referenceValue: "10-100",
            refMin: 10,
            refMax: 100,
            qualitativeMin: null,
            qualitativeMax: null,
            isAssessed: false,
            isAbnormal: false,
            status: "high",
            sortOrder: 1,
          },
        ],
      }),
    );
    render(<ExaminationPrintArea model={model} />);
    const area = screen.getByTestId("examination-print-area");
    // isAssessed=false → status label is dash even if status string is high
    expect(within(area).getByText("—")).toBeInTheDocument();
  });
});
