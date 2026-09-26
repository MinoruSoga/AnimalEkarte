import type { ComponentProps } from "react";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { MedicalRecordInterview } from "./MedicalRecordInterview";
import {
  DEFAULT_CHIEF_COMPLAINT,
  DEFAULT_TREATMENT_POLICY,
} from "../hooks/use-medical-record-form-model";
import type { InterviewHistoryItem } from "../types";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true }),
}));

vi.mock("../api/get-chief-complaint-types", () => ({
  useGetChiefComplaintTypes: () => ({ data: [], isLoading: false }),
}));

type InterviewProps = ComponentProps<typeof MedicalRecordInterview>;

const COPY_HISTORY_ITEMS: InterviewHistoryItem[] = [
  {
    id: "10",
    date: "2026/02/01 (日)",
    author: "田中",
    type: "再診",
    title: "前回カルテ",
    content: "元気がない",
    copySource: {
      chiefComplaint: "前回の主訴詳細",
      treatmentPolicy: "前回の治療方針",
      chiefComplaintTypeId: 7,
    },
  },
];

function renderInterview(overrides: Partial<InterviewProps> = {}) {
  const props: InterviewProps = {
    chiefComplaint: DEFAULT_CHIEF_COMPLAINT,
    setChiefComplaint: vi.fn(),
    chiefComplaintTypeId: null,
    setChiefComplaintTypeId: vi.fn(),
    treatmentPolicy: DEFAULT_TREATMENT_POLICY,
    setTreatmentPolicy: vi.fn(),
    historyItems: COPY_HISTORY_ITEMS,
    ...overrides,
  };
  render(
    <MemoryRouter>
      <MedicalRecordInterview {...props} />
    </MemoryRouter>,
  );
  return props;
}

describe("MedicalRecordInterview — form field semantics", () => {
  it("問診内の各form fieldに明示labelとidを接続する", () => {
    renderInterview({ historyItems: [] });

    expect(screen.getByRole("combobox", { name: "主訴区分" })).toHaveAttribute(
      "id",
      "medical-record-chief-complaint-type",
    );
    expect(screen.getByRole("textbox", { name: "主訴詳細" })).toHaveAttribute(
      "name",
      "chiefComplaint",
    );
    expect(screen.getByRole("textbox", { name: "治療方針" })).toHaveAttribute(
      "name",
      "treatmentPolicy",
    );
    expect(screen.getByRole("textbox", { name: "過去のカルテを検索" })).toHaveAttribute(
      "name",
      "medicalRecordHistorySearch",
    );
  });
});

describe("MedicalRecordInterview — 前回複写（コピー）", () => {
  it("未編集の既定値では コピー が確認なしで3項目へ即時適用される", async () => {
    const user = userEvent.setup();
    const props = renderInterview();

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(props.setChiefComplaint).toHaveBeenCalledWith("前回の主訴詳細");
    expect(props.setTreatmentPolicy).toHaveBeenCalledWith("前回の治療方針");
    expect(props.setChiefComplaintTypeId).toHaveBeenCalledWith(7);
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  it("主訴詳細が編集済みなら ConfirmDialog の確定で適用される", async () => {
    const user = userEvent.setup();
    const props = renderInterview({ chiefComplaint: "編集済みの主訴" });

    await user.click(screen.getByRole("button", { name: "コピー" }));
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "コピー" }));

    expect(props.setChiefComplaint).toHaveBeenCalledWith("前回の主訴詳細");
    expect(props.setTreatmentPolicy).toHaveBeenCalledWith("前回の治療方針");
    expect(props.setChiefComplaintTypeId).toHaveBeenCalledWith(7);
  });

  it("治療方針が編集済みでも ConfirmDialog のキャンセルは現在値を保持する", async () => {
    const user = userEvent.setup();
    const props = renderInterview({ treatmentPolicy: "独自の治療方針" });

    await user.click(screen.getByRole("button", { name: "コピー" }));
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "キャンセル" }));

    expect(props.setChiefComplaint).not.toHaveBeenCalled();
    expect(props.setTreatmentPolicy).not.toHaveBeenCalled();
    expect(props.setChiefComplaintTypeId).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    });
  });

  it("主訴区分が選択済みでも未編集扱いせず ConfirmDialog を開く", async () => {
    const user = userEvent.setup();
    const props = renderInterview({ chiefComplaintTypeId: 3 });

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(await screen.findByRole("alertdialog")).toBeInTheDocument();
    expect(props.setChiefComplaint).not.toHaveBeenCalled();
    expect(props.setTreatmentPolicy).not.toHaveBeenCalled();
    expect(props.setChiefComplaintTypeId).not.toHaveBeenCalled();
  });

  it("copySource に存在する項目だけを適用する", async () => {
    const user = userEvent.setup();
    const partialItems: InterviewHistoryItem[] = [
      {
        ...COPY_HISTORY_ITEMS[0],
        id: "11",
        copySource: { chiefComplaint: "部分的な主訴" },
      },
    ];
    const props = renderInterview({ historyItems: partialItems });

    await user.click(screen.getByRole("button", { name: "コピー" }));

    expect(props.setChiefComplaint).toHaveBeenCalledWith("部分的な主訴");
    expect(props.setTreatmentPolicy).not.toHaveBeenCalled();
    expect(props.setChiefComplaintTypeId).not.toHaveBeenCalled();
  });
});
