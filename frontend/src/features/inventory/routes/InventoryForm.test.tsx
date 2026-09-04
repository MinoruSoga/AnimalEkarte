import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { InventoryForm } from "./InventoryForm";

const { permission, formReturn, useInventoryFormMock } = vi.hoisted(() => ({
  permission: { current: { canView: true, canCreate: false, canEdit: false, canDelete: false } },
  useInventoryFormMock: vi.fn(),
  formReturn: {
    current: {
      isEdit: true,
      isLoading: false,
      isReadNotFound: false,
      isReadError: false,
      retryRead: undefined as undefined | (() => void),
      existingItem: {
        name: "留置針",
        category: "consumable",
        quantity: 0,
        unit: "本",
        minStockLevel: 50,
        location: "処置室",
      } as
        | {
            name: string;
            category: string;
            quantity: number;
            unit: string;
            minStockLevel: number;
            location: string;
          }
        | undefined,
      category: "consumable",
      setCategory: vi.fn(),
      resolvedExpiry: "",
      setExpiryDate: vi.fn(),
      resolvedLastRestocked: "",
      setLastRestocked: vi.fn(),
      formAction: vi.fn(),
      formState: { success: false, timestamp: 0, fieldErrors: {} },
      isPending: false,
    },
  },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => permission.current,
}));
vi.mock("../hooks/use-inventory-form", () => ({
  useInventoryForm: (...args: unknown[]) => {
    useInventoryFormMock(...args);
    return formReturn.current;
  },
}));
vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

function renderForm(initialEntry = "/inventory/7") {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/inventory/new" element={<InventoryForm />} />
        <Route path="/inventory/:id" element={<InventoryForm />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("InventoryForm BUG-507 not-found / network gate", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: false };
    formReturn.current = {
      ...formReturn.current,
      isEdit: true,
      isLoading: false,
      isReadNotFound: false,
      isReadError: false,
      retryRead: undefined,
      existingItem: {
        name: "留置針",
        category: "consumable",
        quantity: 0,
        unit: "本",
        minStockLevel: 50,
        location: "処置室",
      },
      formState: { success: false, timestamp: 0, fieldErrors: {} },
      formAction: vi.fn(),
    };
  });

  it("不在ID: 在庫情報が見つかりません を表示し、空の在庫編集フォームと更新ボタンを出さない", () => {
    formReturn.current = {
      ...formReturn.current,
      isReadNotFound: true,
      existingItem: undefined,
    };

    renderForm("/inventory/999999001");

    expect(screen.getByText("在庫情報が見つかりません")).toBeInTheDocument();
    expect(screen.queryByText("在庫編集")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
    expect(document.querySelector("form#inventory-form")).not.toBeInTheDocument();
  });

  it("403 相当: 404 と同一の非開示メッセージ", () => {
    formReturn.current = {
      ...formReturn.current,
      isReadNotFound: true,
      existingItem: undefined,
    };

    renderForm("/inventory/42");

    expect(screen.getByText("在庫情報が見つかりません")).toBeInTheDocument();
    expect(document.querySelector("form#inventory-form")).not.toBeInTheDocument();
  });

  it("network error: 取得失敗 + 再試行、blank form ではない", () => {
    const retry = vi.fn();
    formReturn.current = {
      ...formReturn.current,
      isReadError: true,
      retryRead: retry,
      existingItem: undefined,
    };

    renderForm("/inventory/999999001");

    expect(screen.getByText("在庫情報の取得に失敗しました")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "再試行" })).toBeInTheDocument();
    screen.getByRole("button", { name: "再試行" }).click();
    expect(retry).toHaveBeenCalledTimes(1);
    expect(document.querySelector("form#inventory-form")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
  });

  it("既存ID: 在庫編集フォームをデータ付きで開く", () => {
    renderForm("/inventory/7");

    expect(screen.getByText("在庫編集")).toBeInTheDocument();
    expect(document.querySelector("form#inventory-form")).toBeInTheDocument();
    expect(screen.queryByText("在庫情報が見つかりません")).not.toBeInTheDocument();
  });
});

describe("InventoryForm RBAC", () => {
  beforeEach(() => {
    useInventoryFormMock.mockClear();
    permission.current = { canView: true, canCreate: false, canEdit: false, canDelete: false };
    formReturn.current = {
      ...formReturn.current,
      isEdit: true,
      isLoading: false,
      isReadNotFound: false,
      isReadError: false,
      retryRead: undefined,
      formState: { success: false, timestamp: 0, fieldErrors: {} },
    };
  });

  it("edit権限がない場合は閲覧専用banner・disabled fieldset・更新buttonなし", () => {
    renderForm();

    expect(screen.getByRole("status", { name: "閲覧専用モード" })).toHaveTextContent(
      "編集権限がないため変更できません",
    );
    expect(document.querySelector("fieldset")).toBeDisabled();
    expect(screen.queryByRole("button", { name: "更新" })).not.toBeInTheDocument();
  });

  it("usePermission の canCreate/canEdit を useInventoryForm の permissions に渡す", () => {
    permission.current = { canView: true, canCreate: true, canEdit: false, canDelete: false };
    renderForm("/inventory/7");

    expect(useInventoryFormMock).toHaveBeenCalledWith("7", {
      permissions: { canCreate: true, canEdit: false },
    });
  });

  it("新規作成ルートでも canCreate/canEdit を permissions に渡す", () => {
    permission.current = { canView: true, canCreate: false, canEdit: true, canDelete: false };
    formReturn.current = { ...formReturn.current, isEdit: false };
    renderForm("/inventory/new");

    expect(useInventoryFormMock).toHaveBeenCalledWith(undefined, {
      permissions: { canCreate: false, canEdit: true },
    });
  });
});

describe("InventoryForm inline validation (BUG-009)", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: false };
    formReturn.current = {
      ...formReturn.current,
      isEdit: true,
      formState: { success: false, timestamp: 0, fieldErrors: {} },
    };
  });

  it("form は noValidate で HTML5 制約が JS fieldErrors より先にインターセプトしない", () => {
    renderForm();

    expect(document.querySelector("form")).toHaveAttribute("novalidate");
  });

  it("必須未入力の fieldErrors を inline 表示する", () => {
    formReturn.current = {
      ...formReturn.current,
      formState: {
        success: false,
        timestamp: 0,
        fieldErrors: {
          name: "品名を入力してください",
          unit: "単位を入力してください",
        },
      },
    };

    renderForm();

    expect(screen.getByText("品名を入力してください")).toBeInTheDocument();
    expect(screen.getByText("単位を入力してください")).toBeInTheDocument();
  });

  it("負値の fieldErrors を inline 表示する", () => {
    formReturn.current = {
      ...formReturn.current,
      formState: {
        success: false,
        timestamp: 0,
        fieldErrors: {
          quantity: "現在庫数は0以上の整数で入力してください",
          minStockLevel: "最低在庫数は0以上の整数で入力してください",
        },
      },
    };

    renderForm();

    expect(screen.getByText("現在庫数は0以上の整数で入力してください")).toBeInTheDocument();
    expect(screen.getByText("最低在庫数は0以上の整数で入力してください")).toBeInTheDocument();
  });
});

describe("InventoryForm header submit (BUG-007)", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: false };
    formReturn.current = {
      ...formReturn.current,
      isEdit: true,
      formAction: vi.fn(),
      formState: { success: false, timestamp: 0, fieldErrors: {} },
    };
  });

  it("header SubmitButton is associated with inventory-form (BUG-007)", async () => {
    const user = userEvent.setup();
    renderForm();
    const form = document.querySelector("form#inventory-form");
    expect(form).toHaveAttribute("id", "inventory-form");
    const buttons = screen.getAllByRole("button", { name: "更新" });
    expect(buttons.length).toBeGreaterThanOrEqual(2);
    const header = buttons.find((b) => b.getAttribute("form") === "inventory-form");
    expect(header).toBeDefined();
    expect(header).toHaveAttribute("type", "submit");
    expect(header).toHaveAttribute("form", "inventory-form");
    expect(form?.contains(header!)).toBe(false);

    await user.click(header!);
    expect(formReturn.current.formAction).toHaveBeenCalled();
  });
});
