import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { PALETTE } from '@/lib/design-tokens';
import {
  LIFF_BRAND_TOKEN_VALUES,
  NOAH_BRAND_COLORS,
  resolveClinicHeaderText,
} from './brand-tokens';

describe('shared brand tokens (BUG-026)', () => {
  it('LIFF brand token values match LINE-reserve noah teal palette (not mint green)', () => {
    expect(LIFF_BRAND_TOKEN_VALUES['liff-brand']).toBe(NOAH_BRAND_COLORS.teal);
    expect(LIFF_BRAND_TOKEN_VALUES['liff-brand-dark']).toBe(NOAH_BRAND_COLORS.tealDark);
    expect(LIFF_BRAND_TOKEN_VALUES['liff-brand-bg']).toBe(NOAH_BRAND_COLORS.tealLight);
    expect(LIFF_BRAND_TOKEN_VALUES['liff-brand']).toBe('#038B94');
    expect(LIFF_BRAND_TOKEN_VALUES['liff-brand-bg']).toBe('#EDF3F5');
  });

  it('noah teal equals design-system PALETTE.brand (FE-RC-014)', () => {
    expect(NOAH_BRAND_COLORS.teal).toBe(PALETTE.brand);
    expect(NOAH_BRAND_COLORS.tealDark).toBe(PALETTE.brandHover);
  });

  it('brand-tokens.css defines the same hex values for noah and liff aliases', () => {
    const cssPath = resolve(import.meta.dirname, 'brand-tokens.css');
    const css = readFileSync(cssPath, 'utf8');
    expect(css).toContain('--color-noah-teal: #038B94');
    expect(css).toContain('--color-noah-teal-dark: #027078');
    expect(css).toContain('--color-noah-teal-light: #EDF3F5');
    expect(css).toContain('--color-liff-brand: #038B94');
    expect(css).toContain('--color-liff-brand-bg: #EDF3F5');
    expect(css).not.toMatch(/green-50|155\.826/);
    expect(css).not.toMatch(/#008B94|#007079/);
  });

  it('resolveClinicHeaderText trims and never fabricates a clinic name', () => {
    expect(resolveClinicHeaderText('  ノア動物病院 八王子  ')).toBe('ノア動物病院 八王子');
    expect(resolveClinicHeaderText('')).toBe('');
    expect(resolveClinicHeaderText('   ')).toBe('');
    expect(resolveClinicHeaderText(null)).toBe('');
    expect(resolveClinicHeaderText(undefined)).toBe('');
  });
});
