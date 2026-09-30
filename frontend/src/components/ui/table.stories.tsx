import type { Meta, StoryObj } from "@storybook/react-vite";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./table";
import { Badge } from "./badge";

const meta = {
  title: "UI/Table",
  component: Table,
  tags: ["autodocs"],
} satisfies Meta<typeof Table>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

const ROWS = [
  { time: "09:30", pet: "モモ", owner: "山田", status: "診察中" },
  { time: "10:00", pet: "チョコ", owner: "佐藤", status: "待合" },
  { time: "10:30", pet: "リン", owner: "鈴木", status: "会計待ち" },
];

export const Default: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>時刻</TableHead>
          <TableHead>ペット</TableHead>
          <TableHead>飼主</TableHead>
          <TableHead>状態</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {ROWS.map((r) => (
          <TableRow key={r.time}>
            <TableCell>{r.time}</TableCell>
            <TableCell>{r.pet}</TableCell>
            <TableCell>{r.owner}</TableCell>
            <TableCell>
              <Badge variant="secondary">{r.status}</Badge>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  ),
};
