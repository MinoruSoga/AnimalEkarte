import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { C } from "@/lib/design-tokens";

import {
  BasicInfoSection,
  StockInfoSection,
  SupplierInfoSection,
} from "./InventoryFormSections";

function expectResponsiveTwoColumnGrid(heading: string) {
  const grid = screen.getByRole("heading", { name: heading }).nextElementSibling;

  expect(grid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
  expect(grid).not.toHaveClass("grid-cols-2");
}

describe("InventoryFormSections responsive layout", () => {
  it("基本情報はmobileで1列、sm以上で2列になり、品名はsm以上だけ全列を使う", () => {
    render(
      <BasicInfoSection
        defaultName="テスト品"
        defaultUnit="個"
        category="medicine"
        existingCategory={undefined}
        onCategoryChange={vi.fn()}
        onMarkDirty={vi.fn()}
      />,
    );

    expectResponsiveTwoColumnGrid("基本情報");
    const nameField = screen.getByLabelText(/品名/).parentElement;
    expect(nameField).toHaveClass("sm:col-span-2");
    expect(nameField).not.toHaveClass("col-span-2");
  });

  it("在庫情報はmobileで1列、sm以上で2列になる", () => {
    render(
      <StockInfoSection
        defaultQuantity={10}
        defaultMinStockLevel={2}
        defaultLocation="棚A"
        resolvedExpiry="2026-12-31"
        onExpiryChange={vi.fn()}
        onMarkDirty={vi.fn()}
      />,
    );

    expectResponsiveTwoColumnGrid("在庫情報");
  });

  it("仕入先情報はmobileで1列、sm以上で2列になる", () => {
    render(
      <SupplierInfoSection
        defaultSupplier="仕入先A"
        resolvedLastRestocked="2026-07-01"
        onLastRestockedChange={vi.fn()}
        onMarkDirty={vi.fn()}
      />,
    );

    expectResponsiveTwoColumnGrid("仕入先情報");
  });
});

describe("StockInfoSection clinical state and form surface", () => {
  it("現在庫が最低在庫以下なら非色依存の在庫不足表示を出す", () => {
    render(
      <StockInfoSection
        defaultQuantity={1}
        defaultMinStockLevel={2}
        defaultLocation="棚A"
        resolvedExpiry=""
        onExpiryChange={vi.fn()}
        onMarkDirty={vi.fn()}
      />,
    );

    expect(screen.getByRole("status", { name: "在庫状態" })).toHaveTextContent(
      "在庫不足（残少）— 現在庫数 1、最低在庫数 2",
    );
  });

  it("現在庫0なら在庫切れを表示し、数量変更時に状態を再評価する", async () => {
    const user = userEvent.setup();
    render(
      <StockInfoSection
        defaultQuantity={0}
        defaultMinStockLevel={50}
        defaultLocation="棚A"
        resolvedExpiry=""
        onExpiryChange={vi.fn()}
        onMarkDirty={vi.fn()}
      />,
    );

    expect(screen.getByRole("status", { name: "在庫状態" })).toHaveTextContent("在庫切れ");
    const quantity = screen.getByRole("spinbutton", { name: /現在庫数/ });
    await user.clear(quantity);
    await user.type(quantity, "51");
    expect(screen.queryByRole("status", { name: "在庫状態" })).not.toBeInTheDocument();
  });

  it("cardはhairline、入力はwhite/form border、エラーは入力へ関連付く", () => {
    render(
      <StockInfoSection
        defaultQuantity={1}
        defaultMinStockLevel={2}
        defaultLocation="棚A"
        resolvedExpiry=""
        onExpiryChange={vi.fn()}
        onMarkDirty={vi.fn()}
        minStockLevelError="最低在庫数が不正です"
      />,
    );

    const section = screen.getByRole("heading", { name: "在庫情報" }).parentElement;
    expect(section).toHaveClass(C.bgWhite, C.borderLight);
    const minStock = screen.getByRole("spinbutton", { name: /最低在庫数/ });
    expect(minStock).toHaveClass(C.bgWhite, C.borderMedium);
    expect(minStock).toHaveAttribute("aria-invalid", "true");
    expect(minStock).toHaveAttribute("aria-describedby", "minStockLevel-error");
  });
});

describe("InventoryFormSections inline field errors (BUG-009)", () => {
  it("必須未入力でエラー文言が描画される", () => {
    render(
      <BasicInfoSection
        defaultName=""
        defaultUnit=""
        category="medicine"
        existingCategory={undefined}
        onCategoryChange={vi.fn()}
        onMarkDirty={vi.fn()}
        nameError="品名を入力してください"
        unitError="単位を入力してください"
      />,
    );

    expect(screen.getByText("品名を入力してください")).toBeInTheDocument();
    expect(screen.getByText("単位を入力してください")).toBeInTheDocument();

    const name = screen.getByLabelText(/品名/);
    expect(name).toHaveAttribute("aria-invalid", "true");
    expect(name).toHaveAttribute("aria-describedby", "name-error");

    const unit = screen.getByLabelText(/単位/);
    expect(unit).toHaveAttribute("aria-invalid", "true");
    expect(unit).toHaveAttribute("aria-describedby", "unit-error");
  });

  it("バリデーションエラー後も入力済みの品名が保持される", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <BasicInfoSection
        defaultName=""
        defaultUnit=""
        category="medicine"
        existingCategory={undefined}
        onCategoryChange={vi.fn()}
        onMarkDirty={vi.fn()}
      />,
    );

    const name = screen.getByLabelText(/品名/);
    await user.type(name, "留置針");

    rerender(
      <BasicInfoSection
        defaultName=""
        defaultUnit=""
        category="medicine"
        existingCategory={undefined}
        onCategoryChange={vi.fn()}
        onMarkDirty={vi.fn()}
        nameError="単位を入力してください"
      />,
    );

    expect(screen.getByLabelText(/品名/)).toHaveValue("留置針");
  });

  it("負値でエラー文言が描画される", () => {
    render(
      <StockInfoSection
        defaultQuantity={-1}
        defaultMinStockLevel={-1}
        defaultLocation="棚A"
        resolvedExpiry=""
        onExpiryChange={vi.fn()}
        onMarkDirty={vi.fn()}
        quantityError="現在庫数は0以上の整数で入力してください"
        minStockLevelError="最低在庫数は0以上の整数で入力してください"
      />,
    );

    expect(screen.getByText("現在庫数は0以上の整数で入力してください")).toBeInTheDocument();
    expect(screen.getByText("最低在庫数は0以上の整数で入力してください")).toBeInTheDocument();

    const quantity = screen.getByRole("spinbutton", { name: /現在庫数/ });
    expect(quantity).toHaveAttribute("aria-invalid", "true");
    expect(quantity).toHaveAttribute("aria-describedby", "quantity-error");

    const minStock = screen.getByRole("spinbutton", { name: /最低在庫数/ });
    expect(minStock).toHaveAttribute("aria-invalid", "true");
    expect(minStock).toHaveAttribute("aria-describedby", "minStockLevel-error");
  });
});
