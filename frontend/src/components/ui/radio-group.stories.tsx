import type { Meta, StoryObj } from "@storybook/react-vite";

import { RadioGroup, RadioGroupItem } from "./radio-group";
import { Label } from "./label";

const meta = {
  title: "UI/RadioGroup",
  component: RadioGroup,
  tags: ["autodocs"],
} satisfies Meta<typeof RadioGroup>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <RadioGroup defaultValue="male">
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <RadioGroupItem value="male" id="rg-male" />
        <Label htmlFor="rg-male" className="mb-0">
          オス
        </Label>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <RadioGroupItem value="female" id="rg-female" />
        <Label htmlFor="rg-female" className="mb-0">
          メス
        </Label>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <RadioGroupItem value="unknown" id="rg-unknown" disabled />
        <Label htmlFor="rg-unknown" className="mb-0">
          不明（disabled）
        </Label>
      </div>
    </RadioGroup>
  ),
};
