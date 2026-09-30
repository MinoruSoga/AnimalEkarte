import type { Meta, StoryObj } from "@storybook/react-vite";
import { createMemoryRouter, RouterProvider } from "react-router";
import { NavigationBlocker } from "./NavigationBlocker";

// useBlocker は data router 配下でのみ動作するため memory router で包む
const withRouter = (Story: React.ComponentType) => {
  const router = createMemoryRouter([{ path: "*", element: <Story /> }], {
    initialEntries: ["/"],
  });
  return <RouterProvider router={router} />;
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/NavigationBlocker",
  component: NavigationBlocker,
  tags: ["autodocs"],
  decorators: [withRouter],
  args: { when: false },
} satisfies Meta<typeof NavigationBlocker>;

type Story = StoryObj<typeof NavigationBlocker>;

/** when=false: ブロッカー非活性（useBlocker 非マウント、何も描画しない） */
export const Inactive: Story = {};

/**
 * when=true: ナビゲーションを試みると ConfirmDialog が開く。
 * canvas 内でブラウザ戻る/別リンク遷移を試すと確認できる。
 */
export const Active: Story = {
  args: {
    when: true,
    title: "編集中の内容を破棄しますか？",
    description: "保存していない変更があります。",
  },
};
