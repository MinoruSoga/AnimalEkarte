-- EMR-502: persisted claim anchor for manual.article.edit bindings.
-- Records the original global article a claim addressed, keyed by the binding.
-- Apply only through the normal, human-run migration procedure.
-- INSERT/SELECT-only: no article FK (a business-row KEY SHARE during claim is
-- out of lock order), no unique(article_id) (many operations may address the
-- same global article), no created_at/actor duplication (operation_bindings).
CREATE TABLE manual_article_operation_targets (
    clinic_id  bigint NOT NULL CHECK (clinic_id > 0),
    binding_id bigint NOT NULL CHECK (binding_id > 0),
    article_id bigint NOT NULL CHECK (article_id > 0),
    PRIMARY KEY (clinic_id, binding_id),
    FOREIGN KEY (binding_id, clinic_id) REFERENCES operation_bindings (id, clinic_id) ON DELETE RESTRICT
);
COMMENT ON TABLE manual_article_operation_targets IS 'Claim-time original article of a manual.article.edit binding. clinic_id is the authorization origin, not article ownership. Never repaired or backfilled.';
COMMENT ON COLUMN manual_article_operation_targets.clinic_id IS 'Authorization-scope clinic copied from the binding, not ownership of the global article';
COMMENT ON COLUMN manual_article_operation_targets.binding_id IS 'operation_bindings row claimed in the same transaction';
COMMENT ON COLUMN manual_article_operation_targets.article_id IS 'Global manual_articles row the claim addressed, resolved only from canonical request content';

SELECT app_private.apply_rls_policy(
    'manual_article_operation_targets'::regclass,
    'manual_article_operation_targets_clinic_scope',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)');

CREATE TRIGGER manual_article_operation_targets_immutable
    BEFORE UPDATE OR DELETE ON manual_article_operation_targets
    FOR EACH ROW EXECUTE FUNCTION app_private.contract_history_immutable();
