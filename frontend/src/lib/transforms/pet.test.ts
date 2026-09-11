import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, it, expect } from "vitest";
import ts from "typescript";

import {
  mapPetStatusLabel,
  transformBackendPetToFrontend,
  transformCreatePetRequest,
  transformUpdatePetRequest,
} from "./pet";
import type { CreatePetRequest, UpdatePetRequest } from "@/types/pet";
import type { PetResponse } from "@/types/generated/pet-responses";

// src/lib/transforms → frontend root (container: /app)
const FRONTEND_ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const PET_TYPES_PATH = path.join(FRONTEND_ROOT, "src/types/pet.ts");

/** Compile-time keys that must never appear on create/update request types. */
type ForbiddenWritableKeys = "version" | "deceased_at" | "deceased_reason";
type AssertNever<T extends never> = T;
type _CreateForbidsServerColumns = AssertNever<
  Extract<keyof CreatePetRequest, ForbiddenWritableKeys>
>;
type _UpdateForbidsServerColumns = AssertNever<
  Extract<keyof UpdatePetRequest, ForbiddenWritableKeys>
>;
void 0 as unknown as _CreateForbidsServerColumns;
void 0 as unknown as _UpdateForbidsServerColumns;

// makeBackendPet は transformBackendPetToFrontend に渡す最小の PetResponse を組み立てる。
function makeBackendPet(overrides: Partial<PetResponse> = {}): PetResponse {
  return {
    id: 7,
    version: 1,
    clinic_id: 1,
    owner_id: 42,
    animal_species_id: 1,
    pet_number: "42-1",
    name: "ポチ",
    pet_name_kana: "ぽち",
    gender: "male",
    status: "alive",
    breed: "",
    color: "",
    danger_level: "low",
    food: "",
    environment: "",
    phone: "",
    remarks: "",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("transformBackendPetToFrontend", () => {
  it("last_visit を birth_date / neutered_date と同じ『日付のみ』形式へ正規化する", () => {
    const pet = transformBackendPetToFrontend(
      makeBackendPet({
        birth_date: "2015-04-14T00:00:00Z",
        neutered_date: "2016-05-20T00:00:00Z",
        last_visit: "2024-08-25T00:00:00Z",
      }),
    );

    expect(pet.birthDate).toBe("2015-04-14");
    expect(pet.neuteredDate).toBe("2016-05-20");
    // #158: 以前は ISO datetime をそのまま通す非対称があった。兄弟の日付フィールドと揃える。
    expect(pet.lastVisit).toBe("2024-08-25");
  });

  it("日付フィールドが未設定なら undefined のまま（" - " 化やゼロ値で潰さない）", () => {
    const pet = transformBackendPetToFrontend(
      makeBackendPet({ birth_date: undefined, neutered_date: undefined, last_visit: undefined }),
    );

    expect(pet.birthDate).toBeUndefined();
    expect(pet.neuteredDate).toBeUndefined();
    expect(pet.lastVisit).toBeUndefined();
  });

  it("血液型 / マイクロチップ番号をマッピングする", () => {
    const pet = transformBackendPetToFrontend(
      makeBackendPet({ blood_type: "DEA1.1陽性", microchip_number: "392140000123456" }),
    );

    expect(pet.bloodType).toBe("DEA1.1陽性");
    expect(pet.microchipNumber).toBe("392140000123456");
  });

  it("血液型 / マイクロチップ番号 未設定は undefined（捏造しない）", () => {
    const pet = transformBackendPetToFrontend(makeBackendPet());

    expect(pet.bloodType).toBeUndefined();
    expect(pet.microchipNumber).toBeUndefined();
  });

  it("作成リクエストへ血液型 / マイクロチップ番号を含める", () => {
    const request = transformCreatePetRequest({
      ownerId: "42",
      name: "ポチ",
      animalSpeciesId: "1",
      bloodType: "DEA1.1陽性",
      microchipNumber: "392140000123456",
    });

    expect(request.blood_type).toBe("DEA1.1陽性");
    expect(request.microchip_number).toBe("392140000123456");
  });

  it("更新リクエストへ血液型 / マイクロチップ番号を含める", () => {
    const request = transformUpdatePetRequest({
      bloodType: "B",
      microchipNumber: "900000000000001",
    });

    expect(request.blood_type).toBe("B");
    expect(request.microchip_number).toBe("900000000000001");
  });

  // PR#186 P2-2 Bug#1 回帰テスト: deceased_at は response DTO への追加のみで
  // transform 層の配線が漏れると、値が API から届いても UI に渡らない。
  // BUG-003: deceased_reason も staff PetResponse 経由で deceasedReason へ。
  it("deceased_at を deceasedAt へマッピングする", () => {
    const pet = transformBackendPetToFrontend(
      makeBackendPet({
        status: "deceased",
        deceased_at: "2026-07-10T12:00:00+09:00",
        deceased_reason: "老衰",
      }),
    );

    expect(pet.status).toBe("死亡");
    expect(pet.deceasedAt).toBe("2026-07-10T12:00:00+09:00");
    expect(pet.deceasedReason).toBe("老衰");
  });

  it("deceased_at 未設定（生存中）は undefined のまま（捏造しない）", () => {
    const pet = transformBackendPetToFrontend(makeBackendPet());

    expect(pet.deceasedAt).toBeUndefined();
    expect(pet.deceasedReason).toBeUndefined();
  });

  it("未知statusは生存へ推測せず不明にする", () => {
    const pet = transformBackendPetToFrontend(makeBackendPet({ status: "unexpected" }));

    expect(pet.status).toBe("不明");
  });

  it.each([
    ["unexpected", "未知値"],
    ["constructor", "Object prototype key"],
    ["toString", "Object prototype method"],
    ["__proto__", "Object prototype accessor"],
    ["", "空文字"],
    [null, "null"],
    [undefined, "未指定"],
  ])("API境界のstatus %s は不明にする（%s）", (status) => {
    expect(mapPetStatusLabel(status)).toBe("不明");
  });

  // BUG-415: generic PATCH /pets/:id 経由の status 書込は deceased_at・監査ログと
  // 無結合のため除去した。status 変更は監査付きの死亡登録/取消
  // (PetDeceasedRecordButton → /:id/death)に一本化済み。
  // このテストは修正前は落ちていた(RED): 旧実装は status を無条件送信していたため
  // request.status が "alive" になり toBeUndefined() は失敗していた。
  it("更新リクエストは status を送信しない（死亡/復活は /:id/death に一本化）", () => {
    const request = transformUpdatePetRequest({
      name: "ポチ",
      status: "alive",
    });

    expect(request.status).toBeUndefined();
  });

  it("pet_name_kana を petNameKana へマッピングする（models.Pet の name_kana ではない）", () => {
    const pet = transformBackendPetToFrontend(makeBackendPet({ pet_name_kana: "ぽちたろう" }));
    expect(pet.petNameKana).toBe("ぽちたろう");
  });
});

describe("transformUpdatePetRequest", () => {
  it("既存の危険理由をクリアすると danger_reason を null として送信する", () => {
    const request = transformUpdatePetRequest({
      dangerReason: "",
      originalDangerReason: "咬傷歴あり",
    });

    expect(Object.prototype.hasOwnProperty.call(request, "danger_reason")).toBe(true);
    expect(request.danger_reason).toBeNull();
  });

  it("危険理由が未変更なら danger_reason を送信しない", () => {
    const request = transformUpdatePetRequest({
      dangerReason: "咬傷歴あり",
      originalDangerReason: "咬傷歴あり",
    });

    expect(Object.prototype.hasOwnProperty.call(request, "danger_reason")).toBe(false);
  });
});

describe("PetWritable allowlist (TASK-444)", () => {
  it("PetWritable は Omit ではなく Pick 許可リストで定義する", () => {
    const src = readFileSync(PET_TYPES_PATH, "utf8");
    const writable = src.match(/type\s+PetWritable\s*=\s*[^;]+;/)?.[0] ?? "";
    expect(writable, "PetWritable declaration").toMatch(/Pick\s*</);
    expect(writable).not.toMatch(/Omit\s*</);
    expect(writable).not.toMatch(/\bversion\b/);
    expect(writable).not.toMatch(/\bdeceased_at\b/);
    expect(writable).not.toMatch(/\bdeceased_reason\b/);
  });

  it("CreatePetRequest / UpdatePetRequest へ version・deceased_* を代入すると型エラーになる", () => {
    const configPath = ts.findConfigFile(FRONTEND_ROOT, ts.sys.fileExists, "tsconfig.json");
    expect(configPath, "tsconfig.json").toBeTruthy();
    const configFile = ts.readConfigFile(configPath!, ts.sys.readFile);
    const parsed = ts.parseJsonConfigFileContent(
      configFile.config,
      ts.sys,
      FRONTEND_ROOT,
      undefined,
      configPath,
    );
    const fixturePath = path.join(FRONTEND_ROOT, "src/types/__task444_pet_writable_fixture__.ts");
    // One excess key per object so TS reports each forbidden property (not only the first).
    const fixtureSource = [
      'import type { CreatePetRequest, UpdatePetRequest } from "@/types/pet";',
      "export const createVersion: CreatePetRequest = {",
      '  owner_id: 1, animal_species_id: 1, name: "ポチ", version: 1,',
      "};",
      "export const createDeceasedAt: CreatePetRequest = {",
      '  owner_id: 1, animal_species_id: 1, name: "ポチ", deceased_at: "2026-01-01T00:00:00Z",',
      "};",
      "export const createDeceasedReason: CreatePetRequest = {",
      '  owner_id: 1, animal_species_id: 1, name: "ポチ", deceased_reason: "x",',
      "};",
      "export const updateVersion: UpdatePetRequest = { version: 2 };",
      'export const updateDeceasedAt: UpdatePetRequest = { deceased_at: "2026-01-01T00:00:00Z" };',
      'export const updateDeceasedReason: UpdatePetRequest = { deceased_reason: "y" };',
      "",
    ].join("\n");

    const options: ts.CompilerOptions = {
      ...parsed.options,
      noEmit: true,
      // Virtual fixture only — keep diagnostics local to assignability.
      skipLibCheck: true,
    };
    const host = ts.createCompilerHost(options, true);
    const resolvedFixture = path.resolve(fixturePath);
    const originalGetSourceFile = host.getSourceFile.bind(host);
    host.getSourceFile = (fileName, languageVersion, onError, shouldCreateNewSourceFile) => {
      if (path.resolve(fileName) === resolvedFixture) {
        return ts.createSourceFile(fixturePath, fixtureSource, languageVersion, true);
      }
      return originalGetSourceFile(fileName, languageVersion, onError, shouldCreateNewSourceFile);
    };
    const originalFileExists = host.fileExists.bind(host);
    host.fileExists = (fileName) =>
      path.resolve(fileName) === resolvedFixture || originalFileExists(fileName);
    const originalReadFile = host.readFile.bind(host);
    host.readFile = (fileName) =>
      path.resolve(fileName) === resolvedFixture ? fixtureSource : originalReadFile(fileName);

    const program = ts.createProgram({
      rootNames: [fixturePath, PET_TYPES_PATH],
      options,
      host,
    });
    const diagnostics = ts
      .getPreEmitDiagnostics(program)
      .filter((d) => d.file && path.resolve(d.file.fileName) === resolvedFixture)
      .filter((d) => d.category === ts.DiagnosticCategory.Error);
    const messages = diagnostics.map((d) => ts.flattenDiagnosticMessageText(d.messageText, "\n"));
    const joined = messages.join("\n");

    expect(diagnostics.length, joined).toBeGreaterThan(0);
    expect(joined).toMatch(/version/);
    expect(joined).toMatch(/deceased_at/);
    expect(joined).toMatch(/deceased_reason/);
  });

  it("作成 transform は必須3キーと name_kana・status を送り、死亡列は送らない", () => {
    const request = transformCreatePetRequest({
      ownerId: "42",
      name: "ポチ",
      animalSpeciesId: "1",
      petNameKana: "ぽち",
      status: "alive",
    });
    expect(request).toMatchObject({
      owner_id: 42,
      name: "ポチ",
      animal_species_id: 1,
      name_kana: "ぽち",
      status: "alive",
    });
    expect(request).not.toHaveProperty("version");
    expect(request).not.toHaveProperty("deceased_at");
    expect(request).not.toHaveProperty("deceased_reason");
  });
});
