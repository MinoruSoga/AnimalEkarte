import type { Meta, StoryObj } from "@storybook/react-vite";

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "./sheet";
import { Button } from "./button";

const meta = {
  title: "UI/Sheet",
  component: Sheet,
  tags: ["autodocs"],
} satisfies Meta<typeof Sheet>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline">詳細を開く</Button>
      </SheetTrigger>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>ペット詳細</SheetTitle>
          <SheetDescription>サイドパネルからのクイック参照。</SheetDescription>
        </SheetHeader>
        <p className="text-sm text-muted-foreground">詳細コンテンツ</p>
      </SheetContent>
    </Sheet>
  ),
};

export const OpenLeft: Story = {
  render: () => (
    <Sheet defaultOpen>
      <SheetContent side="left">
        <SheetHeader>
          <SheetTitle>ナビゲーション</SheetTitle>
          <SheetDescription>左から出る Sheet。</SheetDescription>
        </SheetHeader>
      </SheetContent>
    </Sheet>
  ),
};
