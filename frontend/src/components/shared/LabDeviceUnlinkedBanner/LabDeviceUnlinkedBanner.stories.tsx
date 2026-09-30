import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { LabDeviceUnlinkedBanner } from "./LabDeviceUnlinkedBanner";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

const authValue: AuthContextValue = {
  user: null,
  currentClinicId: "1",
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => true,
  refreshPermissions: async () => {},
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/LabDeviceUnlinkedBanner",
  component: LabDeviceUnlinkedBanner,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <QueryClientProvider client={queryClient}>
        <AuthContext.Provider value={authValue}>
          <Story />
        </AuthContext.Provider>
      </QueryClientProvider>
    ),
  ],
} satisfies Meta<typeof LabDeviceUnlinkedBanner>;

type Story = StoryObj<typeof LabDeviceUnlinkedBanner>;

/** API 未接続環境では未連携デバイス問合せが resolve しない — 非表示または loading 状態を確認 */
export const Default: Story = {};
