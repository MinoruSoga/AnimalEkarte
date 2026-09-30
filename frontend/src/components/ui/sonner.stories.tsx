import type { Meta, StoryObj } from "@storybook/react-vite";
import { toast } from "sonner";

import { Toaster } from "./sonner";
import { Button } from "./button";

/** アプリルートの Toaster — top-left / offset 72 / closeButton 付き。 */
const meta = {
  title: "UI/Toaster",
  component: Toaster,
  tags: ["autodocs"],
} satisfies Meta<typeof Toaster>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <>
      <Toaster />
      <div style={{ display: "flex", gap: 8 }}>
        <Button variant="outline" onClick={() => toast("保存しました")}>
          success toast
        </Button>
        <Button variant="outline" onClick={() => toast.error("保存に失敗しました")}>
          error toast
        </Button>
      </div>
    </>
  ),
};
