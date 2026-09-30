import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { OwnerSearchModal } from "./OwnerSearchModal";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/OwnerSearchModal",
  component: OwnerSearchModal,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <QueryClientProvider client={queryClient}>
        <Story />
      </QueryClientProvider>
    ),
  ],
  args: {
    open: true,
    onOpenChange: () => {},
    onSelect: () => {},
  },
} satisfies Meta<typeof OwnerSearchModal>;

type Story = StoryObj<typeof OwnerSearchModal>;

export const Default: Story = {};

export const WithCurrentOwner: Story = {
  args: { currentOwnerName: "山田 花子" },
};
