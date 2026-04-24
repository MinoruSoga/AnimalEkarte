# FE-012: Missing Interview Settings Routes

## Issue
Navigating to `/settings/interview/chief-complaint` returns 404 "ページが見つかりません". This breaks the master edit button in MasterSelectModal when used with chief complaint items.

## Root Cause Analysis

1. **Router missing routes**: `app/router.tsx` has no route definitions for `/settings/interview/*`
   - Defined in `paths.ts` (lines 232-241)
   - Missing from `router.tsx` entirely

2. **Missing route components**: No components exist in `features/master/routes/` for:
   - ChiefComplaintSettings.tsx
   - InterviewTemplateSettings.tsx

3. **MasterLink doesn't handle interview categories**: `components/shared/MasterLink.tsx` CATEGORY_PATH_MAP (lines 21-36) has no mapping for "chief_complaint" or interview-related categories

## Acceptance Criteria

- [ ] Create `ChiefComplaintSettings.tsx` in `features/master/routes/`
- [ ] Create `InterviewTemplateSettings.tsx` in `features/master/routes/` (or reuse/extend InquiryTemplateSettings)
- [ ] Add routes in `router.tsx` for:
  - `/settings/interview/chief-complaint` → ChiefComplaintSettings
  - `/settings/interview/templates` → InterviewTemplateSettings
- [ ] Update `MasterLink.tsx` CATEGORY_PATH_MAP to handle:
  - `"chief_complaint"` → `/settings/interview/chief-complaint`
  - `"interview_template"` → `/settings/interview/templates`

## Related

- FE-009 (Master edit button navigation) claims this is fixed but it's broken
- Depends on: Backend chief complaint API endpoint (`GET /v1/masters/chief-complaints`)
