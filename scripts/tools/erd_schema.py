"""Bounded, offline PostgreSQL DDL interpreter for migration-derived ERD data.

This is not a SQL executor or a general PostgreSQL parser. Unknown structural
syntax and unreviewed DO blocks fail closed rather than produce an incomplete
schema. Function, policy and data statements are outside the table catalog.
"""
from __future__ import annotations

from dataclasses import dataclass, replace
import hashlib
from pathlib import Path
import re


class SchemaError(ValueError):
    """A migration cannot be represented safely by this interpreter."""


@dataclass(frozen=True)
class Column:
    name: str
    type_sql: str
    nullable: bool
    default_sql: str | None
    source: str


@dataclass(frozen=True)
class Constraint:
    name: str
    kind: str
    columns: tuple[str, ...]
    target_table: str | None
    target_columns: tuple[str, ...]
    on_delete: str | None
    sql: str
    source: str


@dataclass(frozen=True)
class Index:
    name: str
    columns: tuple[str, ...]
    unique: bool
    predicate: str | None
    sql: str
    source: str


@dataclass(frozen=True)
class Table:
    name: str
    columns: tuple[Column, ...] = ()
    constraints: tuple[Constraint, ...] = ()
    indexes: tuple[Index, ...] = ()


@dataclass(frozen=True)
class MigrationSource:
    name: str
    sha256: str


@dataclass(frozen=True)
class Schema:
    tables: dict[str, Table]
    migrations: tuple[MigrationSource, ...]
    reviewed_do_blocks: tuple[str, ...]


@dataclass(frozen=True)
class Token:
    text: str
    start: int
    end: int
    line: int
    document: str

    @property
    def upper(self):
        return self.text.upper()


def _sql(tokens):
    return tokens[0].document[tokens[0].start:tokens[-1].end] if tokens else ''


def _tokens(document):
    """Lex comments, quoted identifiers, strings and dollar bodies atomically."""
    pos, line = 0, 1
    while pos < len(document):
        start, first_line = pos, line
        if document[pos].isspace():
            pos += 1
        elif document.startswith('--', pos):
            end = document.find('\n', pos)
            pos = len(document) if end < 0 else end
        elif document.startswith('/*', pos):
            depth, pos = 1, pos + 2
            while depth and pos < len(document):
                if document.startswith('/*', pos):
                    depth, pos = depth + 1, pos + 2
                elif document.startswith('*/', pos):
                    depth, pos = depth - 1, pos + 2
                else:
                    pos += 1
            if depth:
                raise SchemaError(f'line {line}: unterminated block comment')
        else:
            marker = re.match(r'\$(?:[A-Za-z_][A-Za-z_0-9]*)?\$', document[pos:])
            if marker:
                delimiter = marker.group()
                end = document.find(delimiter, pos + len(delimiter))
                if end < 0:
                    raise SchemaError(f'line {line}: unterminated dollar body')
                pos = end + len(delimiter)
            elif document[pos] in "'\"":
                quote, pos = document[pos], pos + 1
                escaped = quote == "'" and start > 0 and document[start - 1] in 'eE' and (start < 2 or not document[start - 2].isalnum())
                while pos < len(document):
                    if escaped and document[pos] == '\\':
                        pos += 2
                    elif document[pos] == quote:
                        if pos + 1 < len(document) and document[pos + 1] == quote:
                            pos += 2
                        else:
                            pos += 1
                            break
                    else:
                        pos += 1
                else:
                    raise SchemaError(f'line {line}: unterminated quoted token')
            else:
                word = re.match(r'[A-Za-z_][A-Za-z_0-9$]*|[0-9]+(?:\.[0-9]+)?', document[pos:])
                pos += len(word.group()) if word else 1
            yield Token(document[start:pos], start, pos, first_line, document)
        line += document[start:pos].count('\n')


def statements(document):
    """Yield token tuples for complete top-level statements; retain line numbers."""
    current, depth = [], 0
    for token in _tokens(document):
        if token.text == '(':
            depth += 1
        elif token.text == ')':
            depth -= 1
            if depth < 0:
                raise SchemaError(f'line {token.line}: unmatched closing parenthesis')
        if token.text == ';' and depth == 0:
            if current:
                yield tuple(current)
            current = []
        else:
            current.append(token)
    if depth or current:
        raise SchemaError('unterminated statement or unmatched parenthesis')


def _split(tokens):
    start, depth = 0, 0
    for pos, token in enumerate(tokens):
        depth += (token.text == '(') - (token.text == ')')
        if token.text == ',' and depth == 0:
            yield tokens[start:pos]
            start = pos + 1
    if start < len(tokens):
        yield tokens[start:]


def _group(tokens, pos):
    if pos >= len(tokens) or tokens[pos].text != '(':
        raise SchemaError('expected parenthesized list')
    depth = 1
    for end in range(pos + 1, len(tokens)):
        depth += (tokens[end].text == '(') - (tokens[end].text == ')')
        if depth == 0:
            return tokens[pos + 1:end], end + 1
    raise SchemaError('unclosed parenthesized list')


def _ident(token):
    if token.text.startswith('"'):
        return token.text[1:-1].replace('""', '"')
    if not re.fullmatch(r'[A-Za-z_][A-Za-z_0-9$]*', token.text):
        raise SchemaError(f'invalid identifier {token.text!r}')
    return token.text.lower()


def _name(tokens, pos):
    name = _ident(tokens[pos])
    if pos + 1 < len(tokens) and tokens[pos + 1].text == '.':
        if name != 'public':
            raise SchemaError(f'non-public table namespace {name}')
        return _ident(tokens[pos + 2]), pos + 3
    return name, pos + 1


def _names(tokens):
    result = []
    for part in _split(tokens):
        if len(part) != 1:
            raise SchemaError('expected plain column name')
        result.append(_ident(part[0]))
    return tuple(result)


def _boundary(tokens, pos):
    depth = 0
    for end in range(pos, len(tokens)):
        upper = tokens[end].upper
        if depth == 0 and upper in {'NOT', 'NULL', 'DEFAULT', 'PRIMARY', 'UNIQUE', 'REFERENCES', 'CHECK', 'CONSTRAINT', 'GENERATED', 'COLLATE'}:
            return end
        depth += (tokens[end].text == '(') - (tokens[end].text == ')')
    return len(tokens)


def _object_name(table, columns, label):
    """PostgreSQL 18 makeObjectName: balance byte lengths, preserve label."""
    first, second = table.encode('utf-8'), '_'.join(columns).encode('utf-8')
    available = 63 - len(label.encode('utf-8')) - 1 - bool(columns)
    first_length, second_length = len(first), len(second)
    while first_length + second_length > available:
        if first_length > second_length:
            first_length -= 1
        else:
            second_length -= 1
    parts = [first[:first_length].decode('utf-8', 'ignore')]
    if columns:
        parts.append(second[:second_length].decode('utf-8', 'ignore'))
    return '_'.join((*parts, label))


def _implicit_name(table, columns, suffix, constraints):
    names = {c.name for c in constraints}
    name, count = _object_name(table, columns, suffix), 0
    while name in names:
        count += 1
        name = _object_name(table, columns, suffix + str(count))
    return name


def _expression_columns(tokens, column_names):
    """Find distinct column references, excluding function/schema names."""
    result = []
    for pos, token in enumerate(tokens):
        if not (token.text.startswith('"') or re.fullmatch(r'[A-Za-z_][A-Za-z_0-9$]*', token.text)):
            continue
        name = _ident(token)
        if name not in column_names or name in result:
            continue
        if pos + 1 < len(tokens) and tokens[pos + 1].text in {'(', '.'}:
            continue
        if pos > 1 and tokens[pos - 2].text == tokens[pos - 1].text == ':':
            continue
        result.append(name)
    return tuple(result)


def _constraint(tokens, table, source, inline=None, known_columns=()):
    pos, explicit = 0, None
    if tokens[pos].upper == 'CONSTRAINT':
        explicit, pos = _ident(tokens[1]), 2
    upper = tokens[pos].upper
    columns, target, targets, action = (), None, (), None
    if upper in {'PRIMARY', 'UNIQUE', 'FOREIGN'}:
        kind = {'PRIMARY': 'primary', 'UNIQUE': 'unique', 'FOREIGN': 'foreign'}[upper]
        pos += 1
        if kind in {'primary', 'foreign'}:
            if tokens[pos].upper != 'KEY':
                raise SchemaError('expected KEY')
            pos += 1
        if inline is None:
            inner, pos = _group(tokens, pos)
            columns = _names(inner)
        else:
            columns = (inline,)
    elif upper == 'REFERENCES' and inline:
        kind, columns = 'foreign', (inline,)
    elif upper in {'CHECK', 'EXCLUDE'}:
        kind = upper.lower()
        if kind == 'check':
            inner, pos = _group(tokens, pos + 1)
        else:
            pos += 1
            if tokens[pos].upper != 'USING':
                raise SchemaError('EXCLUDE requires USING')
            inner, pos = _group(tokens, pos + 2)
        existing = {c.name for c in table.columns} | set(known_columns)
        columns = _expression_columns(inner, existing)
        # Exclusion operators and predicates remain verbatim in constraint.sql.
        if kind == 'exclude':
            pos = len(tokens)
    else:
        raise SchemaError(f'unsupported constraint: {_sql(tokens)}')
    if kind == 'foreign':
        if tokens[pos].upper != 'REFERENCES':
            raise SchemaError('expected REFERENCES')
        target, pos = _name(tokens, pos + 1)
        if pos < len(tokens) and tokens[pos].text == '(':
            inner, pos = _group(tokens, pos)
            targets = _names(inner)
        while pos < len(tokens):
            if tokens[pos].upper == 'MATCH':
                if tokens[pos + 1].upper not in {'SIMPLE', 'FULL'}:
                    raise SchemaError('unsupported MATCH mode')
                pos += 2
            elif tokens[pos].upper == 'ON':
                event = tokens[pos + 1].upper
                start, pos = pos + 2, pos + 3
                if tokens[start].upper in {'SET', 'NO'}:
                    pos += 1
                if ' '.join(t.upper for t in tokens[start:pos]) not in {'RESTRICT', 'CASCADE', 'SET NULL', 'SET DEFAULT', 'NO ACTION'}:
                    raise SchemaError('unsupported foreign-key action')
                if pos < len(tokens) and tokens[pos].text == '(':
                    _, pos = _group(tokens, pos)
                if event == 'DELETE':
                    action = _sql(tokens[start:pos])
                elif event != 'UPDATE':
                    raise SchemaError('unsupported foreign-key action')
            elif tokens[pos].upper in {'DEFERRABLE', 'INITIALLY', 'NOT'}:
                break
            else:
                raise SchemaError(f'unsupported foreign-key suffix {_sql(tokens[pos:])}')
    if pos < len(tokens):
        suffix = ' '.join(t.upper for t in tokens[pos:])
        if not re.fullmatch(r'(?:NOT VALID|(?:NOT )?DEFERRABLE(?: INITIALLY (?:DEFERRED|IMMEDIATE))?)', suffix):
            raise SchemaError(f'unsupported constraint suffix {suffix}')
    suffix = {'primary': 'pkey', 'unique': 'key', 'foreign': 'fkey', 'check': 'check', 'exclude': 'excl'}[kind]
    naming_columns = () if kind == 'primary' or (kind == 'check' and len(columns) != 1) else columns
    name = explicit or _implicit_name(table.name, naming_columns, suffix, table.constraints)
    if any(c.name == name for c in table.constraints):
        raise SchemaError(f'duplicate constraint {name}')
    return Constraint(name, kind, columns, target, targets, action, _sql(tokens), source)


def _column(tokens, table, source, known_columns=()):
    name, pos = _ident(tokens[0]), 1
    end = _boundary(tokens, pos)
    type_sql = _sql(tokens[pos:end])
    if not type_sql:
        raise SchemaError(f'{name}: missing type')
    column = Column(name, type_sql, True, None, source)
    table = replace(table, columns=(*table.columns, column))
    pos = end
    while pos < len(tokens):
        upper = tokens[pos].upper
        if upper == 'NOT' and tokens[pos + 1].upper == 'NULL':
            column, pos = replace(column, nullable=False), pos + 2
        elif upper == 'NULL':
            pos += 1
        elif upper == 'DEFAULT':
            end = _boundary(tokens, pos + 1)
            # DEFAULT NULL is an expression, not the following NULL attribute.
            if end == pos + 1 and tokens[end].upper == 'NULL':
                end += 1
            if end == pos + 1:
                raise SchemaError('missing DEFAULT expression')
            column, pos = replace(column, default_sql=_sql(tokens[pos + 1:end])), end
        elif upper in {'CONSTRAINT', 'PRIMARY', 'UNIQUE', 'REFERENCES', 'CHECK'}:
            start = pos
            if upper == 'CONSTRAINT':
                pos += 2
                upper = tokens[pos].upper
            if upper in {'PRIMARY', 'UNIQUE'}:
                end = pos + (2 if upper == 'PRIMARY' else 1)
            elif upper == 'CHECK':
                _, end = _group(tokens, pos + 1)
            else:
                end = len(tokens)
                # Next inline constraint/attribute after REFERENCES is separate.
                depth = 0
                for scan in range(pos + 1, len(tokens)):
                    if depth == 0 and tokens[scan].upper in {'CONSTRAINT', 'CHECK', 'UNIQUE', 'PRIMARY', 'DEFAULT', 'REFERENCES'}:
                        end = scan
                        break
                    depth += (tokens[scan].text == '(') - (tokens[scan].text == ')')
            constraint = _constraint(tokens[start:end], table, source, name, known_columns)
            table = replace(table, constraints=(*table.constraints, constraint))
            if constraint.kind == 'primary':
                column = replace(column, nullable=False)
            pos = end
        else:
            raise SchemaError(f'unsupported column attribute {_sql(tokens[pos:])}')
    return replace(table, columns=(*table.columns[:-1], column))


def _alter(tokens, table, source):
    upper, pos = tokens[0].upper, 1
    if upper == 'ADD':
        if tokens[pos].upper == 'COLUMN':
            pos += 1
        if tokens[pos].upper == 'IF':
            if [t.upper for t in tokens[pos:pos + 3]] != ['IF', 'NOT', 'EXISTS']:
                raise SchemaError('expected IF NOT EXISTS')
            pos += 3
            if any(c.name == _ident(tokens[pos]) for c in table.columns):
                return table
        if tokens[pos].upper in {'CONSTRAINT', 'PRIMARY', 'UNIQUE', 'FOREIGN', 'CHECK', 'EXCLUDE'}:
            constraint = _constraint(tokens[pos:], table, source)
            columns = tuple(replace(c, nullable=False) if constraint.kind == 'primary' and c.name in constraint.columns else c for c in table.columns)
            return replace(table, columns=columns, constraints=(*table.constraints, constraint))
        if any(c.name == _ident(tokens[pos]) for c in table.columns):
            raise SchemaError('duplicate column')
        return _column(tokens[pos:], table, source)
    if upper == 'DROP':
        what, pos = tokens[pos].upper, pos + 1
        conditional = tokens[pos].upper == 'IF'
        if conditional:
            pos += 2
        name, pos = _ident(tokens[pos]), pos + 1
        if pos != len(tokens):
            raise SchemaError('DROP suffix unsupported (including CASCADE)')
        if what == 'CONSTRAINT':
            if not conditional and name not in {c.name for c in table.constraints}:
                raise SchemaError(f'unknown constraint {name}')
            return replace(table, constraints=tuple(c for c in table.constraints if c.name != name))
        if what == 'COLUMN':
            if not conditional and name not in {c.name for c in table.columns}:
                raise SchemaError(f'unknown column {name}')
            # PostgreSQL drops same-table dependent constraints/indexes.
            indexes = tuple(i for i in table.indexes if name not in {_ident(t) for t in _tokens(', '.join(i.columns) + ' ' + (i.predicate or '')) if re.fullmatch(r'[A-Za-z_][A-Za-z_0-9$]*|"(?:[^\"]|\"\")*"', t.text)})
            return replace(table, columns=tuple(c for c in table.columns if c.name != name), constraints=tuple(c for c in table.constraints if name not in c.columns), indexes=indexes)
        raise SchemaError(f'unsupported DROP {what}')
    if upper == 'ALTER':
        if tokens[pos].upper == 'COLUMN':
            pos += 1
        name, pos = _ident(tokens[pos]), pos + 1
        column = next((c for c in table.columns if c.name == name), None)
        if column is None:
            raise SchemaError(f'unknown column {name}')
        suffix = ' '.join(t.upper for t in tokens[pos:])
        if tokens[pos].upper == 'TYPE':
            if any(t.upper in {'USING', 'COLLATE'} for t in tokens[pos + 1:]):
                raise SchemaError('ALTER TYPE USING/COLLATE unsupported')
            column = replace(column, type_sql=_sql(tokens[pos + 1:]), source=source)
        elif suffix in {'SET NOT NULL', 'DROP NOT NULL'}:
            column = replace(column, nullable=suffix.startswith('DROP'), source=source)
        elif suffix == 'DROP DEFAULT':
            column = replace(column, default_sql=None, source=source)
        elif suffix.startswith('SET DEFAULT '):
            column = replace(column, default_sql=_sql(tokens[pos + 2:]), source=source)
        else:
            raise SchemaError(f'unsupported ALTER COLUMN {suffix}')
        return replace(table, columns=tuple(column if c.name == name else c for c in table.columns))
    raise SchemaError(f'unsupported ALTER TABLE action {_sql(tokens)}')


# Exact reviewed statement hashes. Updating one requires reviewing its effects,
# not merely accepting a new digest. Non-catalog DOs include RLS, data assertions,
# grants and checksum bookkeeping, which this catalog deliberately excludes.
_REVIEWED_DO = {
    '1f2d68c8a027d0a56679a58d695bb875d8916a8349b50fd5495ccea35839b2e8': 'rls_policy',
    '7b22dfd85b61582c63f7b7c69eaffe5d6f6c60e4f33be57395476e479c0e335e': 'data_assertion',
    '2d7574c417ab8910cc806bddfdecaca8d3a7386db3a356a50276c028f9885af4': 'data_assertion',
    '45920ac83a22c2ad218f09b590835f888ffcc40130aac19753b9d217a42e1edd': 'data_cleanup',
    '75ddac964c4f3c96c68336d52d31f6d0716f7ed89f4a3fc8bb9ac232fd2aba17': 'drop_checkup_parent',
    '229b993c292b1849d80bf76664311627d4e1743c830260813d5af110ab32462f': 'add_checkup_parent',
    'aa8a1ace3fb3ecfa46c6d0ca4c961cf1072501c0f01fde5bfe69a0eed3b26f09': 'data_assertion',
    '0d8177b1a72683892e1e2ee88f6f3ae2e147ecd10a7e72ce0afc5ac80600ca50': 'checksum',
    'bb8145f87b85e4e433021540faa2cdc794b0a57b0cbebc937c6e396f80f3ba04': 'data_assertion',
    '1ef0b2b7e415cda4a466684aa33120d6c7c70eafea515ce7ef28350349241322': 'checksum',
    '5fa63dccad7cd9c60ceebee5735eed2c41e828f0c5668ba107e5b15d18420afd': 'checksum',
    'f6382ad0b0e4c1b49a7ad03cc005aecaa0180cbaab1a9580c23bbf0aa4948942': 'data_warning',
    'a2a0d54165c328b2ce6e41ba40d202cad9168941086cc888bccaadda1d27cedb': 'data_assertion',
}


def _apply_do(tokens, tables, source):
    digest = hashlib.sha256(_sql(tokens).encode()).hexdigest()
    handler = _REVIEWED_DO.get(digest)
    if handler is None:
        raise SchemaError(f'unreviewed DO block sha256={digest}')
    if handler == 'drop_checkup_parent':
        table = tables['checkup_types']
        tables = {**tables, table.name: replace(table, constraints=tuple(c for c in table.constraints if not (c.kind == 'foreign' and c.columns == ('parent_id',))))}
    elif handler == 'add_checkup_parent':
        table = tables['checkup_types']
        if not any(c.name == 'fk_checkup_types_parent_clinic' for c in table.constraints):
            ddl = 'CONSTRAINT fk_checkup_types_parent_clinic FOREIGN KEY (parent_id, clinic_id) REFERENCES checkup_types (id, clinic_id) ON DELETE SET NULL (parent_id);'
            constraint = _constraint(next(statements(ddl)), table, source)
            tables = {**tables, table.name: replace(table, constraints=(*table.constraints, constraint))}
    return tables, f'{source} sha256={digest} {handler}'


def _create_table(tokens, tables, source):
    pos = 2
    conditional = tokens[pos].upper == 'IF'
    if conditional:
        pos += 3
    name, pos = _name(tokens, pos)
    inner, end = _group(tokens, pos)
    if end != len(tokens):
        raise SchemaError('CREATE TABLE suffix unsupported')
    if name in tables:
        if conditional:
            return tables
        raise SchemaError(f'duplicate table {name}')
    table = Table(name)
    pending = []
    parts = tuple(_split(inner))
    constraint_starts = {'CONSTRAINT', 'PRIMARY', 'UNIQUE', 'FOREIGN', 'CHECK', 'EXCLUDE'}
    known_columns = tuple(_ident(part[0]) for part in parts if part[0].upper not in constraint_starts)
    for part in parts:
        if part[0].upper in {'CONSTRAINT', 'PRIMARY', 'UNIQUE', 'FOREIGN', 'CHECK', 'EXCLUDE'}:
            pending.append(part)
        else:
            if _ident(part[0]) in {c.name for c in table.columns}:
                raise SchemaError('duplicate column')
            table = _column(part, table, f'{source.split(":")[0]}:{part[0].line}', known_columns)
    for part in pending:
        constraint = _constraint(part, table, f'{source.split(":")[0]}:{part[0].line}')
        table = replace(table, constraints=(*table.constraints, constraint))
        if constraint.kind == 'primary':
            table = replace(table, columns=tuple(replace(c, nullable=False) if c.name in constraint.columns else c for c in table.columns))
    return {**tables, name: table}


def _create_index(tokens, tables, source):
    pos, unique = 1, tokens[1].upper == 'UNIQUE'
    pos += 1 if unique else 0
    pos += 1  # INDEX
    if tokens[pos].upper == 'CONCURRENTLY':
        pos += 1
    conditional = tokens[pos].upper == 'IF'
    if conditional:
        pos += 3
    name, pos = _name(tokens, pos)
    if tokens[pos].upper != 'ON':
        raise SchemaError('expected index ON')
    table_name, pos = _name(tokens, pos + 1)
    table = tables[table_name]
    existing = [i for t in tables.values() for i in t.indexes if i.name == name]
    if existing:
        if conditional:
            return tables
        raise SchemaError(f'duplicate index {name}')
    if tokens[pos].upper == 'USING':
        pos += 2
    inner, pos = _group(tokens, pos)
    columns = tuple(_sql(part) for part in _split(inner))
    predicate = None
    if pos < len(tokens):
        if tokens[pos].upper != 'WHERE':
            raise SchemaError(f'unsupported index suffix {_sql(tokens[pos:])}')
        predicate = _sql(tokens[pos + 1:])
    index = Index(name, columns, unique, predicate, _sql(tokens), source)
    return {**tables, table_name: replace(table, indexes=(*table.indexes, index))}


def _reject_select_into(tokens):
    commands = ['']
    for token in tokens:
        if token.text == '(':
            commands.append('')
        elif token.text == ')':
            commands.pop()
        elif token.upper in {'SELECT', 'INSERT', 'UPDATE', 'DELETE'}:
            commands[-1] = token.upper
        elif token.upper == 'INTO' and commands[-1] == 'SELECT':
            raise SchemaError('SELECT INTO creates an unmodeled table')


def _apply(tokens, tables, source):
    words = tuple(t.upper for t in tokens[:4])
    if words[:2] == ('CREATE', 'TABLE'):
        return _create_table(tokens, tables, source)
    if words[:2] == ('ALTER', 'TABLE'):
        name, pos = _name(tokens, 2)
        table = tables[name]
        for part in _split(tokens[pos:]):
            if len(part) > 2 and part[0].upper == 'DROP' and part[1].upper == 'COLUMN':
                column_pos = 4 if part[2].upper == 'IF' else 2
                _check_external_dependents(tables, name, _ident(part[column_pos]))
            table = _alter(part, table, f'{source.split(":")[0]}:{part[0].line}')
        return {**tables, name: table}
    if words[:2] == ('CREATE', 'INDEX') or words[:3] == ('CREATE', 'UNIQUE', 'INDEX'):
        return _create_index(tokens, tables, source)
    if words[:2] in {('DROP', 'TABLE'), ('DROP', 'INDEX')}:
        pos = 2
        conditional = tokens[pos].upper == 'IF'
        if conditional:
            pos += 2
        for part in _split(tokens[pos:]):
            name, end = _name(part, 0)
            if end != len(part):
                raise SchemaError('DROP TABLE/INDEX suffix unsupported')
            if words[1] == 'TABLE':
                if name not in tables and not conditional:
                    raise SchemaError(f'unknown table {name}')
                _check_external_dependents(tables, name)
                tables = {key: value for key, value in tables.items() if key != name}
            else:
                if not conditional and not any(i.name == name for t in tables.values() for i in t.indexes):
                    raise SchemaError(f'unknown index {name}')
                tables = {key: replace(value, indexes=tuple(i for i in value.indexes if i.name != name)) for key, value in tables.items()}
        return tables
    # Bounded table-catalog exclusions, not a generic CREATE/ALTER bypass.
    if words[0] in {'COMMENT', 'INSERT', 'UPDATE', 'DELETE', 'BEGIN', 'COMMIT', 'SET', 'GRANT', 'REVOKE'}:
        return tables
    if words[:2] == ('WITH', 'MIGRATION_TARGETS') or words[0] == 'WITH':
        _reject_select_into(tokens)
        if any(t.upper in {'CREATE', 'ALTER', 'DROP'} for t in tokens):
            raise SchemaError('structural DDL inside WITH unsupported')
        return tables
    if words[0] == 'SELECT':
        if len(tokens) < 3 or tokens[1].text != 'app_private' or tokens[2].text != '.':
            raise SchemaError('unreviewed SELECT (potential opaque function effect)')
        if tokens[3].text != 'apply_rls_policy':
            raise SchemaError('unreviewed function call')
        return tables
    if words[:2] == ('ALTER', 'TYPE'):
        if len(tokens) < 6 or tokens[3].upper != 'ADD' or tokens[4].upper != 'VALUE':
            raise SchemaError('only ALTER TYPE ADD VALUE is outside the table catalog')
        return tables
    # Deferred constraint triggers do not add a table/column; view query
    # projections are outside this table catalog (including security_invoker).
    if words[:3] == ('CREATE', 'CONSTRAINT', 'TRIGGER') or words[:2] == ('CREATE', 'VIEW'):
        return tables
    opaque = {('CREATE', 'TYPE'), ('CREATE', 'EXTENSION'), ('CREATE', 'SCHEMA'), ('CREATE', 'TRIGGER'), ('DROP', 'TRIGGER'), ('CREATE', 'POLICY'), ('DROP', 'POLICY')}
    if words[:2] in opaque or words[:2] == ('CREATE', 'FUNCTION') or words[:4] == ('CREATE', 'OR', 'REPLACE', 'FUNCTION'):
        return tables
    raise SchemaError(f'unsupported statement {_sql(tokens)[:150]}')


def _check_external_dependents(tables, table_name, column=None):
    for table in tables.values():
        if table.name == table_name:
            continue
        for constraint in table.constraints:
            if constraint.kind == 'foreign' and constraint.target_table == table_name and (column is None or not constraint.target_columns or column in constraint.target_columns):
                raise SchemaError(f'external dependency {table.name}.{constraint.name}; CASCADE is unsupported')


def _validate(tables):
    result = {}
    for table in tables.values():
        names = {c.name for c in table.columns}
        resolved = []
        for constraint in table.constraints:
            if not set(constraint.columns) <= names:
                raise SchemaError(f'{table.name}.{constraint.name}: missing source column')
            if constraint.kind != 'foreign':
                resolved.append(constraint)
                continue
            target = tables.get(constraint.target_table)
            if target is None:
                raise SchemaError(f'{table.name}.{constraint.name}: missing target {constraint.target_table}')
            target_columns = constraint.target_columns
            if not target_columns:
                primary = next((c for c in target.constraints if c.kind == 'primary'), None)
                if primary is None:
                    raise SchemaError('implicit reference has no primary key')
                target_columns = primary.columns
            if len(target_columns) != len(constraint.columns) or not set(target_columns) <= {c.name for c in target.columns}:
                raise SchemaError(f'{table.name}.{constraint.name}: invalid target columns')
            resolved.append(replace(constraint, target_columns=target_columns))
        result[table.name] = replace(table, constraints=tuple(resolved))
    return result


def load_schema(migration_dir: Path) -> Schema:
    """Read ordered top-level migrations without connecting to or changing a DB."""
    paths = sorted(migration_dir.glob('*.sql'), key=lambda path: path.name)
    if not paths:
        raise SchemaError(f'no migrations in {migration_dir}')
    tables, migrations, reviewed = {}, [], []
    for path in paths:
        raw = path.read_bytes()
        migrations.append(MigrationSource(path.name, hashlib.sha256(raw).hexdigest()))
        try:
            for tokens in statements(raw.decode('utf-8')):
                source = f'{path.name}:{tokens[0].line}'
                try:
                    if tokens[0].upper == 'DO':
                        tables, evidence = _apply_do(tokens, tables, source)
                        reviewed.append(evidence)
                    else:
                        tables = _apply(tokens, tables, source)
                except (SchemaError, KeyError, IndexError) as error:
                    raise SchemaError(f'{source}: {error}') from error
        except (UnicodeDecodeError, SchemaError) as error:
            raise SchemaError(f'{path.name}: {error}') from error
    tables = _validate(tables)
    return Schema(dict(sorted(tables.items())), tuple(migrations), tuple(reviewed))
