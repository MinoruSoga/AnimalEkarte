import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { PetDeceasedRecordButton } from "./PetDeceasedRecordButton";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PetDeceasedRecordButton",
  component: PetDeceasedRecordButton,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <QueryClientProvider client={queryClient}>
        <Story />
      </QueryClientProvider>
    ),
  ],
  args: {
    petId: "1",
    petName: "ポチ",
    petBreed: "柴犬",
    petGender: "オス",
    birthDate: "2020-04-01",
    petStatus: "alive",
    canEdit: true,
  },
} satisfies Meta<typeof PetDeceasedRecordButton>;

type Story = StoryObj<typeof PetDeceasedRecordButton>;

export const Alive: Story = {};

export const AlreadyDeceased: Story = {
  args: { petStatus: "deceased", deceasedAt: "2026-09-01T10:00:00Z" },
};

export const CannotEdit: Story = {
  args: { canEdit: false },
};
