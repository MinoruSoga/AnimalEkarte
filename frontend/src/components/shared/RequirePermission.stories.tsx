import type { Meta, StoryObj } from "@storybook/react-vite";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { RequirePermission } from "./RequirePermission";

const authValue = (allowed: boolean): AuthContextValue => ({
  user: null,
  currentClinicId: "1",
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => allowed,
  refreshPermissions: async () => {},
});

const withAuth = (allowed: boolean) =>
  function Decorator(Story: React.ComponentType) {
    return (
      <AuthContext.Provider value={authValue(allowed)}>
        <Story />
      </AuthContext.Provider>
    );
  };

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/RequirePermission",
  component: RequirePermission,
  tags: ["autodocs"],
  args: {
    resource: "medical-records",
    action: "edit",
    children: <div className="rounded-md border p-4">権限ありユーザーにだけ見える操作</div>,
  },
} satisfies Meta<typeof RequirePermission>;

type Story = StoryObj<typeof RequirePermission>;

export const Allowed: Story = {
  decorators: [withAuth(true)],
};

export const Denied: Story = {
  decorators: [withAuth(false)],
};

export const DeniedCustomFallback: Story = {
  decorators: [withAuth(false)],
  args: { fallback: <p className="text-muted-foreground">この操作の権限がありません</p> },
};
