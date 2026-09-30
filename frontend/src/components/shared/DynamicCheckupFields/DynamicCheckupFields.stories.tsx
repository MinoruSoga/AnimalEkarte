import type { Meta, StoryObj } from "@storybook/react-vite";
import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";
import { DynamicCheckupFields } from "./DynamicCheckupFields";

const FIELDS: CheckupTypeFieldRow[] = [
  {
    id: 1,
    checkupTypeId: 1,
    name: "体重",
    fieldType: "number",
    unit: "kg",
    options: [],
    isProvisional: false,
    sortOrder: 1,
  },
  {
    id: 2,
    checkupTypeId: 1,
    name: "食欲",
    fieldType: "single_select",
    unit: "",
    options: [
      { value: "good", label: "良好" },
      { value: "poor", label: "不良" },
    ],
    isProvisional: false,
    sortOrder: 2,
  },
  {
    id: 3,
    checkupTypeId: 1,
    name: "備考",
    fieldType: "text",
    unit: "",
    options: [],
    isProvisional: true,
    sortOrder: 3,
  },
];

// eslint-disable-next-line no-restricted-syntax -- Storybook CSF requires default export
export default {
  title: "Shared/DynamicCheckupFields",
  component: DynamicCheckupFields,
  tags: ["autodocs"],
  args: {
    fields: FIELDS,
    values: { 1: { number: "5.2" }, 2: { list: undefined, text: undefined, bool: undefined } },
    onChange: () => {},
  },
} satisfies Meta<typeof DynamicCheckupFields>;

type Story = StoryObj<typeof DynamicCheckupFields>;

export const Default: Story = {};

export const Empty: Story = {
  args: { fields: [], values: {} },
};
