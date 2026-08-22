import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { InventoryForm } from "./InventoryForm";

const { permission, formReturn } = vi.hoisted(() => ({
  permission: { current: { canView: true, canCreate: false, canEdit: false, canDelete: false } },
  formReturn: {
    current: {
      isEdit: true,
      isLoading: false,
      existingItem: {
        name: "留置針",
        category: "consumable",
        quantity: 0,
        unit: "本",
        minStockLevel: 50,
        location: "処置室",
      },
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
  useInventoryForm: () => formReturn.current,
}));
vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

function renderForm() {
  return render(
    <MemoryRouter initialEntries={["/inventory/7"]}>
      <Routes>
        <Route path="/inventory/:id" element={<InventoryForm />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("InventoryForm RBAC", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: false, canEdit: false, canDelete: false };
    formReturn.current = {
      ...formReturn.current,
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
});

describe("InventoryForm inline validation (BUG-009)", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: false };
    formReturn.current = {
      ...formReturn.current,
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
