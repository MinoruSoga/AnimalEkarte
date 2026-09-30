import type { Meta, StoryObj } from "@storybook/react-vite";
import { createMemoryRouter, RouterProvider } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { Layout } from "./Layout";

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

const authValue: AuthContextValue = {
  user: {
    id: "1",
    email: "sato@example.com",
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
  const router = createMemoryRouter(
    [
      {
        path: "*",
        element: <Story />,
        children: [{ path: "*", element: <div className="p-8">ページコンテンツ</div> }],
      },
    ],
    { initialEntries: ["/"] },
  );
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
  title: "Shared/Layout",
  component: Layout,
  tags: ["autodocs"],
  decorators: [withProviders],
} satisfies Meta<typeof Layout>;

type Story = StoryObj<typeof Layout>;

/** 認証済み + 全権限のシェル表示 */
export const Authenticated: Story = {};
