import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryRouter, RouterProvider } from "react-router";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { PartnerRecordLink } from "./PartnerRecordLink";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

const authValue: AuthContextValue = {
  user: {
    id: "1",
    email: "staff@example.com",
    displayName: "佐藤 一郎",
    isSystemAdmin: false,
    mainClinicId: "1",
    clinic: null,
    clinics: [{ clinicId: "1", clinicName: "本院", isMain: true }],
    permissions: {},
  },
  currentClinicId: "1",
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => true,
  refreshPermissions: async () => {},
};

const withProviders = (Story: React.ComponentType) => {
  const router = createMemoryRouter([{ path: "*", element: <Story /> }], {
    initialEntries: ["/"],
  });
  return (
    <QueryClientProvider client={queryClient}>
      <AuthContext.Provider value={authValue}>
        <RouterProvider router={router} />
      </AuthContext.Provider>
    </QueryClientProvider>
  );
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PartnerRecordLink",
  component: PartnerRecordLink,
  tags: ["autodocs"],
  decorators: [withProviders],
  args: {
    kind: "medical-record",
    petId: "1",
    visitDate: "2026-09-30",
  },
} satisfies Meta<typeof PartnerRecordLink>;

type Story = StoryObj<typeof PartnerRecordLink>;

/** 相方 record 解決中・非表示状態（API 未接続でも描画自体は確認できる） */
export const Default: Story = {};

/** 死亡ペットでは遷移導線を出さない（FE12 表示側防壁） */
export const DeceasedPet: Story = {
  args: { isPetDeceased: true },
};
