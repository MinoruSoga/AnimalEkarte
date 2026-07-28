import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { MedicalRecordInterview } from "./MedicalRecordInterview";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true }),
}));

vi.mock("../api/get-chief-complaint-types", () => ({
  useGetChiefComplaintTypes: () => ({ data: [], isLoading: false }),
}));

describe("MedicalRecordInterview — form field semantics", () => {
  it("問診内の各form fieldに明示labelとidを接続する", () => {
    render(
      <MemoryRouter>
        <MedicalRecordInterview
          chiefComplaint=""
          setChiefComplaint={vi.fn()}
          chiefComplaintTypeId={null}
          setChiefComplaintTypeId={vi.fn()}
          treatmentPolicy=""
          setTreatmentPolicy={vi.fn()}
          historyItems={[]}
        />
      </MemoryRouter>,
    );

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
