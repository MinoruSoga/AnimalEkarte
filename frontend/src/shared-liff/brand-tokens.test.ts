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

  it('FE-RC-051: brand-tokens.css defines semantic status tokens aliased to Tailwind defaults (no new colors)', () => {
    const cssPath = resolve(import.meta.dirname, 'brand-tokens.css');
    const css = readFileSync(cssPath, 'utf8');
    // 値は Tailwind 既定パレットの var() 参照でエイリアスする（見た目を変えない）。
    expect(css).toContain('--color-noah-success-bg: var(--color-green-100)');
    expect(css).toContain('--color-noah-success-text: var(--color-green-800)');
    expect(css).toContain('--color-noah-warning-bg: var(--color-yellow-100)');
    expect(css).toContain('--color-noah-info-bg: var(--color-blue-100)');
    expect(css).toContain('--color-noah-danger: var(--color-red-500)');
    expect(css).toContain('--color-noah-danger-bg: var(--color-red-50)');
    expect(css).toContain('--color-noah-neutral-bg: var(--color-gray-100)');
  });

  it('resolveClinicHeaderText trims and never fabricates a clinic name', () => {
    expect(resolveClinicHeaderText('  ノア動物病院 八王子  ')).toBe('ノア動物病院 八王子');
    expect(resolveClinicHeaderText('')).toBe('');
    expect(resolveClinicHeaderText('   ')).toBe('');
    expect(resolveClinicHeaderText(null)).toBe('');
    expect(resolveClinicHeaderText(undefined)).toBe('');
  });
});
