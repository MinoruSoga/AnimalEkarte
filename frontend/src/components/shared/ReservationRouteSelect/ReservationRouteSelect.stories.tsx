import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReservationRouteSelect } from "./ReservationRouteSelect";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/ReservationRouteSelect",
  component: ReservationRouteSelect,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <QueryClientProvider client={queryClient}>
        <Story />
      </QueryClientProvider>
    ),
  ],
  args: {
    reservationId: "1",
    value: "reception",
  },
} satisfies Meta<typeof ReservationRouteSelect>;

type Story = StoryObj<typeof ReservationRouteSelect>;

export const Reception: Story = {};

export const Line: Story = {
  args: { value: "line" },
};

export const Empty: Story = {
  args: { value: null },
};

export const Disabled: Story = {
  args: { disabled: true },
};
