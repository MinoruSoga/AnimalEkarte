import type { Meta, StoryObj } from "@storybook/react-vite";
import { PawPrint } from "lucide-react";
import { FormHeader } from "./FormHeader";
import { SubmitButton } from "./SubmitButton";
import { PrimaryButton } from "./PrimaryButton";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/Form",
  component: FormHeader,
  tags: ["autodocs"],
  args: { title: "診療内容の登録" },
} satisfies Meta<typeof FormHeader>;

type Story = StoryObj<typeof FormHeader>;

export const Header: Story = {};

export const HeaderFull: Story = {
  args: {
    title: "ワクチン接種記録",
    description: "接種日とワクチン種別を入力します",
    icon: <PawPrint className="size-5" />,
    onBack: () => {},
    action: <PrimaryButton>保存する</PrimaryButton>,
  },
};

export const SubmitButtons: Story = {
  render: () => (
    <div className="flex flex-col items-start gap-4">
      <PrimaryButton>PrimaryButton（primary）</PrimaryButton>
      <PrimaryButton colorVariant="brand">PrimaryButton（brand）</PrimaryButton>
      <PrimaryButton colorVariant="default">PrimaryButton（default alias）</PrimaryButton>
      <PrimaryButton disabled>disabled</PrimaryButton>
      <SubmitButton>SubmitButton（form 連動）</SubmitButton>
      <SubmitButton colorVariant="destructive">SubmitButton（destructive）</SubmitButton>
    </div>
  ),
};
