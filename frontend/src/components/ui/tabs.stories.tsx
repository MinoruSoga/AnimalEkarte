import type { Meta, StoryObj } from "@storybook/react-vite";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "./tabs";

const meta = {
  title: "UI/Tabs",
  component: Tabs,
  tags: ["autodocs"],
} satisfies Meta<typeof Tabs>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Tabs defaultValue="record" style={{ maxWidth: 480 }}>
      <TabsList>
        <TabsTrigger value="record">カルテ</TabsTrigger>
        <TabsTrigger value="vaccine">ワクチン</TabsTrigger>
        <TabsTrigger value="lab">検査</TabsTrigger>
      </TabsList>
      <TabsContent value="record">
        <p className="text-sm text-muted-foreground">カルテ一覧のコンテンツ</p>
      </TabsContent>
      <TabsContent value="vaccine">
        <p className="text-sm text-muted-foreground">ワクチン履歴のコンテンツ</p>
      </TabsContent>
      <TabsContent value="lab">
        <p className="text-sm text-muted-foreground">検査結果のコンテンツ</p>
      </TabsContent>
    </Tabs>
  ),
};

export const WithDisabled: Story = {
  render: () => (
    <Tabs defaultValue="record" style={{ maxWidth: 480 }}>
      <TabsList>
        <TabsTrigger value="record">カルテ</TabsTrigger>
        <TabsTrigger value="archive" disabled>
          過去カルテ（権限なし）
        </TabsTrigger>
      </TabsList>
      <TabsContent value="record">
        <p className="text-sm text-muted-foreground">
          権限のないタブは disabled — 視覚だけでなく操作も不可。
        </p>
      </TabsContent>
    </Tabs>
  ),
};
