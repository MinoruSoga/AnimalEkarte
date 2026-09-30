import type { Meta, StoryObj } from "@storybook/react-vite";
import { PrintPortal } from "./PrintPortal";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PrintPortal",
  component: PrintPortal,
  tags: ["autodocs"],
  args: {
    testId: "print-demo",
    orientation: "portrait",
    children: (
      <div>
        <h1>印刷用レイアウト（A4 portrait）</h1>
        <p>この内容は document.body 直下のポータルに描画されます。</p>
      </div>
    ),
  },
} satisfies Meta<typeof PrintPortal>;

type Story = StoryObj<typeof PrintPortal>;

export const Portrait: Story = {};

export const Landscape: Story = {
  args: { testId: "print-demo-landscape", orientation: "landscape" },
};
