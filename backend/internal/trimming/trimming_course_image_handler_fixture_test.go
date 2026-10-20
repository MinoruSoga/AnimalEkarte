package trimming

// Legacy handler fixtures have no course registry row. They do not fabricate ownership.
func legacyUnknownCourseImageDrafts() CourseImageDrafts {
	return &courseImageDrafts{tx: &mockTransactor{}, repo: &draftTestRepo{}, lifecycle: &draftTestLife{}}
}
