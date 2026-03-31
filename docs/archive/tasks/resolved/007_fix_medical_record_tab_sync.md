# Task: Fix Medical Record Tab Synchronization ("Diagnosis/Plan" vs "Treatment")

## Description
The "Diagnosis/Plan" tab and the "Treatment" tab both handle treatment items but suffer from inconsistent state synchronization when navigating between them.

## Problems Identified
- **New Rows Disappearing**: Adding a row in the "Diagnosis/Plan" tab might not immediately appear in the "Treatment" tab.
- **Status Linkage**: Checking a plan as "Completed" in the "Diagnosis/Plan" checklist does not always correctly update the billing status in the "Treatment" tab.
- **Stale Data on Switch**: Navigating between these tabs without saving can lead to data drift where one tab shows newer information than the other.

## Requirements
- Synchronize the treatment list state globally across all tabs in the Medical Record form.
- Ensure that adding, updating, or deleting a treatment item in one tab triggers an immediate update in all other tabs.
- Unify the underlying data structure to prevent mapping discrepancies between different views.
