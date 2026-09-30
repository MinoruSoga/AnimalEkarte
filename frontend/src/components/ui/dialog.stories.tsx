import type { Meta, StoryObj } from "@storybook/react-vite";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./dialog";
import { Button } from "./button";

/**
 * EMR-64 の focus restore 実装を含む。DialogContent は open 毎に再マウントされる。
 */
const meta = {
  title: "UI/Dialog",
  component: Dialog,
  tags: ["autodocs"],
} satisfies Meta<typeof Dialog>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline">新規登録</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>飼主の新規登録</DialogTitle>
          <DialogDescription>必須項目を入力して登録してください。</DialogDescription>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">フォーム本体はここに配置します。</p>
        <DialogFooter>
          <Button variant="outline">キャンセル</Button>
          <Button>登録</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

/** 初期表示で開いた状態（a11y addon がオーバーレイを検査できるように）。 */
export const Open: Story = {
  render: () => (
    <Dialog defaultOpen>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>削除の確認</DialogTitle>
          <DialogDescription>この操作は取り消せません。</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline">キャンセル</Button>
          <Button variant="destructive">削除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};
