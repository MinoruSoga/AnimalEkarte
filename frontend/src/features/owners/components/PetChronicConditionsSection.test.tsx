import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { PetChronicCondition } from "@/types/generated/models";
import type { UsePermissionResult } from "@/hooks/use-permission";

import { PetChronicConditionsSection } from "./PetChronicConditionsSection";

const mocks = vi.hoisted(() => {
  const defaultConditions: PetChronicCondition[] = [
    {
      id: 1,
      clinic_id: 1,
      pet_id: 7,
      condition_code: "CKD",
      condition_name: "慢性腎臓病",
      diagnosed_at: "2024-03-01",
      notes: "定期検査",
      is_active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 2,
      clinic_id: 1,
      pet_id: 7,
      condition_code: "DM",
      condition_name: "糖尿病",
      diagnosed_at: "2023-11-10",
      is_active: false,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ];
  return {
    defaultConditions,
    conditions: defaultConditions,
    conditionsError: false,
    permissions: {
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    } as UsePermissionResult,
    createMutateAsync: vi.fn(),
    updateMutateAsync: vi.fn(),
    deleteMutateAsync: vi.fn(),
  };
});

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), warning: vi.fn(), error: vi.fn() },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mocks.permissions,
}));

vi.mock("@/components/shared/DatePicker", () => ({
  DatePicker: ({
    id,
    value,
    onChange,
    placeholder,
  }: {
    id?: string;
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
  }) => (
    <input
      id={id}
      value={value}
      placeholder={placeholder}
      onChange={(event) => onChange(event.target.value)}
    />
  ),
}));

vi.mock("../api/pet-chronic-conditions", () => ({
  useGetPetChronicConditions: () => ({
    // React Query は再取得失敗時も既存 data を保持する — data は常に返す
    data: mocks.conditions,
    isLoading: false,
    error: mocks.conditionsError ? new Error("load failed") : null,
  }),
  useCreatePetChronicCondition: () => ({
    mutateAsync: mocks.createMutateAsync,
    isPending: false,
  }),
  useUpdatePetChronicCondition: () => ({
    mutateAsync: mocks.updateMutateAsync,
    isPending: false,
  }),
  useDeletePetChronicCondition: () => ({
    mutateAsync: mocks.deleteMutateAsync,
    isPending: false,
  }),
}));

function renderSection(element: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(<QueryClientProvider client={queryClient}>{element}</QueryClientProvider>);
}

describe("PetChronicConditionsSection", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.conditions = mocks.defaultConditions;
    mocks.conditionsError = false;
    mocks.permissions = {
      canView: true,
      canCreate: true,
      canEdit: true,
      canDelete: true,
    };
    mocks.createMutateAsync.mockResolvedValue(undefined);
    mocks.updateMutateAsync.mockResolvedValue(undefined);
    mocks.deleteMutateAsync.mockResolvedValue(undefined);
  });

  it("登録済みの慢性疾患を一覧表示する", () => {
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    expect(screen.getByRole("heading", { name: "慢性疾患" })).toBeInTheDocument();
    expect(screen.getByText("慢性腎臓病")).toBeInTheDocument();
    expect(screen.getByText(/CKD/)).toBeInTheDocument();
    expect(screen.getByText(/2024-03-01/)).toBeInTheDocument();
    expect(screen.getByText("糖尿病")).toBeInTheDocument();
    expect(screen.getByText("無効")).toBeInTheDocument();
    expect(screen.getByText("有効")).toBeInTheDocument();
  });

  it("未登録なら空メッセージを表示する", () => {
    mocks.conditions = [];
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    expect(screen.getByText("慢性疾患は登録されていません。")).toBeInTheDocument();
  });

  it("追加フォームで登録すると POST mutation を発行する", async () => {
    const user = userEvent.setup();
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    await user.type(screen.getByLabelText("疾患名"), " 高血圧 ");
    await user.type(screen.getByLabelText("疾患コード"), " HTN ");
    await user.type(screen.getByLabelText("診断日"), "2024-05-10");
    await user.type(screen.getByLabelText("備考"), " 経過観察 ");
    await user.click(screen.getByRole("button", { name: "慢性疾患を追加" }));

    await waitFor(() =>
      expect(mocks.createMutateAsync).toHaveBeenCalledWith({
        petId: "7",
        request: {
          condition_code: "HTN",
          condition_name: "高血圧",
          diagnosed_at: "2024-05-10",
          notes: "経過観察",
          is_active: true,
        },
      }),
    );
  });

  it("編集開始で既存値をフォームへ読み込み、保存で PATCH mutation を発行する", async () => {
    const user = userEvent.setup();
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    await user.click(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を編集" }));

    const nameInput = screen.getByLabelText("疾患名（慢性腎臓病）");
    expect(nameInput).toHaveValue("慢性腎臓病");
    expect(screen.getByLabelText("疾患コード（慢性腎臓病）")).toHaveValue("CKD");
    expect(screen.getByLabelText("診断日（慢性腎臓病）")).toHaveValue("2024-03-01");
    expect(screen.getByLabelText("備考（慢性腎臓病）")).toHaveValue("定期検査");
    expect(screen.getByRole("checkbox", { name: "有効（慢性腎臓病）" })).toBeChecked();

    await user.clear(nameInput);
    await user.type(nameInput, "慢性腎臓病（進行）");
    await user.click(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を保存" }));

    await waitFor(() =>
      expect(mocks.updateMutateAsync).toHaveBeenCalledWith({
        petId: "7",
        conditionId: 1,
        request: {
          condition_code: "CKD",
          condition_name: "慢性腎臓病（進行）",
          diagnosed_at: "2024-03-01",
          notes: "定期検査",
          is_active: true,
        },
      }),
    );
  });

  it("編集で有効チェックを外すと is_active:false を明示送信する", async () => {
    const user = userEvent.setup();
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    await user.click(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を編集" }));
    await user.click(screen.getByRole("checkbox", { name: "有効（慢性腎臓病）" }));
    await user.click(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を保存" }));

    await waitFor(() =>
      expect(mocks.updateMutateAsync).toHaveBeenCalledWith({
        petId: "7",
        conditionId: 1,
        request: expect.objectContaining({ is_active: false }),
      }),
    );
  });

  it("編集キャンセルで一覧表示へ戻る", async () => {
    const user = userEvent.setup();
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    await user.click(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を編集" }));
    await user.click(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病の編集をキャンセル" }));

    expect(screen.queryByLabelText("疾患名（慢性腎臓病）")).not.toBeInTheDocument();
    expect(mocks.updateMutateAsync).not.toHaveBeenCalled();
  });

  it("削除ボタンで DELETE mutation を発行する", async () => {
    const user = userEvent.setup();
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    await user.click(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を削除" }));

    await waitFor(() =>
      expect(mocks.deleteMutateAsync).toHaveBeenCalledWith({ petId: "7", conditionId: 1 }),
    );
  });

  it("owners:view 権限がなければ何も描画しない", () => {
    mocks.permissions = { ...mocks.permissions, canView: false };
    const { container } = renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    expect(container).toBeEmptyDOMElement();
  });

  it("owners:create がなければ追加フォームの入力とボタンを disabled にする", () => {
    mocks.permissions = { ...mocks.permissions, canCreate: false };
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    expect(screen.getByLabelText("疾患名")).toBeDisabled();
    expect(screen.getByLabelText("疾患コード")).toBeDisabled();
    expect(screen.getByLabelText("診断日")).toBeDisabled();
    expect(screen.getByLabelText("備考")).toBeDisabled();
    expect(screen.getByRole("button", { name: "慢性疾患を追加" })).toBeDisabled();
  });

  it("owners:edit がなければ編集ボタンを disabled にする", () => {
    mocks.permissions = { ...mocks.permissions, canEdit: false };
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    expect(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を編集" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を削除" })).toBeEnabled();
  });

  it("owners:delete がなければ削除ボタンを disabled にする", () => {
    mocks.permissions = { ...mocks.permissions, canDelete: false };
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    expect(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を削除" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を編集" })).toBeEnabled();
  });

  it("canEdit=false では追加・編集・削除の操作を全て disabled にする", () => {
    renderSection(<PetChronicConditionsSection petId="7" canEdit={false} />);

    expect(screen.getByRole("button", { name: "慢性疾患を追加" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を編集" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を削除" })).toBeDisabled();
  });

  it("一覧取得に失敗した場合はエラーを表示し追加フォームを disabled にする", () => {
    mocks.conditionsError = true;
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    expect(screen.getByRole("alert")).toHaveTextContent("慢性疾患情報を取得できませんでした。");
    expect(screen.getByRole("button", { name: "慢性疾患を追加" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "慢性疾患 慢性腎臓病を編集" })).toBeDisabled();
  });

  it("疾患名が空ならエラーを表示して送信しない", async () => {
    const user = userEvent.setup();
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    await user.type(screen.getByLabelText("疾患コード"), "CKD");
    await user.type(screen.getByLabelText("診断日"), "2024-05-10");
    await user.click(screen.getByRole("button", { name: "慢性疾患を追加" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("疾患名を入力してください。");
    expect(mocks.createMutateAsync).not.toHaveBeenCalled();
  });

  it("登録成功を支援技術へ通知する", async () => {
    const user = userEvent.setup();
    renderSection(<PetChronicConditionsSection petId="7" canEdit />);

    await user.type(screen.getByLabelText("疾患名"), "高血圧");
    await user.type(screen.getByLabelText("疾患コード"), "HTN");
    await user.type(screen.getByLabelText("診断日"), "2024-05-10");
    await user.click(screen.getByRole("button", { name: "慢性疾患を追加" }));

    expect(await screen.findByRole("status")).toHaveTextContent("慢性疾患を登録しました");
  });
});
