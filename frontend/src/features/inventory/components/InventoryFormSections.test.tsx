import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

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
