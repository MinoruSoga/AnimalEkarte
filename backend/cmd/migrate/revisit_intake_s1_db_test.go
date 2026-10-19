package main

// This file deliberately uses only the standard library. It can be compiled as
// a standalone test binary and run inside an approved disposable PostgreSQL
// container with psql, without a network or an existing dependency/DB volume.
// It never creates Docker resources, applies migrations, or selects a normal DB.
// Run only after applying the reviewed 001 to a NEW empty synthetic database.

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

const revisitFixtureSQL = `
CREATE SCHEMA emr460_test;
INSERT INTO companies(id,name) VALUES (100001,'SYNTHETIC EMR460');
INSERT INTO clinics(id,company_id,name) VALUES (100001,100001,'SYNTHETIC A'),(100002,100001,'SYNTHETIC B');
INSERT INTO owners(id,clinic_id,name) VALUES (100001,100001,'SYNTHETIC A'),(100002,100002,'SYNTHETIC B');
INSERT INTO animal_species(id,name) VALUES (100001,'SYNTHETIC species');
INSERT INTO pets(id,clinic_id,owner_id,name,animal_species_id) VALUES
 (100001,100001,100001,'SYNTHETIC A',100001),(100002,100001,100001,'SYNTHETIC A2',100001),
 (100003,100002,100002,'SYNTHETIC B',100001);
INSERT INTO line_customers(id,clinic_id,line_user_id,owner_id) VALUES
 (100001,100001,'SYNTHETIC-LINE-A',100001),(100002,100002,'SYNTHETIC-LINE-B',100002);
INSERT INTO reservation_types(id,clinic_id,name) VALUES (100001,100001,'SYNTHETIC A'),(100002,100002,'SYNTHETIC B');
CREATE FUNCTION emr460_test.assert_true(value boolean, label text) RETURNS void LANGUAGE plpgsql AS $$
BEGIN IF value IS DISTINCT FROM true THEN RAISE EXCEPTION 'assertion failed: %',label; END IF; END; $$;
CREATE FUNCTION emr460_test.expect_state(statement text, expected text) RETURNS void LANGUAGE plpgsql AS $$
DECLARE actual text;
BEGIN
 BEGIN EXECUTE statement;
 EXCEPTION WHEN OTHERS THEN
  GET STACKED DIAGNOSTICS actual = RETURNED_SQLSTATE;
  IF actual <> expected THEN RAISE EXCEPTION 'expected SQLSTATE %, got %',expected,actual; END IF;
  RETURN;
 END;
 RAISE EXCEPTION 'statement unexpectedly succeeded, expected %',expected;
END; $$;
CREATE FUNCTION emr460_test.appointment(clinic bigint DEFAULT 100001) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE a bigint; owner_ref bigint; pet_ref bigint; type_ref bigint; customer_ref bigint;
BEGIN
 IF clinic=100001 THEN owner_ref:=100001; pet_ref:=100001; type_ref:=100001; customer_ref:=100001;
 ELSE owner_ref:=100002; pet_ref:=100003; type_ref:=100002; customer_ref:=100002; END IF;
 INSERT INTO appointments(clinic_id,owner_id,pet_id,line_customer_id,reservation_type_id,start_time,end_time,status)
 VALUES (clinic,owner_ref,pet_ref,customer_ref,type_ref,'2030-01-01T09:00:00Z','2030-01-01T09:30:00Z','confirmed') RETURNING id INTO a;
 RETURN a;
END; $$;
-- Test helpers are not production submission/auth/replay adapters.
CREATE FUNCTION emr460_test.target(a bigint, generation bigint DEFAULT 1) RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE target_ref bigint;
BEGIN
 INSERT INTO revisit_intake_targets(clinic_id,appointment_id,pet_id,target_generation)
 SELECT clinic_id,id,pet_id,revisit_intake_generation FROM appointments
 WHERE id=a AND revisit_intake_generation=generation ON CONFLICT DO NOTHING;
 SELECT id INTO STRICT target_ref FROM revisit_intake_targets WHERE appointment_id=a AND target_generation=generation;
 RETURN target_ref;
END; $$;
CREATE FUNCTION emr460_test.claim(customer bigint, clinic bigint, operation uuid) RETURNS void LANGUAGE sql AS $$
 INSERT INTO revisit_intake_submit_receipts(clinic_id,author_customer_id,operation_key,author_subject_hash,request_hash)
 VALUES (clinic,customer,operation,decode(repeat('ab',32),'hex'),decode(repeat('cd',32),'hex')) ON CONFLICT DO NOTHING;
$$;
CREATE FUNCTION emr460_test.submit(a bigint, revision_ref bigint DEFAULT 1, operation uuid DEFAULT gen_random_uuid(), generation bigint DEFAULT 1)
RETURNS bigint LANGUAGE plpgsql AS $$
DECLARE target_ref bigint; answer_ref bigint; appt appointments%ROWTYPE;
BEGIN
 target_ref:=emr460_test.target(a,generation);
 SELECT * INTO STRICT appt FROM appointments WHERE id=a;
 PERFORM emr460_test.claim(appt.line_customer_id,appt.clinic_id,operation);
 INSERT INTO revisit_intake_answers(clinic_id,target_id,revision,form_definition_version,visit_purpose,
 author_customer_id,author_subject_hash,author_owner_id_at_submit,operation_key)
 VALUES (appt.clinic_id,target_ref,revision_ref,1,'SYNTHETIC purpose',appt.line_customer_id,
 decode(repeat('ab',32),'hex'),appt.owner_id,operation) RETURNING id INTO answer_ref;
 RETURN answer_ref;
END; $$;
CREATE ROLE emr460_probe NOLOGIN;
GRANT SELECT,INSERT ON revisit_intake_targets,revisit_intake_answers,revisit_intake_submit_receipts TO emr460_probe;
`

func revisitPSQL(ctx context.Context, script, app string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "psql", "-X", "-qAt", "-v", "ON_ERROR_STOP=1", "-v", "VERBOSITY=verbose",
		"-h", "/var/run/postgresql", "-U", "emr460_test", "-d", os.Getenv("PGDATABASE"))
	cmd.Env = append(os.Environ(), "PGPASSFILE=/tmp/emr460-no-password-file", "PGSERVICEFILE=/dev/null", "PGCONNECT_TIMEOUT=3", "PGAPPNAME="+app)
	if script != "" {
		cmd.Stdin = strings.NewReader(script)
	}
	return cmd
}

func revisitSQL(t *testing.T, script string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := revisitPSQL(ctx, script, "emr460_assert")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("synthetic SQL failed: %v\n%s\n%s", err, out, stderr.String())
	}
	return strings.TrimSpace(string(out))
}

type revisitHeldSession struct {
	input io.WriteCloser
	lines chan string
	done  chan error
	err   bytes.Buffer
}

func revisitHold(t *testing.T, script string) *revisitHeldSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	s := &revisitHeldSession{lines: make(chan string, 16), done: make(chan error, 1)}
	cmd := revisitPSQL(ctx, "", "emr460_holder")
	cmd.Stderr = &s.err
	var err error
	s.input, err = cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			s.lines <- scanner.Text()
		}
		close(s.lines)
	}()
	go func() { s.done <- cmd.Wait() }()
	t.Cleanup(func() { _ = s.input.Close(); cancel() })
	if _, err := io.WriteString(s.input, "BEGIN;\n"+script+"\n\\echo EMR460_LOCK_HELD\n"); err != nil {
		t.Fatal(err)
	}
	for {
		select {
		case line, ok := <-s.lines:
			if !ok {
				<-s.done
				t.Fatalf("holder exited before lock marker: %s", s.err.String())
			}
			if line == "EMR460_LOCK_HELD" {
				return s
			}
		case <-ctx.Done():
			t.Fatal("holder did not reach lock marker")
		}
	}
}

func (s *revisitHeldSession) commit(t *testing.T) {
	t.Helper()
	if _, err := io.WriteString(s.input, "COMMIT;\n\\q\n"); err != nil {
		t.Fatal(err)
	}
	_ = s.input.Close()
	if err := <-s.done; err != nil {
		t.Fatalf("holder commit failed: %v: %s", err, s.err.String())
	}
}

func TestRevisitIntakeS1DisposableDatabase(t *testing.T) {
	if testing.Short() || os.Getenv("EMR460_S1_TEST_RUN_ID") == "" {
		t.Skip("requires an explicitly approved new synthetic PostgreSQL container")
	}
	run := os.Getenv("EMR460_S1_TEST_RUN_ID")
	if !regexp.MustCompile(`^[a-f0-9]{12}$`).MatchString(run) || os.Getenv("PGDATABASE") != "emr460_s1_"+run {
		t.Fatal("refusing a database outside the unique EMR460 scratch name")
	}
	marker := revisitSQL(t, "SELECT shobj_description(oid,'pg_database') FROM pg_database WHERE datname=current_database();")
	if marker != "codex-owned-emr460-synthetic:"+run {
		t.Fatalf("refusing a database without the matching scratch ownership marker: got %q", marker)
	}
	if revisitSQL(t, "SELECT (SELECT count(*) FROM companies)+(SELECT count(*) FROM clinics);") != "0" {
		t.Fatal("refusing a nonempty baseline; no seed/real data may be used")
	}
	revisitSQL(t, revisitFixtureSQL)
	cases := []struct{ name, sql string }{
		{"catalog_and_unanswered", `
SELECT emr460_test.assert_true((SELECT count(*)=141 FROM pg_tables WHERE schemaname='public'),'141 public tables');
SELECT emr460_test.assert_true(NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='medical_records' AND column_name='revisit_intake_target_id'),'no S2');
SELECT emr460_test.assert_true((SELECT count(*)=3 FROM pg_class WHERE relname IN ('revisit_intake_targets','revisit_intake_answers','revisit_intake_submit_receipts') AND relrowsecurity AND NOT relforcerowsecurity),'RLS enabled not forced');
SELECT emr460_test.assert_true((SELECT count(*)=8 FROM pg_trigger WHERE NOT tgisinternal AND tgname LIKE 'revisit_%'),'eight S1 triggers');
DO $$ DECLARE a bigint:=emr460_test.appointment(); BEGIN
 PERFORM emr460_test.assert_true((SELECT revisit_intake_generation=1 FROM appointments WHERE id=a),'default generation');
 UPDATE appointments SET status='checked_in' WHERE id=a;
 PERFORM emr460_test.assert_true(NOT EXISTS(SELECT 1 FROM revisit_intake_answers),'unanswered does not block check-in or create an answer');
END; $$;`},
		{"revision_history_and_optional_nulls", `DO $$ DECLARE a bigint:=emr460_test.appointment(); first_id bigint; second_id bigint; BEGIN
 first_id:=emr460_test.submit(a); second_id:=emr460_test.submit(a,2);
 PERFORM emr460_test.assert_true((SELECT count(*)=2 FROM revisit_intake_answers),'immutable revisions');
 PERFORM emr460_test.assert_true((SELECT onset IS NULL AND energy IS NULL AND appetite IS NULL AND current_symptoms IS NULL AND improvement IS NULL AND examination_preference IS NULL AND main_concern IS NULL FROM revisit_intake_answers WHERE id=first_id),'no invented optional values');
 PERFORM emr460_test.assert_true((SELECT count(*)=2 FROM revisit_intake_submit_receipts WHERE state='succeeded' AND result_answer_id IN(first_id,second_id) AND committed_at>=claimed_at),'atomic success receipts');
END; $$;`},
		{"reschedule_preserves_generation", `DO $$ DECLARE a bigint:=emr460_test.appointment(); ans bigint; BEGIN
 ans:=emr460_test.submit(a); UPDATE appointments SET start_time=start_time+interval '1 day',end_time=end_time+interval '1 day',notes='SYNTHETIC note' WHERE id=a;
 PERFORM emr460_test.assert_true((SELECT revisit_intake_generation=1 FROM appointments WHERE id=a),'date and notes keep target');
 PERFORM emr460_test.assert_true((SELECT count(*)=1 FROM revisit_intake_answers WHERE id=ans),'answer retained');
END; $$;`},
		{"pet_A_B_A_and_null_generation", `DO $$ DECLARE a bigint:=emr460_test.appointment(); tid bigint; BEGIN
 PERFORM emr460_test.submit(a); SELECT id INTO tid FROM revisit_intake_targets WHERE appointment_id=a;
 UPDATE appointments SET pet_id=100002 WHERE id=a; UPDATE appointments SET pet_id=100001 WHERE id=a;
 PERFORM emr460_test.assert_true((SELECT revisit_intake_generation=3 FROM appointments WHERE id=a),'A-B-A is not original target');
 PERFORM emr460_test.expect_state(format('SELECT emr460_test.submit(%s,2)',a),'23514');
 UPDATE appointments SET pet_id=NULL WHERE id=a; UPDATE appointments SET pet_id=100001 WHERE id=a;
 UPDATE appointments SET pet_id=100001 WHERE id=a;
 PERFORM emr460_test.assert_true((SELECT revisit_intake_generation=5 FROM appointments WHERE id=a),'null changes increment, same value does not');
 PERFORM emr460_test.assert_true((SELECT count(*)=1 FROM revisit_intake_answers WHERE target_id=tid),'old history retained');
 PERFORM emr460_test.expect_state(format('UPDATE appointments SET revisit_intake_generation=1 WHERE id=%s',a),'23514');
END; $$;`},
		{"reception_and_cancellation_boundaries", `DO $$ DECLARE a bigint; st text; BEGIN
 FOREACH st IN ARRAY ARRAY['checked_in','in_consultation','accounting','completed','cancelled'] LOOP
  a:=emr460_test.appointment(); PERFORM emr460_test.submit(a);
  EXECUTE format('UPDATE appointments SET status=%L WHERE id=%s',st,a);
  PERFORM emr460_test.expect_state(format('SELECT emr460_test.submit(%s,2)',a),'23514');
 END LOOP;
 a:=emr460_test.appointment(); PERFORM emr460_test.submit(a); UPDATE appointments SET deleted_at=clock_timestamp() WHERE id=a;
 PERFORM emr460_test.expect_state(format('SELECT emr460_test.submit(%s,2)',a),'23514');
 PERFORM emr460_test.assert_true((SELECT count(*)=6 FROM revisit_intake_answers),'cancel/soft-delete never erase answers');
a:=emr460_test.appointment(); UPDATE appointments SET status=NULL WHERE id=a;
 PERFORM emr460_test.expect_state(format('SELECT emr460_test.submit(%s)',a),'23514');
END; $$;`},
		{"history_and_receipt_mutation_rejected", `DO $$ DECLARE a bigint:=emr460_test.appointment(); ans bigint; tid bigint; BEGIN
 ans:=emr460_test.submit(a); SELECT target_id INTO tid FROM revisit_intake_answers WHERE id=ans;
 PERFORM emr460_test.expect_state(format('UPDATE revisit_intake_answers SET visit_purpose=''overwrite'' WHERE id=%s',ans),'23514');
 PERFORM emr460_test.expect_state(format('DELETE FROM revisit_intake_answers WHERE id=%s',ans),'23514');
 PERFORM emr460_test.expect_state(format('UPDATE revisit_intake_targets SET pet_id=100002 WHERE id=%s',tid),'23514');
 PERFORM emr460_test.expect_state(format('DELETE FROM revisit_intake_targets WHERE id=%s',tid),'23514');
 PERFORM emr460_test.expect_state('UPDATE revisit_intake_submit_receipts SET request_hash=decode(repeat(''00'',32),''hex'')','23514');
 PERFORM emr460_test.expect_state('DELETE FROM revisit_intake_submit_receipts','23514');
 PERFORM emr460_test.expect_state(format('DELETE FROM appointments WHERE id=%s',a),'23001');
END; $$;`},
		{"closed_form_checks_and_revision_gap", `DO $$ DECLARE a bigint:=emr460_test.appointment(); tid bigint; k uuid:=gen_random_uuid(); base_sql text; BEGIN
 tid:=emr460_test.target(a); PERFORM emr460_test.claim(100001,100001,k);
 base_sql:=format('INSERT INTO revisit_intake_answers(clinic_id,target_id,revision,form_definition_version,visit_purpose,energy,author_customer_id,author_subject_hash,author_owner_id_at_submit,operation_key) VALUES(100001,%s,1,1,%%L,%%L,100001,decode(repeat(''ab'',32),''hex''),100001,%L)',tid,k);
 PERFORM emr460_test.expect_state(format(base_sql,'','usual'),'23514');
 PERFORM emr460_test.expect_state(format(base_sql,'purpose','unknown'),'23514');
 PERFORM emr460_test.expect_state(format(base_sql,'purpose','usual,usual'),'23514');
 PERFORM emr460_test.expect_state(format('SELECT emr460_test.submit(%s,3)',a),'23514');
 PERFORM emr460_test.submit(a,1,k);
 PERFORM emr460_test.expect_state(format('SELECT emr460_test.submit(%s,2,%L)',a,k),'23514');
 PERFORM emr460_test.assert_true((SELECT count(*)=1 FROM revisit_intake_answers),'one effect per operation');
END; $$;`},
		{"target_clinic_actor_and_generation_mismatch", `DO $$ DECLARE a bigint:=emr460_test.appointment(); BEGIN
 PERFORM emr460_test.expect_state(format('INSERT INTO revisit_intake_targets(clinic_id,appointment_id,pet_id,target_generation) VALUES(100002,%s,100003,1)',a),'23514');
 PERFORM emr460_test.expect_state(format('INSERT INTO revisit_intake_targets(clinic_id,appointment_id,pet_id,target_generation) VALUES(100001,%s,100002,1)',a),'23514');
 PERFORM emr460_test.expect_state(format('INSERT INTO revisit_intake_targets(clinic_id,appointment_id,pet_id,target_generation) VALUES(100001,%s,100001,99)',a),'23514');
 PERFORM emr460_test.assert_true(NOT EXISTS(SELECT 1 FROM revisit_intake_targets),'no invalid target inserted');
END; $$;`},
		{"receipt_sink_failure_rolls_back_answer", `
CREATE FUNCTION emr460_test.reject_receipt() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic audit sink failure'; END; $$;
CREATE TRIGGER zz_emr460_reject_receipt BEFORE UPDATE ON revisit_intake_submit_receipts FOR EACH ROW EXECUTE FUNCTION emr460_test.reject_receipt();
DO $$ DECLARE a bigint:=emr460_test.appointment(); k uuid:=gen_random_uuid(); BEGIN
 PERFORM emr460_test.claim(100001,100001,k);
 PERFORM emr460_test.expect_state(format('SELECT emr460_test.submit(%s,1,%L)',a,k),'P0001');
 PERFORM emr460_test.assert_true(NOT EXISTS(SELECT 1 FROM revisit_intake_answers),'answer rollback');
 PERFORM emr460_test.assert_true(NOT EXISTS(SELECT 1 FROM revisit_intake_targets),'new target rollback');
 PERFORM emr460_test.assert_true((SELECT state='claimed' AND result_answer_id IS NULL FROM revisit_intake_submit_receipts WHERE operation_key=k),'claim is not success');
END; $$;`},
		{"nonowner_role_RLS_is_clinic_only", `
SELECT emr460_test.submit(emr460_test.appointment(100001)); SELECT emr460_test.submit(emr460_test.appointment(100002));
SET LOCAL ROLE emr460_probe; SET LOCAL app.current_clinic_ids='100001';
DO $$ BEGIN
 IF (SELECT count(*) FROM revisit_intake_targets)<>1 OR (SELECT count(*) FROM revisit_intake_answers)<>1 OR (SELECT count(*) FROM revisit_intake_submit_receipts)<>1 THEN RAISE EXCEPTION 'RLS clinic leak'; END IF;
END; $$;
SET LOCAL app.current_clinic_ids='';
DO $$ BEGIN IF EXISTS(SELECT 1 FROM revisit_intake_answers) THEN RAISE EXCEPTION 'missing scope leaked'; END IF; END; $$;
RESET ROLE;`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { revisitSQL(t, "BEGIN;\n"+tc.sql+"\nROLLBACK;") })
	}
	for _, state := range []string{"checked_in", "cancelled", "pet_rebind"} {
		for _, submitFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("concurrent_%s_submit_first_%t", state, submitFirst), func(t *testing.T) {
				a := revisitSQL(t, "SELECT emr460_test.appointment();")
				revisitSQL(t, "SELECT emr460_test.target("+a+");")
				transition := "UPDATE appointments SET status='" + state + "' WHERE id=" + a + ";"
				if state == "pet_rebind" {
					transition = "UPDATE appointments SET pet_id=100002 WHERE id=" + a + ";"
				}
				submit := "SELECT emr460_test.submit(" + a + ");"
				first, second := transition, submit
				if submitFirst {
					first, second = submit, transition
				}
				holder := revisitHold(t, first)
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				waiter := revisitPSQL(ctx, "BEGIN; SET LOCAL lock_timeout='10s';\n"+second+"\nCOMMIT;", "emr460_waiter")
				var output bytes.Buffer
				waiter.Stdout, waiter.Stderr = &output, &output
				if err := waiter.Start(); err != nil {
					t.Fatal(err)
				}
				done := make(chan error, 1)
				go func() { done <- waiter.Wait() }()
				deadline := time.Now().Add(8 * time.Second)
				waiting := false
				for time.Now().Before(deadline) {
					if revisitSQL(t, "SELECT count(*) FROM pg_stat_activity WHERE application_name='emr460_waiter' AND wait_event_type='Lock';") == "1" {
						waiting = true
						break
					}
				}
				if !waiting {
					t.Fatal("waiter never entered a verified PostgreSQL lock wait")
				}
				holder.commit(t)
				err := <-done
				if submitFirst && err != nil {
					t.Fatalf("transition after answer failed: %v: %s", err, output.String())
				}
				if !submitFirst && (err == nil || !strings.Contains(output.String(), "23514")) {
					t.Fatalf("stale answer was not rejected: %v: %s", err, output.String())
				}
				want := "0"
				if submitFirst {
					want = "1"
				}
				if got := revisitSQL(t, "SELECT count(*) FROM revisit_intake_answers a JOIN revisit_intake_targets t ON t.id=a.target_id AND t.clinic_id=a.clinic_id WHERE t.appointment_id="+a+";"); got != want {
					t.Fatalf("answer count=%s want %s", got, want)
				}
			})
		}
	}
}
