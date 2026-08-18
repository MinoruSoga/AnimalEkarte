import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { usePermission } from "@/hooks/use-permission";
import { ResourceMedicalRecords } from "@/types/generated/models";
import { CheckupForm } from "./CheckupForm";
import { useCheckupForm } from "../hooks/use-checkup-form";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(),
}));

vi.mock("../hooks/use-checkup-form", () => ({
  useCheckupForm: vi.fn(),
}));

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllCheckupTypes: () => ({ data: [{ id: 1, name: "定期健診" }] }),
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: () => ({ data: [{ id: "10", name: "獣医師A", isActive: true }] }),
}));

vi.mock("../api/get-checkups", () => ({
  useGetCheckups: () => ({ data: { data: [], total: 0, page: 1, limit: 100 }, isLoading: false }),
}));

function renderForm() {
  return render(
    <MemoryRouter>
      <CheckupForm />
    </MemoryRouter>,
  );
}

const formActionMock = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useCheckupForm).mockReturnValue({
    pet: {
      id: "pet-1",
      ownerId: "owner-1",
      ownerName: "山田 太郎",
      name: "ポチ",
      petNumber: "0001",
      weight: "5.0kg",
    },
    isPetLoading: false,
    form: {
      checkupTypeId: "",
      date: "",
      nextScheduleType: "1year",
      nextDate: "",
      doctorId: "",
      result: "",
    },
    formAction: formActionMock,
    formState: { success: false, timestamp: 0 },
    isPending: false,
    fieldErrors: {},
    checkupFields: [],
    fieldValues: {},
    setFieldValue: vi.fn(),
    setCheckupTypeId: vi.fn(),
    setDate: vi.fn(),
    setNextScheduleType: vi.fn(),
    setNextDate: vi.fn(),
    setDoctorId: vi.fn(),
    setResult: vi.fn(),
  } as ReturnType<typeof useCheckupForm>);
});

describe("CheckupForm permissions", () => {
  it("作成権限がない場合は入力fieldsetと保存操作を実際にdisabledにする", () => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: false,
      canEdit: false,
      canDelete: false,
    });

    renderForm();

    expect(usePermission).toHaveBeenCalledWith(ResourceMedicalRecords);
    expect(useCheckupForm).toHaveBeenCalledWith({
      canCreate: false,
      canEdit: false,
    });
    expect(screen.getByText("閲覧のみ")).toBeInTheDocument();
    expect(screen.getByRole("group", { name: "定期健診入力" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
    expect(screen.getByRole("textbox", { name: "結果・所見" })).toBeDisabled();
    const inputGroup = screen.getByRole("group", { name: "定期健診入力" });
    for (const combobox of inputGroup.querySelectorAll('[role="combobox"]')) {
      expect(combobox).toBeDisabled();
    }

    fireEvent.submit(screen.getByRole("form", { name: "定期健診登録フォーム" }));
    expect(formActionMock).not.toHaveBeenCalled();
  });

  it.each([
    { canCreate: true, canEdit: false },
    { canCreate: false, canEdit: true },
  ])(
    "medical-records:create/editの片方だけでは入力とsubmitをfail-closedにする ($canCreate/$canEdit)",
    ({ canCreate, canEdit }) => {
      vi.mocked(usePermission).mockReturnValue({
        canView: true,
        canCreate,
        canEdit,
        canDelete: false,
      });

      renderForm();

      expect(useCheckupForm).toHaveBeenCalledWith({
        canCreate,
        canEdit,
      });
      expect(screen.getByRole("group", { name: "定期健診入力" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();

      fireEvent.submit(screen.getByRole("form", { name: "定期健診登録フォーム" }));
      expect(formActionMock).not.toHaveBeenCalled();
    },
  );

  it("作成・編集権限が両方ある場合は入力fieldsetと保存操作を有効のまま保つ", () => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: false,
    });

    renderForm();

    expect(useCheckupForm).toHaveBeenCalledWith({
      canCreate: true,
      canEdit: true,
    });
    expect(screen.getByRole("group", { name: "定期健診入力" })).not.toBeDisabled();
    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
    expect(screen.getByRole("textbox", { name: "結果・所見" })).toBeEnabled();
  });
});
