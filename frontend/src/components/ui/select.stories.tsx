import type { Meta, StoryObj } from "@storybook/react-vite";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";
import { Label } from "./label";

const meta = {
  title: "UI/Select",
  component: Select,
  tags: ["autodocs"],
} satisfies Meta<typeof Select>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

const SPECIES = [
  { value: "dog", label: "犬" },
  { value: "cat", label: "猫" },
  { value: "rabbit", label: "うさぎ" },
  { value: "other", label: "その他" },
];

export const Default: Story = {
  render: () => (
    <div style={{ maxWidth: 320 }}>
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="種別を選択" />
        </SelectTrigger>
        <SelectContent>
          {SPECIES.map((s) => (
            <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  ),
};

export const WithLabel: Story = {
  render: () => (
    <div style={{ display: "grid", gap: 8, maxWidth: 320 }}>
      <Label htmlFor="species">動物種別</Label>
      <Select defaultValue="cat">
        <SelectTrigger id="species">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {SPECIES.map((s) => (
            <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  ),
};

export const Invalid: Story = {
  render: () => (
    <div style={{ maxWidth: 320 }}>
      <Select>
        <SelectTrigger aria-invalid="true">
          <SelectValue placeholder="未選択（必須）" />
        </SelectTrigger>
        <SelectContent>
          {SPECIES.map((s) => (
            <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  ),
};

export const Disabled: Story = {
  render: () => (
    <div style={{ maxWidth: 320 }}>
      <Select disabled defaultValue="dog">
        <SelectTrigger>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {SPECIES.map((s) => (
            <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  ),
};
