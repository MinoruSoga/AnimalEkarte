import type { Meta, StoryObj } from "@storybook/react-vite";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FormDialog } from "./FormDialog";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/FormDialog",
  component: FormDialog,
  tags: ["autodocs"],
  args: {
    open: true,
    onClose: () => {},
    onSave: () => {},
    title: "項目を編集",
    description: "フォーム内容を編集して保存します",
    children: (
      <div className="grid gap-2">
        <Label htmlFor="fd-name">名前</Label>
        <Input id="fd-name" defaultValue="柴犬" />
      </div>
    ),
  },
} satisfies Meta<typeof FormDialog>;

type Story = StoryObj<typeof FormDialog>;

export const Default: Story = {};

export const Pending: Story = {
  args: { isPending: true },
};

export const CustomLabels: Story = {
  args: { saveLabel: "登録する", cancelLabel: "やめる" },
};
