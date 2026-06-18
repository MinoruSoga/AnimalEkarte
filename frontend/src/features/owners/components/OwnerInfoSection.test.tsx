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
  render(<OwnerInfoHarness onChange={onChange} />);
  return { onChange };
}

describe("OwnerInfoSection", () => {
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
