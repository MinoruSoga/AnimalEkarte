package medicalrecord

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
)

func TestExaminationRevision_RepositoryMethodsParticipateInAmbientTransaction(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	actorID := makeExaminationActor(t, db, clinicID, "ambient transaction actor")
	examType := makeExamTypeMaster(t, db, clinicID, "ambient transaction exam type")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID:   clinicID,
		ExamTypeID: examType.ID,
		Date:       time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusCompleted,
	})
	repository := NewExaminationRepository(db)
	revisions, ok := repository.(ExaminationRevisionRepository)
	require.True(t, ok)
	errRollback := errors.New("rollback revision repository ambient transaction")

	err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		version, appendErr := revisions.AppendOfficialRevision(
			txCtx,
			clinicID,
			exam.ID,
			actorID,
			examinationInitialConfirmReason,
		)
		require.NoError(t, appendErr)
		assert.Equal(t, initialExaminationRevisionVersion, version)

		uncommittedOfficial, readErr := revisions.FindOfficialByID(txCtx, clinicID, exam.ID)
		require.NoError(t, readErr)
		require.NotNil(t, uncommittedOfficial)
		assert.Equal(t, initialExaminationRevisionVersion, uncommittedOfficial.OfficialVersion)
		assert.Nil(t, uncommittedOfficial.CurrentRevisionVersion)

		uncommittedParent, confirmErr := revisions.ConfirmWithRevisionCAS(
			txCtx,
			clinicID,
			exam.ID,
			model.ExaminationStatusCompleted,
			version,
		)
		require.NoError(t, confirmErr)
		require.NotNil(t, uncommittedParent.CurrentRevisionVersion)
		assert.Equal(t, version, *uncommittedParent.CurrentRevisionVersion)
		assert.Equal(t, model.ExaminationStatusConfirmed, uncommittedParent.Status)

		return errRollback
	})
	require.ErrorIs(t, err, errRollback)

	persisted, err := repository.FindByID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	assert.Equal(t, model.ExaminationStatusCompleted, persisted.Status)
	assert.Nil(t, persisted.CurrentRevisionVersion)
	assertExaminationRevisionRows(t, db, clinicID, exam.ID, 0, 0)
	rolledBackOfficial, err := revisions.FindOfficialByID(ctx, clinicID, exam.ID)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, rolledBackOfficial)
}

func TestExaminationRevision_ModelMatchesFrozenSchemaContract(t *testing.T) {
	tests := []struct {
		name         string
		modelType    reflect.Type
		fieldName    string
		wantType     string
		tagFragments []string
	}{
		{
			name:      "schema version uses smallint one",
			modelType: reflect.TypeOf(model.ExaminationRevision{}), fieldName: "SchemaVersion",
			wantType: "int16", tagFragments: []string{"type:smallint", "not null", "default:1"},
		},
		{
			name:      "change reason is nullable",
			modelType: reflect.TypeOf(model.ExaminationRevision{}), fieldName: "ChangeReason",
			wantType: "*string", tagFragments: []string{"type:text"},
		},
		{
			name:      "revision created at has database default",
			modelType: reflect.TypeOf(model.ExaminationRevision{}), fieldName: "CreatedAt",
			wantType: "time.Time", tagFragments: []string{"not null", "default:CURRENT_TIMESTAMP", "autoCreateTime"},
		},
		{
			name:      "revision item created at has database default",
			modelType: reflect.TypeOf(model.ExaminationRevisionItem{}), fieldName: "CreatedAt",
			wantType: "time.Time", tagFragments: []string{"not null", "default:CURRENT_TIMESTAMP", "autoCreateTime"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := tt.modelType.FieldByName(tt.fieldName)
			require.True(t, ok)
			assert.Equal(t, tt.wantType, field.Type.String())
			gormTag := field.Tag.Get("gorm")
			for _, fragment := range tt.tagFragments {
				assert.Contains(t, gormTag, fragment)
			}
			if tt.fieldName == "ChangeReason" {
				assert.NotContains(t, gormTag, "not null")
			}
		})
	}
}

func TestExaminationRevisionMigration_SealsSelectedItemVersions(t *testing.T) {
	withExaminationRevisionMigrationSchema(t, func(db *gorm.DB) {
		tests := []struct {
			name     string
			wantCode string
			run      func(*gorm.DB) error
		}{
			{
				name:     "late insert into selected version is rejected",
				wantCode: "23514",
				run: func(db *gorm.DB) error {
					return db.Exec(`
						INSERT INTO examination_revision_items (
							id, clinic_id, examination_id, version, name,
							is_assessed, is_abnormal, status
						) VALUES (2, 1, 10, 1, 'late item', FALSE, FALSE, 'normal')
					`).Error
				},
			},
			{
				name: "next version item before pointer CAS is allowed",
				run: func(db *gorm.DB) error {
					if err := insertExaminationRevisionVersion(db, 2); err != nil {
						return err
					}
					return db.Exec(`
						INSERT INTO examination_revision_items (
							id, clinic_id, examination_id, version, name,
							is_assessed, is_abnormal, status
						) VALUES (2, 1, 10, 2, 'next item', FALSE, FALSE, 'normal')
					`).Error
				},
			},
			{
				name:     "version gap is rejected",
				wantCode: "23514",
				run: func(db *gorm.DB) error {
					if err := insertExaminationRevisionVersion(db, 3); err != nil {
						return err
					}
					return db.Exec(`
						INSERT INTO examination_revision_items (
							id, clinic_id, examination_id, version, name,
							is_assessed, is_abnormal, status
						) VALUES (3, 1, 10, 3, 'gap item', FALSE, FALSE, 'normal')
					`).Error
				},
			},
		}

		for i, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				savepoint := fmt.Sprintf("examination_revision_item_seal_%d", i)
				require.NoError(t, db.Exec("SAVEPOINT "+savepoint).Error)
				defer func() {
					require.NoError(t, db.Exec("ROLLBACK TO SAVEPOINT "+savepoint).Error)
				}()

				err := tt.run(db)
				if tt.wantCode == "" {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					assertPostgresCode(t, err, tt.wantCode)
				}
			})
		}
	})
}

func TestExaminationRevisionMigration_CatalogMatchesFrozenContract(t *testing.T) {
	withExaminationRevisionMigrationSchema(t, func(db *gorm.DB) {
		type foreignKeyMetadata struct {
			Name              string
			ChildTable        string
			ParentTable       string
			DeleteAction      string
			Deferrable        bool
			InitiallyDeferred bool
			Definition        string
		}
		var foreignKeys []foreignKeyMetadata
		require.NoError(t, db.Raw(`
			SELECT
				constraint_row.conname AS name,
				child.relname AS child_table,
				parent.relname AS parent_table,
				constraint_row.confdeltype::text AS delete_action,
				constraint_row.condeferrable AS deferrable,
				constraint_row.condeferred AS initially_deferred,
				pg_get_constraintdef(constraint_row.oid) AS definition
			FROM pg_constraint constraint_row
			JOIN pg_class child ON child.oid = constraint_row.conrelid
			JOIN pg_namespace child_namespace ON child_namespace.oid = child.relnamespace
			JOIN pg_class parent ON parent.oid = constraint_row.confrelid
			WHERE child_namespace.nspname = current_schema()
			  AND constraint_row.contype = 'f'
			  AND constraint_row.conname IN (
				'fk_examination_revisions_exam',
				'fk_examination_revision_items_revision',
				'fk_exams_current_revision'
			  )
			ORDER BY constraint_row.conname
		`).Scan(&foreignKeys).Error)
		require.Len(t, foreignKeys, 3)
		expectedForeignKeys := map[string]struct {
			child, parent, columns string
		}{
			"fk_examination_revisions_exam": {
				child: "examination_revisions", parent: "exams",
				columns: "FOREIGN KEY (clinic_id, examination_id) REFERENCES exams(clinic_id, id) ON DELETE RESTRICT",
			},
			"fk_examination_revision_items_revision": {
				child: "examination_revision_items", parent: "examination_revisions",
				columns: "FOREIGN KEY (clinic_id, examination_id, version) REFERENCES examination_revisions(clinic_id, examination_id, version) ON DELETE RESTRICT",
			},
			"fk_exams_current_revision": {
				child: "exams", parent: "examination_revisions",
				columns: "FOREIGN KEY (clinic_id, id, current_revision_version) REFERENCES examination_revisions(clinic_id, examination_id, version) ON DELETE RESTRICT",
			},
		}
		for _, metadata := range foreignKeys {
			want := expectedForeignKeys[metadata.Name]
			assert.Equal(t, want.child, metadata.ChildTable)
			assert.Equal(t, want.parent, metadata.ParentTable)
			assert.Equal(t, "r", metadata.DeleteAction)
			assert.False(t, metadata.Deferrable)
			assert.False(t, metadata.InitiallyDeferred)
			assert.Contains(t, metadata.Definition, want.columns)
		}

		type indexMetadata struct {
			TableName  string
			IndexName  string
			Definition string
			Nonpartial bool
		}
		var indexes []indexMetadata
		require.NoError(t, db.Raw(`
			SELECT
				table_row.relname AS table_name,
				index_row.relname AS index_name,
				pg_get_indexdef(index_row.oid) AS definition,
				index_metadata.indpred IS NULL AS nonpartial
			FROM pg_index index_metadata
			JOIN pg_class table_row ON table_row.oid = index_metadata.indrelid
			JOIN pg_namespace table_namespace ON table_namespace.oid = table_row.relnamespace
			JOIN pg_class index_row ON index_row.oid = index_metadata.indexrelid
			WHERE table_namespace.nspname = current_schema()
			  AND index_row.relname IN (
				'uq_examination_revisions_clinic_exam_version',
				'idx_examination_revision_items_revision_sort',
				'idx_exams_current_revision_pointer'
			  )
		`).Scan(&indexes).Error)
		require.Len(t, indexes, 3)
		expectedIndexes := map[string]struct {
			table, prefix string
		}{
			"uq_examination_revisions_clinic_exam_version": {
				table: "examination_revisions", prefix: "(clinic_id, examination_id, version)",
			},
			"idx_examination_revision_items_revision_sort": {
				table: "examination_revision_items", prefix: "(clinic_id, examination_id, version, sort_order, id)",
			},
			"idx_exams_current_revision_pointer": {
				table: "exams", prefix: "(clinic_id, id, current_revision_version)",
			},
		}
		for _, metadata := range indexes {
			want := expectedIndexes[metadata.IndexName]
			assert.Equal(t, want.table, metadata.TableName)
			assert.True(t, metadata.Nonpartial)
			assert.Contains(t, metadata.Definition, want.prefix)
		}

		type columnMetadata struct {
			TableName     string
			ColumnName    string
			DataType      string
			IsNullable    string
			ColumnDefault *string
		}
		var columns []columnMetadata
		require.NoError(t, db.Raw(`
			SELECT table_name, column_name, data_type, is_nullable, column_default
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND (
				(table_name = 'exams' AND column_name = 'current_revision_version') OR
				(table_name = 'examination_revisions' AND column_name IN ('schema_version', 'change_reason', 'created_at')) OR
				(table_name = 'examination_revision_items' AND column_name = 'created_at')
			  )
		`).Scan(&columns).Error)
		columnByKey := make(map[string]columnMetadata, len(columns))
		for _, column := range columns {
			columnByKey[column.TableName+"."+column.ColumnName] = column
		}
		assert.Equal(t, "YES", columnByKey["exams.current_revision_version"].IsNullable)
		assert.Equal(t, "bigint", columnByKey["exams.current_revision_version"].DataType)
		assert.Equal(t, "smallint", columnByKey["examination_revisions.schema_version"].DataType)
		assert.Equal(t, "NO", columnByKey["examination_revisions.schema_version"].IsNullable)
		assert.Equal(t, "YES", columnByKey["examination_revisions.change_reason"].IsNullable)
		for _, key := range []string{
			"examination_revisions.created_at",
			"examination_revision_items.created_at",
		} {
			column := columnByKey[key]
			assert.Equal(t, "NO", column.IsNullable)
			require.NotNil(t, column.ColumnDefault)
			assert.Contains(t, strings.ToLower(*column.ColumnDefault), "current_timestamp")
		}

		var checkDefinitions []string
		require.NoError(t, db.Raw(`
			SELECT pg_get_constraintdef(constraint_row.oid)
			FROM pg_constraint constraint_row
			JOIN pg_class table_row ON table_row.oid = constraint_row.conrelid
			JOIN pg_namespace table_namespace ON table_namespace.oid = table_row.relnamespace
			WHERE table_namespace.nspname = current_schema()
			  AND constraint_row.contype = 'c'
			  AND constraint_row.conname IN (
				'ck_examination_revisions_schema_version',
				'ck_examination_revisions_change_reason',
				'ck_examination_revision_items_reference_range'
			  )
		`).Scan(&checkDefinitions).Error)
		require.Len(t, checkDefinitions, 3)
		joinedChecks := strings.Join(checkDefinitions, " ")
		assert.Contains(t, joinedChecks, "schema_version = 1")
		assert.Contains(t, joinedChecks, "change_reason IS NULL")
		assert.Contains(t, joinedChecks, "btrim(change_reason)")
		assert.Contains(t, joinedChecks, "ref_min <= ref_max")

		type policyMetadata struct {
			TableName       string
			RLSEnabled      bool
			RLSForced       bool
			PolicyName      string
			UsingExpression string
			CheckExpression string
		}
		var policies []policyMetadata
		require.NoError(t, db.Raw(`
			SELECT
				table_row.relname AS table_name,
				table_row.relrowsecurity AS rls_enabled,
				table_row.relforcerowsecurity AS rls_forced,
				policy_row.polname AS policy_name,
				pg_get_expr(policy_row.polqual, policy_row.polrelid) AS using_expression,
				pg_get_expr(policy_row.polwithcheck, policy_row.polrelid) AS check_expression
			FROM pg_class table_row
			JOIN pg_namespace table_namespace ON table_namespace.oid = table_row.relnamespace
			JOIN pg_policy policy_row ON policy_row.polrelid = table_row.oid
			WHERE table_namespace.nspname = current_schema()
			  AND table_row.relname IN ('examination_revisions', 'examination_revision_items')
			ORDER BY table_row.relname
		`).Scan(&policies).Error)
		require.Len(t, policies, 2)
		expectedPolicies := map[string]string{
			"examination_revisions":      "tenant_examination_revisions_isolation",
			"examination_revision_items": "tenant_examination_revision_items_isolation",
		}
		for _, policy := range policies {
			assert.True(t, policy.RLSEnabled)
			assert.False(t, policy.RLSForced)
			assert.Equal(t, expectedPolicies[policy.TableName], policy.PolicyName)
			assert.Equal(t, "app_private.has_clinic_access(clinic_id)", policy.UsingExpression)
			assert.Equal(t, "app_private.has_clinic_access(clinic_id)", policy.CheckExpression)
		}
	})
}

func TestExaminationRevisionMigration_RLSPoliciesRejectCrossClinicAccess(t *testing.T) {
	withExaminationRevisionMigrationSchema(t, func(db *gorm.DB) {
		token := fmt.Sprintf("%d_%d", os.Getpid(), time.Now().UnixNano())
		probeRole := "examination_revision_rls_" + token
		require.NoError(t, db.Exec(fmt.Sprintf(
			"CREATE ROLE %s NOLOGIN NOSUPERUSER NOBYPASSRLS",
			probeRole,
		)).Error)
		require.NoError(t, db.Exec(fmt.Sprintf(
			"GRANT USAGE ON SCHEMA %s TO %s",
			examinationRevisionMigrationTestSchema,
			probeRole,
		)).Error)
		require.NoError(t, db.Exec(fmt.Sprintf(
			"GRANT SELECT, INSERT ON examination_revisions, examination_revision_items TO %s",
			probeRole,
		)).Error)
		require.NoError(t, db.Exec(fmt.Sprintf(
			"GRANT SELECT, UPDATE ON exams TO %s",
			probeRole,
		)).Error)

		var roleMetadata struct {
			Superuser bool
			BypassRLS bool
		}
		require.NoError(t, db.Raw(`
			SELECT rolsuper AS superuser, rolbypassrls AS bypass_rls
			FROM pg_roles
			WHERE rolname = ?
		`, probeRole).Scan(&roleMetadata).Error)
		assert.False(t, roleMetadata.Superuser)
		assert.False(t, roleMetadata.BypassRLS)

		require.NoError(t, db.Exec(`
			INSERT INTO exams (id, clinic_id, status) VALUES
				(11, 2, 'completed'),
				(20, 1, 'completed'),
				(21, 2, 'completed')
		`).Error)
		require.NoError(t, db.Exec(`
			INSERT INTO examination_revisions (
				id, clinic_id, examination_id, version, kind, status,
				exam_type_id, actor_id, date, display_snapshot, change_reason
			) VALUES (
				11, 2, 11, 1, 'official', 'confirmed', 200, 84, DATE '2026-08-03',
				jsonb_build_object(
					'medical_record_no', '', 'pet_name', '',
					'medical_record_owner_name', '', 'pet_owner_name', '',
					'species_name', '', 'exam_type_name', 'type-b', 'doctor_name', ''
				),
				'initial_confirmation'
			)
		`).Error)
		require.NoError(t, db.Exec(`
			INSERT INTO examination_revision_items (
				id, clinic_id, examination_id, version, name,
				is_assessed, is_abnormal, status
			) VALUES (11, 2, 11, 1, 'seed-b', FALSE, FALSE, 'normal')
		`).Error)
		require.NoError(t, db.Exec(`
			UPDATE exams
			SET status = 'confirmed', current_revision_version = 1
			WHERE id = 11 AND clinic_id = 2
		`).Error)
		require.NoError(t, insertExaminationRevisionVersionForClinic(db, 12, 1, 10, 2))
		require.NoError(t, insertExaminationRevisionVersionForClinic(db, 13, 2, 11, 2))

		roleCase := 0
		runAsClinicOne := func(run func(*gorm.DB) error) error {
			roleCase++
			savepoint := fmt.Sprintf("examination_revision_rls_case_%d", roleCase)
			if err := db.Exec("SAVEPOINT " + savepoint).Error; err != nil {
				return err
			}
			setupStatements := []string{
				"SET LOCAL app.current_clinic_ids = '1'",
				"SET LOCAL app.bypass_rls = 'off'",
				fmt.Sprintf("SET LOCAL ROLE %s", probeRole),
			}
			for _, statement := range setupStatements {
				if err := db.Exec(statement).Error; err != nil {
					_ = db.Exec("ROLLBACK TO SAVEPOINT " + savepoint).Error
					return err
				}
			}
			runErr := run(db)
			rollbackErr := db.Exec("ROLLBACK TO SAVEPOINT " + savepoint).Error
			if runErr != nil {
				return runErr
			}
			return rollbackErr
		}

		t.Run("clinic one sees only clinic one revision data", func(t *testing.T) {
			for _, tableName := range []string{"examination_revisions", "examination_revision_items"} {
				t.Run(tableName, func(t *testing.T) {
					var clinics []int64
					err := runAsClinicOne(func(roleDB *gorm.DB) error {
						return roleDB.Raw(
							"SELECT DISTINCT clinic_id FROM " + tableName + " ORDER BY clinic_id",
						).Scan(&clinics).Error
					})
					require.NoError(t, err)
					assert.Equal(t, []int64{1}, clinics)
				})
			}
		})

		tests := []struct {
			name     string
			wantCode string
			run      func(*gorm.DB) error
		}{
			{
				name: "same clinic revision insert is allowed",
				run: func(roleDB *gorm.DB) error {
					return roleDB.Exec(`
						INSERT INTO examination_revisions (
							id, clinic_id, examination_id, version, kind, status,
							exam_type_id, actor_id, date, display_snapshot, change_reason
						) VALUES (
							20, 1, 20, 1, 'working', 'completed', 100, 42, DATE '2026-08-03',
							jsonb_build_object(
								'medical_record_no', '', 'pet_name', '',
								'medical_record_owner_name', '', 'pet_owner_name', '',
								'species_name', '', 'exam_type_name', 'type-a', 'doctor_name', ''
							),
							'same-clinic'
						)
					`).Error
				},
			},
			{
				name:     "cross clinic revision insert is rejected",
				wantCode: "42501",
				run: func(roleDB *gorm.DB) error {
					return roleDB.Exec(`
						INSERT INTO examination_revisions (
							id, clinic_id, examination_id, version, kind, status,
							exam_type_id, actor_id, date, display_snapshot, change_reason
						) VALUES (
							21, 2, 21, 1, 'working', 'completed', 200, 84, DATE '2026-08-03',
							jsonb_build_object(
								'medical_record_no', '', 'pet_name', '',
								'medical_record_owner_name', '', 'pet_owner_name', '',
								'species_name', '', 'exam_type_name', 'type-b', 'doctor_name', ''
							),
							'cross-clinic'
						)
					`).Error
				},
			},
			{
				name: "same clinic revision item insert is allowed",
				run: func(roleDB *gorm.DB) error {
					return roleDB.Exec(`
						INSERT INTO examination_revision_items (
							id, clinic_id, examination_id, version, name,
							is_assessed, is_abnormal, status
						) VALUES (20, 1, 10, 2, 'same-clinic', FALSE, FALSE, 'normal')
					`).Error
				},
			},
			{
				name:     "cross clinic revision item insert is rejected",
				wantCode: "42501",
				run: func(roleDB *gorm.DB) error {
					return roleDB.Exec(`
						INSERT INTO examination_revision_items (
							id, clinic_id, examination_id, version, name,
							is_assessed, is_abnormal, status
						) VALUES (21, 2, 11, 2, 'cross-clinic', FALSE, FALSE, 'normal')
					`).Error
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := runAsClinicOne(tt.run)
				if tt.wantCode == "" {
					require.NoError(t, err)
					return
				}
				require.Error(t, err)
				assertPostgresCode(t, err, tt.wantCode)
			})
		}
	})
}

type examinationRevisionCrossClinicFixture struct {
	db         *gorm.DB
	clinicA    uint64
	exam       *model.Examination
	petA       *model.Pet
	recordA    *model.MedicalRecord
	ownerB     *model.Owner
	petB       *model.Pet
	recordB    *model.MedicalRecord
	doctorB    *model.Staff
	actorAID   uint64
	actorBID   uint64
	examTypeB  *model.ExaminationType
	fieldB     *model.ExamTypeField
	service    ExaminationService
	auditCalls *int
}

func TestExaminationRevision_ConfirmRejectsCrossClinicRelationsWithoutWrites(t *testing.T) {
	tests := []struct {
		name    string
		pollute func(*testing.T, *examinationRevisionCrossClinicFixture) uint64
	}{
		{
			name: "medical record belongs to another clinic",
			pollute: func(t *testing.T, fixture *examinationRevisionCrossClinicFixture) uint64 {
				require.NoError(t, fixture.db.Model(fixture.exam).
					Update("medical_record_id", fixture.recordB.ID).Error)
				return fixture.actorAID
			},
		},
		{
			name: "pet belongs to another clinic",
			pollute: func(t *testing.T, fixture *examinationRevisionCrossClinicFixture) uint64 {
				require.NoError(t, fixture.db.Model(fixture.exam).
					Update("pet_id", fixture.petB.ID).Error)
				return fixture.actorAID
			},
		},
		{
			name: "current pet owner belongs to another clinic",
			pollute: func(t *testing.T, fixture *examinationRevisionCrossClinicFixture) uint64 {
				require.NoError(t, fixture.db.Model(fixture.petA).
					Update("owner_id", fixture.ownerB.ID).Error)
				return fixture.actorAID
			},
		},
		{
			name: "medical record owner belongs to another clinic",
			pollute: func(t *testing.T, fixture *examinationRevisionCrossClinicFixture) uint64 {
				require.NoError(t, fixture.db.Model(fixture.recordA).
					Update("owner_id", fixture.ownerB.ID).Error)
				return fixture.actorAID
			},
		},
		{
			name: "doctor lacks current clinic assignment",
			pollute: func(t *testing.T, fixture *examinationRevisionCrossClinicFixture) uint64 {
				require.NoError(t, fixture.db.Model(fixture.exam).
					Update("doctor_id", fixture.doctorB.ID).Error)
				return fixture.actorAID
			},
		},
		{
			name: "exam type belongs to another clinic",
			pollute: func(t *testing.T, fixture *examinationRevisionCrossClinicFixture) uint64 {
				require.NoError(t, fixture.db.Model(fixture.exam).
					Update("exam_type_id", fixture.examTypeB.ID).Error)
				return fixture.actorAID
			},
		},
		{
			name: "exam result field belongs to another clinic",
			pollute: func(t *testing.T, fixture *examinationRevisionCrossClinicFixture) uint64 {
				require.NoError(t, fixture.db.Model(&model.ExamResult{}).
					Where("exam_id = ?", fixture.exam.ID).
					Update("exam_type_field_id", fixture.fieldB.ID).Error)
				return fixture.actorAID
			},
		},
		{
			name: "actor lacks current clinic assignment",
			pollute: func(_ *testing.T, fixture *examinationRevisionCrossClinicFixture) uint64 {
				return fixture.actorBID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupExaminationRevisionCrossClinicFixture(t)
			actorID := tt.pollute(t, fixture)
			confirmed := model.ExaminationStatusConfirmed

			got, err := fixture.service.Update(
				context.Background(),
				fixture.clinicA,
				fixture.exam.ID,
				UpdateExaminationInput{Status: &confirmed, ActorID: &actorID},
			)

			require.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(t, *fixture.auditCalls)
			var persisted model.Examination
			require.NoError(t, fixture.db.
				Where("id = ? AND clinic_id = ?", fixture.exam.ID, fixture.clinicA).
				First(&persisted).Error)
			assert.Equal(t, model.ExaminationStatusCompleted, persisted.Status)
			assert.Nil(t, persisted.CurrentRevisionVersion)
			assertExaminationRevisionRows(t, fixture.db, fixture.clinicA, fixture.exam.ID, 0, 0)
			var mutableItemCount int64
			require.NoError(t, fixture.db.Model(&model.ExamResult{}).
				Where("exam_id = ?", fixture.exam.ID).
				Count(&mutableItemCount).Error)
			assert.Equal(t, int64(1), mutableItemCount)
		})
	}
}

func setupExaminationRevisionCrossClinicFixture(t *testing.T) *examinationRevisionCrossClinicFixture {
	t.Helper()
	db := setupExaminationTestDB(t)
	const clinicA, clinicB = uint64(1), uint64(2)
	ownerA := makeTestOwner(t, db, clinicA, "tenant-a-owner")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "tenant-a-pet")
	recordA := makeClinicalRelationRecord(t, db, clinicA, ownerA.ID, petA.ID, "tenant-a-record")
	ownerB := makeTestOwner(t, db, clinicB, "tenant-b-owner")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "tenant-b-pet")
	recordB := makeClinicalRelationRecord(t, db, clinicB, ownerB.ID, petB.ID, "tenant-b-record")
	doctorA := makeDoctor(t, db, clinicA, "tenant-a-doctor")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: doctorA.ID, ClinicID: clinicA,
	}).Error)
	doctorB := makeDoctor(t, db, clinicB, "tenant-b-doctor")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: doctorB.ID, ClinicID: clinicB,
	}).Error)
	actorAID := makeExaminationActor(t, db, clinicA, "tenant-a-actor")
	actorBID := makeExaminationActor(t, db, clinicB, "tenant-b-actor")
	examTypeA := makeExamTypeMaster(t, db, clinicA, "tenant-a-type")
	fieldA := &model.ExamTypeField{
		ClinicID: clinicA, ExamTypeID: examTypeA.ID, Name: "tenant-a-field",
	}
	require.NoError(t, db.Create(fieldA).Error)
	examTypeB := makeExamTypeMaster(t, db, clinicB, "tenant-b-type")
	fieldB := &model.ExamTypeField{
		ClinicID: clinicB, ExamTypeID: examTypeB.ID, Name: "tenant-b-field",
	}
	require.NoError(t, db.Create(fieldB).Error)
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, MedicalRecordID: &recordA.ID, PetID: &petA.ID,
		ExamTypeID: examTypeA.ID, DoctorID: &doctorA.ID,
		Date:   time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
		Status: model.ExaminationStatusCompleted,
	})
	require.NoError(t, db.Create(&model.ExamResult{
		ExamID: exam.ID, ExamTypeItemID: &fieldA.ID, Name: "tenant-a-result",
		Status: model.ExaminationResultStatusNormal,
	}).Error)
	auditCalls := 0
	repository := NewExaminationRepository(db)
	service := NewExaminationService(
		repository,
		NewMedicalRecordRepository(db),
		NewExamTypeRepository(db),
		&mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
			auditCalls++
			return nil
		}},
		persistence.NewTransactor(db),
		reservation.NewReservationRepository(db),
	)
	return &examinationRevisionCrossClinicFixture{
		db: db, clinicA: clinicA, exam: exam, petA: petA, recordA: recordA,
		ownerB: ownerB, petB: petB, recordB: recordB, doctorB: doctorB,
		actorAID: actorAID, actorBID: actorBID, examTypeB: examTypeB, fieldB: fieldB,
		service: service, auditCalls: &auditCalls,
	}
}

func insertExaminationRevisionVersion(db *gorm.DB, version uint64) error {
	return insertExaminationRevisionVersionForClinic(db, version, 1, 10, version)
}

func insertExaminationRevisionVersionForClinic(
	db *gorm.DB,
	id, clinicID, examinationID, version uint64,
) error {
	return db.Exec(`
		INSERT INTO examination_revisions (
			id, clinic_id, examination_id, version, kind, status,
			exam_type_id, actor_id, date, display_snapshot, change_reason
		)
		SELECT ?, clinic_id, examination_id, ?, 'working', 'completed',
			exam_type_id, actor_id, date, display_snapshot, 'next-version'
		FROM examination_revisions
		WHERE clinic_id = ? AND examination_id = ? AND version = 1
	`, id, version, clinicID, examinationID).Error
}

func installExaminationRevisionRLSFunctions(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`CREATE SCHEMA IF NOT EXISTS app_private`,
		`CREATE OR REPLACE FUNCTION app_private.current_clinic_ids()
			RETURNS bigint[] LANGUAGE sql STABLE AS $$
				SELECT COALESCE(
					string_to_array(NULLIF(current_setting('app.current_clinic_ids', true), ''), ',')::bigint[],
					ARRAY[]::bigint[]
				);
			$$`,
		`CREATE OR REPLACE FUNCTION app_private.bypass_rls()
			RETURNS boolean LANGUAGE sql STABLE AS $$
				SELECT COALESCE(NULLIF(current_setting('app.bypass_rls', true), '')::boolean, false);
			$$`,
		`CREATE OR REPLACE FUNCTION app_private.has_clinic_access(row_clinic_id bigint)
			RETURNS boolean LANGUAGE sql STABLE AS $$
				SELECT app_private.bypass_rls()
					OR row_clinic_id = ANY(app_private.current_clinic_ids());
			$$`,
		`CREATE OR REPLACE FUNCTION app_private.apply_rls_policy(
			target_table regclass,
			policy_name text,
			using_expr text,
			check_expr text
		)
		RETURNS void LANGUAGE plpgsql AS $$
		BEGIN
			EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', target_table);
			EXECUTE format('DROP POLICY IF EXISTS %I ON %s', policy_name, target_table);
			EXECUTE format(
				'CREATE POLICY %I ON %s FOR ALL USING (%s) WITH CHECK (%s)',
				policy_name,
				target_table,
				using_expr,
				check_expr
			);
		END;
		$$`,
		`GRANT USAGE ON SCHEMA app_private TO PUBLIC`,
		`GRANT EXECUTE ON FUNCTION app_private.current_clinic_ids() TO PUBLIC`,
		`GRANT EXECUTE ON FUNCTION app_private.bypass_rls() TO PUBLIC`,
		`GRANT EXECUTE ON FUNCTION app_private.has_clinic_access(bigint) TO PUBLIC`,
		`REVOKE ALL ON FUNCTION app_private.apply_rls_policy(regclass, text, text, text) FROM PUBLIC`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}
}
