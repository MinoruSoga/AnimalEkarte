import type { Meta, StoryObj } from "@storybook/react-vite";

import { Card, CardContent, CardHeader, CardTitle } from "./card";
import { Button } from "./button";

const meta = {
  title: "UI/Card",
  component: Card,
  tags: ["autodocs"],
} satisfies Meta<typeof Card>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Card style={{ maxWidth: 360 }}>
      <CardHeader>
        <CardTitle>本日の予約</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">午前 9 件 / 午後 12 件</p>
      </CardContent>
    </Card>
  ),
};

export const WithAction: Story = {
  render: () => (
    <Card style={{ maxWidth: 360 }}>
      <CardHeader>
        <CardTitle>会計待ち</CardTitle>
        <Button variant="outline" size="sm" data-slot="card-action">
          すべて表示
        </Button>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">3 件</p>
      </CardContent>
    </Card>
  ),
};
