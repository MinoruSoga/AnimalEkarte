import type { Meta, StoryObj } from "@storybook/react-vite";
import { createMemoryRouter, RouterProvider } from "react-router";
import { MasterLink } from "./MasterLink";

const withRouter = (Story: React.ComponentType) => {
  const router = createMemoryRouter([{ path: "*", element: <Story /> }], {
    initialEntries: ["/"],
  });
  return <RouterProvider router={router} />;
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/MasterLink",
  component: MasterLink,
  tags: ["autodocs"],
  decorators: [withRouter],
  args: { category: "reservationType" },
} satisfies Meta<typeof MasterLink>;

type Story = StoryObj<typeof MasterLink>;

export const Default: Story = {};

export const CustomLabel: Story = {
  args: { category: "medicine", label: "薬品マスタを開く" },
};
