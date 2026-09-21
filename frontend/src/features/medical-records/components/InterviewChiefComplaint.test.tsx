import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { InterviewChiefComplaint } from "./InterviewChiefComplaint";

const mockUsePermission = vi.hoisted(() => vi.fn());

vi.mock("@/hooks/use-permission", () => ({
  usePermission: mockUsePermission,
}));

vi.mock("react-router", () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock("../api/get-chief-complaint-types", () => ({
  useGetChiefComplaintTypes: () => ({
    data: [
      { id: 5, name: "消化器" },
      { id: 9, name: "皮膚" },
    ],
    isLoading: false,
  }),
}));

/** Expose empty-value path without depending on SearchableSelect clear UI (C0). */
vi.mock("@/components/ui/searchable-select", () => ({
  SearchableSelect: ({
    id,
    value,
    onValueChange,
    placeholder,
    disabled,
  }: {
    id?: string;
    value: string;
    onValueChange: (value: string) => void;
    placeholder?: string;
    disabled?: boolean;
  }) => (
    <div>
      <button
        type="button"
        id={id}
        role="combobox"
        aria-label="主訴区分"
        disabled={disabled}
        data-value={value}
        data-placeholder={placeholder}
      >
        {value || placeholder}
      </button>
      <button type="button" onClick={() => onValueChange("5")}>
        select-type-5
      </button>
      <button type="button" onClick={() => onValueChange("")}>
        clear-type
      </button>
    </div>
  ),
}));

describe("InterviewChiefComplaint chief_complaint_type unset/clear", () => {
  beforeEach(() => {
    mockUsePermission.mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    });
  });

  it("N unset: null type renders empty select value and keeps detail editable without required", () => {
    render(
      <InterviewChiefComplaint
        chiefComplaint="食欲低下"
        setChiefComplaint={vi.fn()}
        chiefComplaintTypeId={null}
        setChiefComplaintTypeId={vi.fn()}
        templates={[]}
        onInsertTemplate={vi.fn()}
      />,
    );

    const typeSelect = screen.getByRole("combobox", { name: "主訴区分" });
    expect(typeSelect).toHaveAttribute("data-value", "");
    expect(typeSelect).not.toBeRequired();
    expect(screen.getByLabelText("主訴詳細")).toHaveValue("食欲低下");
    expect(screen.getByLabelText("主訴詳細")).not.toBeDisabled();
  });

  it("C intentional clear receive path: empty onValueChange maps to null setter", async () => {
    const user = userEvent.setup();
    const setChiefComplaintTypeId = vi.fn();

    render(
      <InterviewChiefComplaint
        chiefComplaint="食欲低下"
        setChiefComplaint={vi.fn()}
        chiefComplaintTypeId={5}
        setChiefComplaintTypeId={setChiefComplaintTypeId}
        templates={[]}
        onInsertTemplate={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "clear-type" }));
    expect(setChiefComplaintTypeId).toHaveBeenCalledWith(null);
  });

  it("selecting a type maps string id to number setter", async () => {
    const user = userEvent.setup();
    const setChiefComplaintTypeId = vi.fn();

    render(
      <InterviewChiefComplaint
        chiefComplaint=""
        setChiefComplaint={vi.fn()}
        chiefComplaintTypeId={null}
        setChiefComplaintTypeId={setChiefComplaintTypeId}
        templates={[]}
        onInsertTemplate={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "select-type-5" }));
    expect(setChiefComplaintTypeId).toHaveBeenCalledWith(5);
  });
});
