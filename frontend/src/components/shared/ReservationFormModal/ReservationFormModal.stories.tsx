import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryRouter, RouterProvider } from "react-router";
import { ReservationFormModal } from "./ReservationFormModal";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/ReservationFormModal",
  component: ReservationFormModal,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <QueryClientProvider client={queryClient}>
        <RouterProvider
          router={createMemoryRouter([{ path: "*", element: <Story /> }], {
            initialEntries: ["/"],
          })}
        />
      </QueryClientProvider>
    ),
  ],
  args: {
    isOpen: true,
    onClose: () => {},
    onSave: () => {},
    initialData: null,
    canCreate: true,
    canEdit: true,
  },
} satisfies Meta<typeof ReservationFormModal>;

type Story = StoryObj<typeof ReservationFormModal>;

export const NewReservation: Story = {};

export const EditReservation: Story = {
  args: {
    initialData: {
      start: new Date("2026-10-01T10:00:00+09:00"),
      end: new Date("2026-10-01T10:30:00+09:00"),
      notes: "再診",
    },
  },
};
