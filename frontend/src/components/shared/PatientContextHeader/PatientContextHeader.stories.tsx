import type { Meta, StoryObj } from "@storybook/react-vite";
import { createMemoryRouter, RouterProvider } from "react-router";
import { PatientContextHeader } from "./PatientContextHeader";

const withRouter = (Story: React.ComponentType) => {
  const router = createMemoryRouter([{ path: "*", element: <Story /> }], {
    initialEntries: ["/"],
  });
  return <RouterProvider router={router} />;
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PatientContextHeader",
  component: PatientContextHeader,
  tags: ["autodocs"],
  decorators: [withRouter],
  args: {
    ownerName: "山田 花子",
    petName: "ポチ",
    petNumber: "P-0123",
    weight: "5.2kg",
    status: "alive",
    birthDate: "2020-04-01",
    species: "犬",
    breed: "柴犬",
    gender: "オス",
  },
} satisfies Meta<typeof PatientContextHeader>;

type Story = StoryObj<typeof PatientContextHeader>;

export const Default: Story = {};

export const Deceased: Story = {
  args: { status: "deceased" },
};

export const WithInsurance: Story = {
  args: { insuranceName: "アニコム" },
};
