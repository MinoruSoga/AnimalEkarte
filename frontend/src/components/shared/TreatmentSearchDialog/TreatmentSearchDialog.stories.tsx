import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TreatmentSearchDialog } from "./TreatmentSearchDialog";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/TreatmentSearchDialog",
  component: TreatmentSearchDialog,
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
} satisfies Meta<typeof TreatmentSearchDialog>;

type Story = StoryObj<typeof TreatmentSearchDialog>;

/** マスタ取得中のローディング状態（API 未接続環境ではこの状態で確認できる） */
export const Loading: Story = {};
