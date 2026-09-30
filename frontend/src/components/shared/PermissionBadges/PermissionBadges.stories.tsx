import type { Meta, StoryObj } from "@storybook/react-vite";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { PermissionBadges } from "./PermissionBadges";

const authValue = (actions: string[] | "all"): AuthContextValue => ({
  user: null,
  currentClinicId: "1",
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: (_resource, action) => actions === "all" || actions.includes(action),
  refreshPermissions: async () => {},
});

const withAuth = (actions: string[] | "all") =>
  function Decorator(Story: React.ComponentType) {
    return (
      <AuthContext.Provider value={authValue(actions)}>
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

/** 「閲覧のみ」専用バッジ */
export const ViewOnly: Story = {
  decorators: [withAuth(["view"])],
};

/** 一部権限のみ → 持っている権限のバッジを表示 */
export const ViewAndEdit: Story = {
  decorators: [withAuth(["view", "edit"])],
};

/** 全権限あり → コンポーネントの契約上 何も描画しない（空白が正解） */
export const AllAllowed: Story = {
  decorators: [withAuth("all")],
};

/** 権限ゼロ → AccessDenied 前提のため通常到達しないが、こちらも非表示 */
export const NoneAllowed: Story = {
  decorators: [withAuth([])],
};
