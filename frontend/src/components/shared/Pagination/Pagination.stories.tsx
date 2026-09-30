import type { Meta, StoryObj } from "@storybook/react-vite";
import { Pagination } from "./Pagination";

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/Pagination",
  component: Pagination,
  tags: ["autodocs"],
  args: {
    currentPage: 2,
    totalPages: 10,
    totalCount: 196,
    startIndex: 11,
    endIndex: 20,
    onPageChange: () => {},
    onPrev: () => {},
    onNext: () => {},
  },
} satisfies Meta<typeof Pagination>;

type Story = StoryObj<typeof Pagination>;

export const MiddlePage: Story = {};

export const FirstPage: Story = {
  args: { currentPage: 1, startIndex: 1, endIndex: 10 },
};

export const LastPage: Story = {
  args: { currentPage: 10, startIndex: 191, endIndex: 196 },
};

export const SinglePage: Story = {
  args: { currentPage: 1, totalPages: 1, totalCount: 7, startIndex: 1, endIndex: 7 },
};
