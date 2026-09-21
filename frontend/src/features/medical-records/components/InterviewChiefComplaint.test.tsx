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

describe("InterviewChiefComplaint chief_complaint_type unset/clear", () => {
  beforeEach(() => {
    mockUsePermission.mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    });
  });

  it("N unset: null type shows placeholder, no clear affordance, detail stays editable", async () => {
    const user = userEvent.setup();

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
    expect(typeSelect).toHaveTextContent("選択してください");
    expect(typeSelect).not.toBeRequired();
    expect(screen.getByLabelText("主訴詳細")).toHaveValue("食欲低下");
    expect(screen.getByLabelText("主訴詳細")).not.toBeDisabled();

    await user.click(typeSelect);
    expect(screen.queryByRole("option", { name: "選択をクリア" })).not.toBeInTheDocument();
  });

  it("C intentional clear: real clear affordance emits empty and maps to null setter", async () => {
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

    const typeSelect = screen.getByRole("combobox", { name: "主訴区分" });
    expect(typeSelect).toHaveTextContent("消化器");

    await user.click(typeSelect);
    await user.click(screen.getByRole("option", { name: "選択をクリア" }));
    expect(setChiefComplaintTypeId).toHaveBeenCalledWith(null);
  });

  it("reload hydrate null: parent-driven null shows empty without clear click", () => {
    const setChiefComplaintTypeId = vi.fn();
    const props = {
      chiefComplaint: "食欲低下",
      setChiefComplaint: vi.fn(),
      setChiefComplaintTypeId,
      templates: [] as { label: string; text: string }[],
      onInsertTemplate: vi.fn(),
    };

    const { rerender } = render(<InterviewChiefComplaint {...props} chiefComplaintTypeId={5} />);
    expect(screen.getByRole("combobox", { name: "主訴区分" })).toHaveTextContent("消化器");

    rerender(<InterviewChiefComplaint {...props} chiefComplaintTypeId={null} />);
    expect(screen.getByRole("combobox", { name: "主訴区分" })).toHaveTextContent(
      "選択してください",
    );
    expect(setChiefComplaintTypeId).not.toHaveBeenCalled();
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

    await user.click(screen.getByRole("combobox", { name: "主訴区分" }));
    await user.click(screen.getByRole("option", { name: "消化器" }));
    expect(setChiefComplaintTypeId).toHaveBeenCalledWith(5);
  });
});
