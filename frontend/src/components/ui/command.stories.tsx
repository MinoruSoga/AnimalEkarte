import type { Meta, StoryObj } from "@storybook/react-vite";

import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "./command";

/** cmdk ラッパ — 実消費は searchable-select.tsx の検索ポップオーバー。 */
const meta = {
  title: "UI/Command",
  component: Command,
  tags: ["autodocs"],
} satisfies Meta<typeof Command>;

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF は `export default meta` が必須形式
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Command style={{ maxWidth: 320, border: "1px solid var(--border)", borderRadius: 6 }}>
      <CommandInput placeholder="マスタを検索" />
      <CommandList>
        <CommandEmpty>該当なし</CommandEmpty>
        <CommandGroup heading="診療">
          <CommandItem>初診料</CommandItem>
          <CommandItem>再診料</CommandItem>
          <CommandItem>ワクチン接種</CommandItem>
        </CommandGroup>
        <CommandGroup heading="処置">
          <CommandItem>爪切り</CommandItem>
          <CommandItem>耳掃除</CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  ),
};
