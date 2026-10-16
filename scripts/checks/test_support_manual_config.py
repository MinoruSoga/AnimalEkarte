#!/usr/bin/env python3
"""Offline fixtures for the initial support rollout guard."""

import contextlib
import copy
import importlib.util
import io
import json
import pathlib
import sys
import tempfile
import unittest
from unittest import mock


SPEC = importlib.util.spec_from_file_location(
    'support_config', pathlib.Path(__file__).with_name('check-support-manual-config.py')
)
policy = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(policy)


def valid_config() -> dict:
    return {'vars': {
        'SUPPORT_HYBRID_SEARCH_ENABLED': 'false',
        'SUPPORT_EXTERNAL_DATA_APPROVED': 'false',
        'SUPPORT_LLM_MODEL': 'glm-4.5-flash',
    }}


class SupportConfigTests(unittest.TestCase):
    def test_jsonc_preserves_strings_and_accepts_comments_and_trailing_commas(self) -> None:
        strings = {
            'url': 'https://example.test/a//b?x=/*data*/',
            'escaped': 'quote " then backslash \\ and // comment markers',
            'punctuation': ',]} /* not a comment */',
        }
        fields = ',\n'.join(json.dumps(key) + ': ' + json.dumps(value)
                            for key, value in strings.items())
        text = '\ufeff/* head " */{\n' + fields + ',\n'
        text += '"nested": {"items": [1,2, /* tail */],}, // tail\r\n}\n// EOF'
        self.assertEqual(policy.decode_jsonc(text), {
            **strings, 'nested': {'items': [1, 2]},
        })

    def test_jsonc_rejects_ambiguous_and_malformed_input(self) -> None:
        cases = [
            '{"a":1,"a":2}', '{"a":1,"\\u0061":2}',
            '{"x":{"a":1,"a":2}}', '{"x":NaN}', '[Infinity]', '[-Infinity]',
            '/* unfinished', '{"x":"unfinished}', '{} {}',
            '{,}', '[,]', '{"a":,}', '[1,,]', '{"a":1,,}',
        ]
        for text in cases:
            with self.subTest(text=text), self.assertRaises(ValueError):
                policy.decode_jsonc(text)

    def test_valid_defaults_and_secret_store_bindings_do_not_mutate_input(self) -> None:
        config = valid_config()
        config['env'] = {'preview': valid_config()}
        config['secrets_store_secrets'] = [{'binding': 'SUPPORT_EMBEDDING_API_KEY'}]
        before = copy.deepcopy(config)
        policy.validate_config(config)
        self.assertEqual(config, before)

    def test_flags_require_explicit_string_false(self) -> None:
        for key in policy.DISABLED_FLAGS:
            for value in ('true', 'TRUE', '1', 1, True, False, None, ' false ', []):
                with self.subTest(key=key, value=value), self.assertRaises(ValueError):
                    config = valid_config()
                    config['vars'][key] = value
                    policy.validate_config(config)
            with self.subTest(missing=key), self.assertRaises(ValueError):
                config = valid_config()
                del config['vars'][key]
                policy.validate_config(config)

    def test_nano_and_missing_models_are_not_initial_defaults(self) -> None:
        for model in ('gpt-5-nano', '', 'glm-', None, 1, []):
            with self.subTest(model=model), self.assertRaises(ValueError):
                config = valid_config()
                config['vars']['SUPPORT_LLM_MODEL'] = model
                policy.validate_config(config)

    def test_inline_credentials_are_rejected_without_echoing_values(self) -> None:
        for key in ('SUPPORT_LLM_API_KEY', 'SUPPORT_EMBEDDING_API_KEY'):
            with self.subTest(key=key), self.assertRaises(ValueError) as error:
                config = valid_config()
                config['vars'][key] = 'unit-sentinel-value'
                policy.validate_config(config)
            self.assertNotIn('unit-sentinel-value', str(error.exception))

    def test_named_envs_cannot_implicitly_inherit_or_enable_support(self) -> None:
        invalid = [None, [], {}, {'vars': {}}, valid_config()]
        invalid[-1]['vars']['SUPPORT_EXTERNAL_DATA_APPROVED'] = 'true'
        for environment in invalid:
            with self.subTest(environment=environment), self.assertRaises(ValueError):
                config = valid_config()
                config['env'] = {'preview': environment}
                policy.validate_config(config)
        for config in ([], None, {}, {'vars': []}, {**valid_config(), 'env': []}):
            with self.subTest(config=config), self.assertRaises(ValueError):
                policy.validate_config(config)

    def test_cli_checks_both_fixed_files_and_sanitizes_failures(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            for relative in policy.CONFIG_PATHS:
                path = root / relative
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text(json.dumps(valid_config()))
            for invalid in (False, True):
                if invalid:
                    path.write_text('{"private":"unit-sentinel-value",')
                output = io.StringIO()
                with mock.patch.object(policy, 'ROOT', root), \
                        mock.patch.object(sys, 'argv', ['check-support-manual-config.py']), \
                        contextlib.redirect_stdout(output):
                    result = policy.main()
                self.assertEqual(result, int(invalid))
                record = json.loads(output.getvalue())
                self.assertEqual(record['status'], 'FAIL' if invalid else 'PASS')
                self.assertNotIn('unit-sentinel-value', output.getvalue())
                self.assertEqual(record['artifacts'],
                                 [policy.CONFIG_PATHS[-1]] if invalid else list(policy.CONFIG_PATHS))

    def test_cli_rejects_symlink_inputs(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            target = root / 'fixture.json'
            target.write_text(json.dumps(valid_config()))
            path = root / policy.CONFIG_PATHS[0]
            path.parent.mkdir(parents=True)
            path.symlink_to(target)
            with mock.patch.object(policy, 'ROOT', root), \
                    mock.patch.object(sys, 'argv', ['check-support-manual-config.py']), \
                    contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(policy.main(), 1)


if __name__ == '__main__':
    unittest.main()
