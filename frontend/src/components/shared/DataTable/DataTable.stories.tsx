import type { Meta, StoryObj } from "@storybook/react-vite";
import { TableCell, TableRow } from "@/components/ui/table";
import { DataTable } from "./DataTable";

interface Pet {
  id: number;
  name: string;
  species: string;
  owner: string;
}

const DATA: Pet[] = [
  { id: 1, name: "ポチ", species: "犬", owner: "山田" },
  { id: 2, name: "ミケ", species: "猫", owner: "鈴木" },
  { id: 3, name: "ピーちゃん", species: "鳥", owner: "佐藤" },
];

const COLUMNS = [{ header: "名前" }, { header: "種別" }, { header: "飼い主" }];

function renderPetRow(pet: Pet) {
  return (
    <TableRow>
      <TableCell>{pet.name}</TableCell>
      <TableCell>{pet.species}</TableCell>
      <TableCell>{pet.owner}</TableCell>
    </TableRow>
  );
}

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/DataTable",
  component: DataTable<Pet>,
  tags: ["autodocs"],
  args: {
    columns: COLUMNS,
    data: DATA,
    renderRow: renderPetRow,
  },
} satisfies Meta<typeof DataTable<Pet>>;

type Story = StoryObj<typeof DataTable<Pet>>;

export const Default: Story = {};

export const Empty: Story = {
  args: { data: [], emptyMessage: "対象の記録がありません" },
};
