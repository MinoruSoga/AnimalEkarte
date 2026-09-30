import type { Meta, StoryObj } from "@storybook/react-vite";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { Pet } from "@/types";
import type { PetSelectionResultPage } from "@/hooks/use-pet-selection-page";
import { PetSelectionResultsTable } from "./PetSelectionResultsTable";
import { PetSelectionSearchForm } from "./PetSelectionSearchForm";

const PET: Pet = {
  id: "1",
  clinicId: undefined,
  ownerId: "10",
  ownerNumber: 10,
  ownerName: "山田 花子",
  ownerNameKana: undefined,
  address: undefined,
  phone: "090-0000-0000",
  petNumber: "P-0123",
  name: "ポチ",
  petNameKana: "ぽち",
  species: "犬",
  animalSpeciesId: undefined,
  breed: "柴犬",
  color: "茶",
  bloodType: undefined,
  microchipNumber: undefined,
  gender: "オス",
  status: "生存",
  birthDate: "2020-04-01",
  neuteredDate: undefined,
  weight: "5.2",
  food: "ドライフード",
  environment: "室内",
  acquisitionType: undefined,
  dangerLevel: undefined,
  dangerReason: undefined,
  lastVisit: undefined,
  insuranceId: undefined,
  insuranceName: undefined,
  insuranceDetails: undefined,
  remarks: "",
  nameOrigin: undefined,
  meetingStory: undefined,
  deceasedAt: undefined,
  deceasedReason: undefined,
};

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

const PAGE: PetSelectionResultPage = {
  items: [
    PET,
    {
      ...PET,
      id: "2",
      name: "ミケ",
      petNameKana: "みけ",
      species: "猫",
      breed: "三毛猫",
      dangerLevel: "高",
      dangerReason: "噛みつき歴あり",
    },
  ],
  totalCount: 2,
  currentPage: 1,
  totalPages: 1,
  startIndex: 1,
  endIndex: 2,
  onPageChange: () => {},
};

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/PetSelection",
  component: PetSelectionResultsTable,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <QueryClientProvider client={queryClient}>
        <Story />
      </QueryClientProvider>
    ),
  ],
} satisfies Meta<typeof PetSelectionResultsTable>;

type Story = StoryObj<typeof PetSelectionResultsTable>;

export const Results: Story = {
  args: { pets: PAGE, onSelect: () => {} },
};

export const Loading: Story = {
  args: { pets: { ...PAGE, items: [] }, onSelect: () => {}, isLoading: true },
};

export const SearchForm: Story = {
  render: () => (
    <PetSelectionSearchForm
      searchParams={{ search: "", ownerId: "", species: "" }}
      setSearchParams={() => {}}
      onClear={() => {}}
    />
  ),
};
