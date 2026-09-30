import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext } from "@/hooks/auth-context";
import { queryKeys } from "@/lib/query-keys";
import type { AuthContextValue } from "@/types/auth";
import type { LabDeviceJobCard } from "@/hooks/use-lab-device-unlinked";
import { LabDeviceUnlinkedBanner } from "./LabDeviceUnlinkedBanner";

const UNLINKED_JOB: LabDeviceJobCard = {
  jobId: "job-001",
  sourceType: "fujifilm",
  deviceHint: "FUJI DRI-CHEM",
  status: "unlinked",
  specimenIdRaw: "SP-20260930-001",
  itemCount: 12,
  unmappedItemCount: 0,
  clockSkew: false,
  items: [],
  measuredAt: "2026-09-30T09:00:00+09:00",
  receivedAt: "2026-09-30T09:05:00+09:00",
};

const seededClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});
seededClient.setQueryData(queryKeys.labDevice.unlinked(), [UNLINKED_JOB]);

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
  args: { petId: "1" },
  decorators: [
    (Story) => (
      <QueryClientProvider client={seededClient}>
        <AuthContext.Provider value={authValue}>
          <Story />
        </AuthContext.Provider>
      </QueryClientProvider>
    ),
  ],
} satisfies Meta<typeof LabDeviceUnlinkedBanner>;

type Story = StoryObj<typeof LabDeviceUnlinkedBanner>;

/** 未紐付け受信あり → バナーが表示される（クエリキャッシュをシード済み） */
export const WithUnlinkedJobs: Story = {};
