import type { VaccinationRecord } from "@/types";

export const MOCK_VACCINATION_RECORDS: VaccinationRecord[] = [
  {
    id: "1",
    ownerName: "林 文明",
    petName: "Iris",
    vaccineName: "狂犬病ワクチン",
    date: "2025/10/10",
    nextDate: "2026/10/10",
  },
  {
    id: "2",
    ownerName: "田中　花子",
    petName: "ミケ",
    vaccineName: "3種混合ワクチン",
    date: "2025/09/15",
    nextDate: "2026/09/15",
  },
  {
    id: "3",
    ownerName: "山本 健太",
    petName: "ポチ",
    vaccineName: "フィラリア予防",
    date: "2025/08/20",
    nextDate: "2025/09/20",
  },
  {
    id: "4",
    ownerName: "鈴木 一郎",
    petName: "タロウ",
    vaccineName: "ノミダニ予防",
    date: "2025/08/22",
    nextDate: "2025/09/22",
  },
  {
    id: "5",
    ownerName: "佐々木 亮",
    petName: "チビ",
    vaccineName: "3種混合ワクチン",
    date: "2026/01/06",
    nextDate: "2027/01/06",
  },
  // Leo: Last year's record, making him due today
  {
    id: "6",
    ownerName: "佐藤 花子",
    petName: "レオ",
    vaccineName: "猫ウイルス性鼻気管炎",
    date: "2025/01/06",
    nextDate: "2026/01/06",
  },
  // Mike: Today's new record
  {
    id: "7",
    ownerName: "田中　花子",
    petName: "ミケ",
    vaccineName: "3種混合ワクチン",
    date: "2026/01/06",
    nextDate: "2027/01/06",
  },
];
