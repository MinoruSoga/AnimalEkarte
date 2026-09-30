import type { Meta, StoryObj } from "@storybook/react-vite";
import { ImageWithFallback } from "./ImageWithFallback";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/Feedback",
  component: ImageWithFallback,
  tags: ["autodocs"],
  args: {
    alt: "ペットの写真",
    className: "size-24 rounded-md object-cover",
    // 1x1 px の PNG（画像読み込み成功ケース）
    src: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
  },
} satisfies Meta<typeof ImageWithFallback>;

type Story = StoryObj<typeof ImageWithFallback>;

export const Loaded: Story = {};

/** 読み込み失敗時のフォールバック表示（無効な data URI でネットワーク 404 を発生させない） */
export const ErrorFallback: Story = {
  args: { src: "data:image/png;base64,INVALID" },
};
