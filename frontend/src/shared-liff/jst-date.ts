// FE4-10: line-reserve/src/lib/jst-date.ts から逐語移動（main の src/lib/jst-date.ts と
// アルゴリズムが重複していた・line-reserve だけ日付系で唯一テストが無かったため統合）。
//
// parse 契約: "YYYY-MM-DD" → Date の解釈は UTC 深夜（この shared-liff 版・line-reserve 系）と
// ローカル正午（src/components/shared/DatePicker/DatePickerModel.ts の parseLocalDate・main 系）
// の 2 契約が併存する。相互交換は日付ズレを起こすため禁止。formatJapaneseDate はこのファイルの
// UTC getter 実装であり、main 側 utils/format/date.ts の formatJapaneseDate（FE4-8・ローカル
// getter・シグネチャも Date 引数で異なる）とは契約が異なるため統合しない。

const JST_OFFSET_MS = 9 * 60 * 60 * 1000;

const WEEK_DAYS = ['日', '月', '火', '水', '木', '金', '土'] as const;

function pad(value: number): string {
  return String(value).padStart(2, '0');
}

function toJSTWallDate(date: Date = new Date()): Date {
  return new Date(date.getTime() + JST_OFFSET_MS);
}

export function parseISODate(dateStr: string): Date {
  const [year, month, day] = dateStr.split('-').map(Number);
  return new Date(Date.UTC(year, month - 1, day));
}

export function addDaysISO(dateStr: string, days: number): string {
  const date = parseISODate(dateStr);
  date.setUTCDate(date.getUTCDate() + days);
  return `${date.getUTCFullYear()}-${pad(date.getUTCMonth() + 1)}-${pad(date.getUTCDate())}`;
}

export function getJSTToday(): Date {
  const wall = toJSTWallDate();
  return new Date(Date.UTC(wall.getUTCFullYear(), wall.getUTCMonth(), wall.getUTCDate()));
}

export function formatJapaneseDate(dateStr: string, padded = false): string {
  if (!dateStr) return '';
  const date = parseISODate(dateStr);
  const year = date.getUTCFullYear();
  const month = date.getUTCMonth() + 1;
  const day = date.getUTCDate();
  const weekDay = WEEK_DAYS[date.getUTCDay()];
  if (padded) {
    return `${year}年${pad(month)}月${pad(day)}日(${weekDay})`;
  }
  return `${year}年${month}月${day}日（${weekDay}）`;
}

export function formatJSTApplicationDate(isoStr: string): string {
  if (!isoStr) return '';
  const date = toJSTWallDate(new Date(isoStr));
  return `${date.getUTCFullYear()}年${date.getUTCMonth() + 1}月${date.getUTCDate()}日 申し込み`;
}

export function formatTimeHHMM(hhmm: string): string {
  if (!hhmm || hhmm.length < 4) return hhmm;
  return `${hhmm.slice(0, 2)}:${hhmm.slice(2, 4)}`;
}
