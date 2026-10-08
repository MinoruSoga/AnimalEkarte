import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
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
  useGetStaffs: () => ({
    data: [{ id: "10", name: "獣医師A", isActive: true, staffType: "doctor" }],
  }),
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

describe("CheckupForm 死亡ペット render側二重防壁 (FE-RC-004)", () => {
  beforeEach(() => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: false,
    });
  });

  it("死亡ペットでは SubmitButton を非表示にし、理由を表示する（callback ガードとの二重防壁）", () => {
    vi.mocked(useCheckupForm).mockReturnValue({
      pet: {
        id: "pet-1",
        ownerId: "owner-1",
        ownerName: "山田 太郎",
        name: "ポチ",
        petNumber: "0001",
        weight: "5.0kg",
        status: "死亡",
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

    renderForm();

    // render 側防壁: SubmitButton 自体をレンダリングしない（disabled だけにしない）
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.getByText("死亡したペットの定期健診記録は保存できません")).toBeInTheDocument();
    expect(screen.getByText("【死亡】")).toBeInTheDocument();

    // callback 側防壁: 何らかの理由で form submit が発火しても action を呼ばない
    fireEvent.submit(screen.getByRole("form", { name: "定期健診登録フォーム" }));
    expect(formActionMock).not.toHaveBeenCalled();
  });

  it("生存ペットでは SubmitButton を表示し、死亡理由は表示しない", () => {
    renderForm();

    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    expect(
      screen.queryByText("死亡したペットの定期健診記録は保存できません"),
    ).not.toBeInTheDocument();
  });
});

describe("CheckupForm 誕生月バッジ (EMR-301)", () => {
  function renderWithPet(petOverrides: Record<string, unknown>) {
    vi.mocked(useCheckupForm).mockReturnValue({
      pet: {
        id: "pet-1",
        ownerId: "owner-1",
        ownerName: "山田 太郎",
        name: "ポチ",
        petNumber: "0001",
        weight: "5.0kg",
        ...petOverrides,
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

    renderForm();
  }

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: false,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットは PatientInfoCard 上に Cake バッジを表示する", () => {
    renderWithPet({ birthDate: "2018-10-05", status: "生存" });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05"],
    ["birthDate 未登録", undefined],
  ])("%s なら Cake バッジを表示しない", (_label, birthDate) => {
    renderWithPet({ birthDate, status: "生存" });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）には誕生月でも Cake バッジを表示しない", () => {
    renderWithPet({ birthDate: "2018-10-05", status: "死亡" });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

describe("CheckupForm 必須エラー表示 (EMR-319)", () => {
  beforeEach(() => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: false,
    });
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
      fieldErrors: {
        date: "実施日を入力してください",
        checkupTypeId: "健診種別を選択してください",
      },
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

  it("実施日・健診種別の必須エラーで各コントロールが aria-invalid になりエラー文が出る", () => {
    const { container } = renderForm();

    const dateInput = container.querySelector("#checkup-date");
    expect(dateInput).toHaveAttribute("aria-invalid", "true");

    const typeSelect = container.querySelector("#checkup-type-select");
    expect(typeSelect).toHaveAttribute("aria-invalid", "true");

    expect(screen.getByText("実施日を入力してください")).toBeInTheDocument();
    expect(screen.getByText("健診種別を選択してください")).toBeInTheDocument();
  });
});
