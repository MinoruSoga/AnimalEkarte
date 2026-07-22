import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { OwnerInfoSection } from "./OwnerInfoSection";
import type { OwnerData } from "../types";

const baseOwner: OwnerData = {
  ownerId: "42",
  postalCode: "",
  company: "",
  membershipType: "会員",
  ownerName: "山田太郎",
  address1: "",
  ownerNameKana: "やまだたろう",
  address2: "",
  homeAddress1: "",
  homeAddress2: "",
  isDangerous: false,
  birthDate: "",
  email: "",
  phone: "090-1111-2222",
  companyPhone: "",
  remarks: "",
  dmPreference: undefined,
};

function OwnerInfoHarness({ onChange }: { onChange: ReturnType<typeof vi.fn> }) {
  const [ownerData, setOwnerData] = useState<OwnerData>(baseOwner);
  const handleChange = (field: string, value: string | boolean | number | null | undefined) => {
    onChange(field, value);
    setOwnerData((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <OwnerInfoSection
      ownerData={ownerData}
      fieldErrors={{}}
      isEdit
      canEditDiscount
      onChange={handleChange}
      onClearError={() => {}}
      onMembershipChange={() => {}}
      onPostalCodeLookup={() => {}}
    />
  );
}

function renderOwnerInfo() {
  const onChange = vi.fn();
  const { container } = render(<OwnerInfoHarness onChange={onChange} />);
  return { container, onChange };
}

describe("OwnerInfoSection", () => {
  it("飼主生年月日のfocusable inputを44px以上に保つ", () => {
    renderOwnerInfo();

    expect(screen.getByLabelText("飼主生年月日")).toHaveClass("min-h-11");
  });

  it("mobileでは単一列にし、sm以上で2列、lg以上で既存の4列へ戻す", () => {
    const { container } = renderOwnerInfo();
    const dangerousField = screen.getByText("危険人物").parentElement;

    expect(container.firstElementChild).toHaveClass(
      "w-full",
      "grid-cols-1",
      "sm:grid-cols-2",
      "lg:grid-cols-4",
    );
    expect(dangerousField).toHaveClass(
      "col-span-1",
      "sm:col-span-2",
      "lg:col-span-1",
    );
    expect(dangerousField).not.toHaveClass("col-span-2");
  });

  it("DM区分を未設定/必要/不要として編集できる", async () => {
    const user = userEvent.setup();
    const { onChange } = renderOwnerInfo();

    expect(screen.getByText("未設定")).toBeInTheDocument();

    await user.click(screen.getByRole("combobox", { name: "DM" }));
    await user.click(screen.getByRole("option", { name: "必要" }));
    expect(onChange).toHaveBeenCalledWith("dmPreference", true);

    await user.click(screen.getByRole("combobox", { name: "DM" }));
    await user.click(screen.getByRole("option", { name: "不要" }));
    expect(onChange).toHaveBeenCalledWith("dmPreference", false);

    await user.click(screen.getByRole("combobox", { name: "DM" }));
    await user.click(screen.getByRole("option", { name: "未設定" }));
    expect(onChange).toHaveBeenCalledWith("dmPreference", null);
  });
});
