import type { Meta, StoryObj } from "@storybook/react-vite";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { PermissionBadges } from "./PermissionBadges";

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
  title: "Shared/PermissionBadges",
  component: PermissionBadges,
  tags: ["autodocs"],
  args: { resource: "medical-records" },
} satisfies Meta<typeof PermissionBadges>;

type Story = StoryObj<typeof PermissionBadges>;

export const AllAllowed: Story = {
  decorators: [withAuth(true)],
};

export const NoneAllowed: Story = {
  decorators: [withAuth(false)],
};
