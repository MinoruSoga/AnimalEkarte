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

describe('getMedicalRecordStatusColor', () => {
  it('returns blue for 作成中', () => {
    expect(getMedicalRecordStatusColor('作成中')).toContain('blue');
  });

  it('returns gray for 確定済', () => {
    expect(getMedicalRecordStatusColor('確定済')).toContain('gray');
  });

  it('returns empty string for unknown status', () => {
    expect(getMedicalRecordStatusColor('unknown' as '作成中')).toBe('');
  });
});

describe('getHospitalizationStatusColor', () => {
  it('returns blue for 入院中', () => {
    expect(getHospitalizationStatusColor('入院中')).toContain('blue');
  });

  it('returns gray for 退院済', () => {
    expect(getHospitalizationStatusColor('退院済')).toContain('gray');
  });

  it('returns green for 予約', () => {
    expect(getHospitalizationStatusColor('予約')).toContain('green');
  });
});

describe('getHospitalizationTypeColor', () => {
  it('returns red for 入院', () => {
    expect(getHospitalizationTypeColor('入院')).toContain('red');
  });

  it('returns purple for ホテル', () => {
    expect(getHospitalizationTypeColor('ホテル')).toContain('purple');
  });
});

describe('getDashboardColumnColor', () => {
  it('returns correct colors for 受付予約', () => {
    const result = getDashboardColumnColor('受付予約');
    expect(result.dot).toContain('gray');
  });

  it('returns correct colors for 受付済', () => {
    const result = getDashboardColumnColor('受付済');
    expect(result.dot).toContain('blue');
  });

  it('returns correct colors for 診療中', () => {
    const result = getDashboardColumnColor('診療中');
    expect(result.dot).toContain('yellow');
  });

  it('returns correct colors for 会計待ち', () => {
    const result = getDashboardColumnColor('会計待ち');
    expect(result.dot).toContain('orange');
  });

  it('returns correct colors for 会計済', () => {
    const result = getDashboardColumnColor('会計済');
    expect(result.dot).toContain('green');
  });

  it('returns default colors for unknown column', () => {
    const result = getDashboardColumnColor('unknown');
    expect(result.dot).toContain('gray');
  });
});

describe('getReservationTypeColor', () => {
  it('returns blue for 診療', () => {
    expect(getReservationTypeColor('診療')).toContain('blue');
  });

  it('returns green for 検診', () => {
    expect(getReservationTypeColor('検診')).toContain('green');
  });

  it('returns red for 手術', () => {
    expect(getReservationTypeColor('手術')).toContain('red');
  });

  it('returns orange for トリミング', () => {
    expect(getReservationTypeColor('トリミング')).toContain('orange');
  });

  it('returns purple for ワクチン', () => {
    expect(getReservationTypeColor('ワクチン')).toContain('purple');
  });

  it('returns gray for unknown type', () => {
    expect(getReservationTypeColor('unknown')).toContain('gray');
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
    expect(getExaminationStatusColor('依頼中')).toContain('yellow');
  });

  it('returns blue for 検査中', () => {
    expect(getExaminationStatusColor('検査中')).toContain('blue');
  });

  it('returns green for 完了', () => {
    expect(getExaminationStatusColor('完了')).toContain('green');
  });
});

describe('getAccountingStatusColor', () => {
  it('returns red for 未収', () => {
    expect(getAccountingStatusColor('未収')).toContain('red');
  });

  it('returns green for 回収済', () => {
    expect(getAccountingStatusColor('回収済')).toContain('green');
  });

  it('returns gray for キャンセル', () => {
    expect(getAccountingStatusColor('キャンセル')).toContain('gray');
  });
});

describe('getTrimmingStatusColor', () => {
  it('returns green for 完了', () => {
    expect(getTrimmingStatusColor('完了')).toContain('E8F5E9');
  });

  it('returns blue for 予約', () => {
    expect(getTrimmingStatusColor('予約')).toContain('E3F2FD');
  });

  it('returns orange for 進行中', () => {
    expect(getTrimmingStatusColor('進行中')).toContain('FFF3E0');
  });
});

describe('getPetStatusColor', () => {
  it('returns green for 生存', () => {
    expect(getPetStatusColor('生存')).toContain('DDEDEA');
  });

  it('returns gray for other status', () => {
    expect(getPetStatusColor('死亡')).toContain('EBECED');
  });
});

describe('getMasterStatusColor', () => {
  it('returns green for active', () => {
    expect(getMasterStatusColor('active')).toContain('green');
  });

  it('returns gray for inactive', () => {
    expect(getMasterStatusColor('inactive')).toContain('gray');
  });
});
