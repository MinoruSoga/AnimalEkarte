import { describe, it, expect } from 'vitest';
import type React from 'react';
import { BADGE } from './design-tokens';
import {
  getMedicalRecordStatusColor,
  getHospitalizationStatusColor,
  getHospitalizationTypeColor,
  getReceptionColumnColor,
  getReservationTypeColor,
  getReservationTypeName,
  getExaminationStatusColor,
  getAccountingStatusColor,
  getTrimmingStatusColor,
  getPetStatusColor,
  getMasterStatusColor,
  getInventoryStatusColor,
  getInventoryStatusLabel,
  getCalendarViewLabel,
  getReservationStatusLabel,
  getSortOrderLabel,
  getEstimateStatusColor,
} from './status-helpers';

// Notion カラーパレット hex 値（status-helpers.ts の N 定数と一致させること）
const HEX = {
  blue:   'D3E5EF',
  green:  'DDEDEA',
  purple: 'EEE0F7',
  orange: 'FAEBDD',
  yellow: 'FDECC8',
  red:    'FFE2DD',
  gray:   'EBECED',
} as const;

describe('getMedicalRecordStatusColor', () => {
  it('returns blue for 作成中', () => {
    expect(getMedicalRecordStatusColor('作成中')).toContain(HEX.blue);
  });

  it('returns gray for 確定済', () => {
    expect(getMedicalRecordStatusColor('確定済')).toContain(HEX.gray);
  });

  it('returns empty string for unknown status', () => {
    expect(getMedicalRecordStatusColor('unknown' as '作成中')).toBe('');
  });
});

describe('getHospitalizationStatusColor', () => {
  it('returns blue for 入院中', () => {
    expect(getHospitalizationStatusColor('入院中')).toContain(HEX.blue);
  });

  it('returns gray for 退院済', () => {
    expect(getHospitalizationStatusColor('退院済')).toContain(HEX.gray);
  });

  it('returns green for 予約', () => {
    expect(getHospitalizationStatusColor('予約')).toContain(HEX.green);
  });
});

describe('getHospitalizationTypeColor', () => {
  // 実装: 入院 → purple, ホテル → blue
  it('returns purple for 入院', () => {
    expect(getHospitalizationTypeColor('入院')).toContain(HEX.purple);
  });

  it('returns blue for ホテル', () => {
    expect(getHospitalizationTypeColor('ホテル')).toContain(HEX.blue);
  });
});

describe('getReceptionColumnColor', () => {
  // dot 値は design-tokens.ts の C.bgXxx 値と一致する
  it('returns gray dot for 受付予約', () => {
    const result = getReceptionColumnColor('受付予約');
    expect(result.dot).toContain('9B9A97'); // C.bgStatusGrayMedium
  });

  it('returns blue dot for 受付済', () => {
    const result = getReceptionColumnColor('受付済');
    // primary/brand teal の装飾使用を避け、受付済（checked_in）専用トークンを使う
    expect(result.dot).toContain('blue-500'); // C.bgStatusBlueDot
    expect(result.dot).not.toContain('0075DE'); // primary の装飾使用を恒久ガード
    expect(result.dot).not.toContain('038B94');
  });

  it('returns purple dot for 診療中', () => {
    const result = getReceptionColumnColor('診療中');
    expect(result.dot).toContain('6940A5'); // C.bgStatusPurpleDot
  });

  it('returns orange dot for 会計待ち', () => {
    const result = getReceptionColumnColor('会計待ち');
    expect(result.dot).toContain('D9730D'); // C.bgDiscount
  });

  it('returns green dot for 会計済', () => {
    const result = getReceptionColumnColor('会計済');
    expect(result.dot).toContain('0F7B6C'); // C.bgStatusGreenDot
  });

  it('returns gray dot for unknown column', () => {
    const result = getReceptionColumnColor('unknown');
    expect(result.dot).toContain('9B9A97'); // C.bgStatusGrayMedium (default)
  });
});

describe('getReservationTypeColor', () => {
  it('returns blue for 診療', () => {
    expect(getReservationTypeColor('診療')).toContain(HEX.blue);
  });

  it('returns green for 検診', () => {
    expect(getReservationTypeColor('検診')).toContain(HEX.green);
  });

  it('returns red for 手術', () => {
    expect(getReservationTypeColor('手術')).toContain(HEX.red);
  });

  it('returns orange for トリミング', () => {
    expect(getReservationTypeColor('トリミング')).toContain(HEX.orange);
  });

  it('returns purple for ワクチン', () => {
    expect(getReservationTypeColor('ワクチン')).toContain(HEX.purple);
  });

  it('returns gray for unknown type', () => {
    expect(getReservationTypeColor('unknown')).toContain(HEX.gray);
  });
});

describe('getReservationTypeName', () => {
  it('converts treatment to 診療', () => {
    expect(getReservationTypeName('treatment')).toBe('診療');
  });

  it('converts checkup to 検診', () => {
    expect(getReservationTypeName('checkup')).toBe('検診');
  });

  it('converts surgery to 手術', () => {
    expect(getReservationTypeName('surgery')).toBe('手術');
  });

  it('converts trimming to トリミング', () => {
    expect(getReservationTypeName('trimming')).toBe('トリミング');
  });

  it('converts vaccine to ワクチン', () => {
    expect(getReservationTypeName('vaccine')).toBe('ワクチン');
  });

  it('returns the input for unknown types', () => {
    expect(getReservationTypeName('custom')).toBe('custom');
  });

  it('returns その他 for empty string', () => {
    expect(getReservationTypeName('')).toBe('その他');
  });
});

describe('getExaminationStatusColor', () => {
  it('returns yellow for 依頼中', () => {
    expect(getExaminationStatusColor('依頼中')).toContain(HEX.yellow);
  });

  it('returns blue for 検査中', () => {
    expect(getExaminationStatusColor('検査中')).toContain(HEX.blue);
  });

  // SD-12: result_entered/confirmed 追加の回帰テスト
  it('returns purple for 結果入力済み', () => {
    expect(getExaminationStatusColor('結果入力済み')).toContain(HEX.purple);
  });

  it('returns green for 完了', () => {
    expect(getExaminationStatusColor('完了')).toContain(HEX.green);
  });

  it('returns gray for 確定', () => {
    expect(getExaminationStatusColor('確定')).toContain(HEX.gray);
  });

  it('returns empty string for unknown status', () => {
    expect(getExaminationStatusColor('unknown')).toBe('');
  });
});

describe('getAccountingStatusColor', () => {
  // FE5-27で一覧側が履歴側の4値区別ラベルへ統一された正本ケース
  it('returns orange for 未精算', () => {
    expect(getAccountingStatusColor('未精算')).toContain(HEX.orange);
  });

  it('returns yellow for 保留', () => {
    expect(getAccountingStatusColor('保留')).toContain(HEX.yellow);
  });

  it('returns green for 精算済', () => {
    expect(getAccountingStatusColor('精算済')).toContain(HEX.green);
  });

  it('returns gray for 取消', () => {
    expect(getAccountingStatusColor('取消')).toContain(HEX.gray);
  });

  // 旧ラベルは実装に互換caseとして残っているため引き続き検証する
  it('returns orange for 会計待ち (legacy alias)', () => {
    expect(getAccountingStatusColor('会計待ち')).toContain(HEX.orange);
  });

  it('returns green for 会計済 (legacy alias)', () => {
    expect(getAccountingStatusColor('会計済')).toContain(HEX.green);
  });

  it('returns gray for キャンセル (legacy alias)', () => {
    expect(getAccountingStatusColor('キャンセル')).toContain(HEX.gray);
  });
});

describe('getTrimmingStatusColor', () => {
  it('returns green for 完了', () => {
    expect(getTrimmingStatusColor('完了')).toContain(HEX.green);
  });

  it('returns blue for 予約', () => {
    expect(getTrimmingStatusColor('予約')).toContain(HEX.blue);
  });

  it('returns orange for 進行中', () => {
    expect(getTrimmingStatusColor('進行中')).toContain(HEX.orange);
  });
});

describe('getPetStatusColor', () => {
  it.each([
    ['生存', BADGE.greenHover],
    ['死亡', BADGE.grayHover],
    ['unknown', BADGE.grayHover],
  ])('returns the expected badge token for %s', (status, expected) => {
    expect(getPetStatusColor(status)).toBe(expected);
  });
});

describe('getMasterStatusColor', () => {
  it('returns green for active', () => {
    expect(getMasterStatusColor('active')).toContain(HEX.green);
  });

  it('returns gray for inactive', () => {
    expect(getMasterStatusColor('inactive')).toContain(HEX.gray);
  });
});

describe('getInventoryStatusColor', () => {
  it('returns green for sufficient', () => {
    expect(getInventoryStatusColor('sufficient')).toContain(HEX.green);
  });

  it('returns yellow for low', () => {
    expect(getInventoryStatusColor('low')).toContain(HEX.yellow);
  });

  it('returns red for out_of_stock', () => {
    expect(getInventoryStatusColor('out_of_stock')).toContain(HEX.red);
  });

  it('returns empty string for unknown status', () => {
    expect(getInventoryStatusColor('unknown' as 'sufficient')).toBe('');
  });
});

describe('getInventoryStatusLabel', () => {
  it('returns 十分 for sufficient', () => {
    expect(getInventoryStatusLabel('sufficient')).toBe('十分');
  });

  it('returns 残少 for low', () => {
    expect(getInventoryStatusLabel('low')).toBe('残少');
  });

  it('returns 在庫切れ for out_of_stock', () => {
    expect(getInventoryStatusLabel('out_of_stock')).toBe('在庫切れ');
  });

  it('returns empty string for unknown status', () => {
    expect(getInventoryStatusLabel('unknown' as 'sufficient')).toBe('');
  });
});

describe('getCalendarViewLabel', () => {
  it('returns 月表示 for month', () => {
    expect(getCalendarViewLabel('month')).toBe('月表示');
  });

  it('returns 週表示 for week', () => {
    expect(getCalendarViewLabel('week')).toBe('週表示');
  });

  it('returns the input as-is for unknown view', () => {
    expect(getCalendarViewLabel('day')).toBe('day');
  });
});

describe('getReservationStatusLabel', () => {
  it('returns label for known status', () => {
    // ReservationStatus ラベルは RESERVATION_STATUS_LABELS に依存
    const label = getReservationStatusLabel('confirmed');
    expect(typeof label).toBe('string');
    expect(label.length).toBeGreaterThan(0);
  });

  it('returns status itself for unknown status', () => {
    const unknown = 'unknown_status' as Parameters<typeof getReservationStatusLabel>[0];
    expect(getReservationStatusLabel(unknown)).toBe('unknown_status');
  });
});

describe('getSortOrderLabel', () => {
  it('returns 新→古 for desc', () => {
    expect(getSortOrderLabel('desc')).toBe('新→古');
  });

  it('returns 古→新 for asc', () => {
    expect(getSortOrderLabel('asc')).toBe('古→新');
  });
});

describe('getEstimateStatusColor', () => {
  it('returns gray for draft', () => {
    expect(getEstimateStatusColor('draft')).toContain(HEX.gray);
  });

  it('returns blue for sent', () => {
    expect(getEstimateStatusColor('sent')).toContain(HEX.blue);
  });

  it('returns green for approved', () => {
    expect(getEstimateStatusColor('approved')).toContain(HEX.green);
  });

  it('returns red for rejected', () => {
    expect(getEstimateStatusColor('rejected')).toContain(HEX.red);
  });

  it('returns gray for unknown status', () => {
    expect(getEstimateStatusColor('unknown')).toContain(HEX.gray);
  });
});

describe('getReservationTypeColor — dynamicColorMap', () => {
  it('returns mapped style when dynamicColorMap has the type', () => {
    const style = { color: '#fff', backgroundColor: '#000' };
    const colorMap = new Map([['診療', { style, dotStyle: style, hex: '#000' }]]);
    expect(getReservationTypeColor('診療', colorMap)).toBe(style);
  });

  it('falls back to static mapping when dynamicColorMap has no entry for type', () => {
    const colorMap = new Map<string, { style: React.CSSProperties; dotStyle: React.CSSProperties; hex: string }>();
    expect(getReservationTypeColor('手術', colorMap)).toContain(HEX.red);
  });
});

describe('getHospitalizationStatusColor — all branches', () => {
  it('returns yellow for 一時帰宅', () => {
    expect(getHospitalizationStatusColor('一時帰宅')).toContain(HEX.yellow);
  });

  it('returns empty string for unknown status', () => {
    expect(getHospitalizationStatusColor('unknown' as '入院中')).toBe('');
  });
});

describe('getReservationTypeColor — English aliases', () => {
  it('treatment → blue', () => {
    expect(getReservationTypeColor('treatment')).toContain(HEX.blue);
  });

  it('checkup → green', () => {
    expect(getReservationTypeColor('checkup')).toContain(HEX.green);
  });

  it('surgery → red', () => {
    expect(getReservationTypeColor('surgery')).toContain(HEX.red);
  });

  it('trimming → orange', () => {
    expect(getReservationTypeColor('trimming')).toContain(HEX.orange);
  });

  it('vaccine → purple', () => {
    expect(getReservationTypeColor('vaccine')).toContain(HEX.purple);
  });

  it('入院 → green', () => {
    expect(getReservationTypeColor('入院')).toContain(HEX.green);
  });

  it('ホテル → green', () => {
    expect(getReservationTypeColor('ホテル')).toContain(HEX.green);
  });
});
