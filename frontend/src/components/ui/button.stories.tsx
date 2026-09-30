import type { Meta, StoryObj } from "@storybook/react-vite";
import { Plus, Trash2 } from "lucide-react";

import { Button } from "./button";

/**
 * design-system.md §7.2 + design-states.md §1–3 の仕様を全状態で確認するカタログ。
 * focus-visible ring はキーボード起因のみ発火するため、Storybook 上の検証は
 * `button.test.tsx`（EMR-207 回帰）と axe（addon-a11y）で担保する。
 */
const meta = {
  title: "UI/Button",
  component: Button,
  tags: ["autodocs"],
  argTypes: {
    variant: {
      control: "select",
      options: [
        "default",
        "primary",
        "destructive",
        "outline",
        "secondary",
        "ghost",
        "link",
        "ghost-danger",
      ],
    },
    size: { control: "select", options: ["default", "sm", "lg", "icon"] },
    loading: { control: "boolean" },
    disabled: { control: "boolean" },
  },
  args: { children: "保存" },
} satisfies Meta<typeof Button>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

/** Primary CTA — pill + #038B94。design-system.md §7.2 button-primary。 */
export const Default: Story = {};

export const AllVariants: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 12, flexWrap: "wrap", alignItems: "center" }}>
      <Button variant="default">primary（既定）</Button>
      <Button variant="secondary">secondary</Button>
      <Button variant="outline">outline</Button>
      <Button variant="ghost">ghost</Button>
      <Button variant="link">link</Button>
      <Button variant="destructive">destructive</Button>
      <Button variant="ghost-danger">ghost-danger</Button>
    </div>
  ),
};

export const Sizes: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
      <Button size="sm">sm</Button>
      <Button size="default">default</Button>
      <Button size="lg">lg</Button>
      <Button size="icon" aria-label="追加">
        <Plus />
      </Button>
    </div>
  ),
};

/** design-states.md §3 — loading: spinner + aria-busy + disabled、ラベル維持。 */
export const Loading: Story = {
  args: { loading: true, children: "保存中" },
};

export const Disabled: Story = {
  args: { disabled: true, children: "保存" },
};

/** dense 行アクション例外（design-system.md §7.2: h-9 まで縮小可・min-w-11 維持）。 */
export const DenseRowAction: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 8 }}>
      <Button variant="ghost" className="h-9 px-3 text-sm">
        <Trash2 /> 削除
      </Button>
      <Button variant="outline" className="h-9 px-3 text-sm">
        編集
      </Button>
    </div>
  ),
};
