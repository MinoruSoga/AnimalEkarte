import { describe, it, expect } from 'vitest';
import {
  getMedicalRecordStatusColor,
  getHospitalizationStatusColor,
  getHospitalizationTypeColor,
  getDashboardColumnColor,
  getReservationTypeColor,
  getReservationTypeName,
  getExaminationStatusColor,
  getAccountingStatusColor,
  getTrimmingStatusColor,
  getPetStatusColor,
  getMasterStatusColor,
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

describe('getDashboardColumnColor', () => {
  // dot 値は design-tokens.ts の C.bgXxx 値と一致する
  it('returns gray dot for 受付予約', () => {
    const result = getDashboardColumnColor('受付予約');
    expect(result.dot).toContain('9B9A97'); // C.bgStatusGrayMedium
  });

  it('returns blue dot for 受付済', () => {
    const result = getDashboardColumnColor('受付済');
    expect(result.dot).toContain('2383E2'); // C.bgAccent
  });

  it('returns purple dot for 診療中', () => {
    const result = getDashboardColumnColor('診療中');
    expect(result.dot).toContain('6940A5'); // C.bgStatusPurpleDot
  });

  it('returns orange dot for 会計待ち', () => {
    const result = getDashboardColumnColor('会計待ち');
    expect(result.dot).toContain('D9730D'); // C.bgDiscount
  });

  it('returns green dot for 会計済', () => {
    const result = getDashboardColumnColor('会計済');
    expect(result.dot).toContain('0F7B6C'); // C.bgStatusGreenDot
  });

  it('returns gray dot for unknown column', () => {
    const result = getDashboardColumnColor('unknown');
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

  it('returns green for 完了', () => {
    expect(getExaminationStatusColor('完了')).toContain(HEX.green);
  });
});

describe('getAccountingStatusColor', () => {
  // 実装のステータス値: 会計待ち/会計済/キャンセル
  it('returns orange for 会計待ち', () => {
    expect(getAccountingStatusColor('会計待ち')).toContain(HEX.orange);
  });

  it('returns green for 会計済', () => {
    expect(getAccountingStatusColor('会計済')).toContain(HEX.green);
  });

  it('returns gray for キャンセル', () => {
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
  it('returns green for 生存', () => {
    expect(getPetStatusColor('生存')).toContain(HEX.green);
  });

  it('returns gray for other status', () => {
    expect(getPetStatusColor('死亡')).toContain(HEX.gray);
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
