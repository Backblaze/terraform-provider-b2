######################################################################
#
# File: scripts/schema-compat-compare.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

"""Compare two `terraform providers schema -json` documents.

usage: python schema-compat-compare.py BASE.json HEAD.json

Prints BREAKING, NON-BREAKING and DEPRECATION findings. Exits 1 when there is
a BREAKING finding and 2 when the comparison itself fails. The JSON has no default or force-new fields, so those
changes are not detected (see contributing/schema-compat.md).
"""

import json
import sys


def load(path):
    with open(path) as fd:
        providers = json.load(fd).get('provider_schemas', {})
    # Merge all providers; the check builds a single provider.
    out = {'resource': {}, 'data source': {}}
    for schema in providers.values():
        out['resource'].update(schema.get('resource_schemas', {}))
        out['data source'].update(schema.get('data_source_schemas', {}))
    return out


def flatten(block, prefix=''):
    """Return ({path: attribute}, {path: block_type}) for a block, recursively."""
    attrs, blocks = {}, {}
    for name, attr in block.get('attributes', {}).items():
        attrs[prefix + name] = attr
    for name, nested in block.get('block_types', {}).items():
        blocks[prefix + name] = nested
        sub_attrs, sub_blocks = flatten(nested.get('block', {}), prefix + name + '.')
        attrs.update(sub_attrs)
        blocks.update(sub_blocks)
    return attrs, blocks


def is_required(attr):
    return bool(attr.get('required'))


def compare_attrs(label, base, head, findings):
    for path in sorted(set(base) | set(head)):
        old, new = base.get(path), head.get(path)
        if old is None:
            kind = 'BREAKING' if is_required(new) else 'NON-BREAKING'
            what = 'new required attribute' if is_required(new) else 'attribute added'
            findings.append((kind, f'{label}: {path}: {what}'))
        elif new is None:
            findings.append(('BREAKING', f'{label}: {path}: attribute removed'))
        else:
            if old.get('type') != new.get('type'):
                findings.append(
                    (
                        'BREAKING',
                        f'{label}: {path}: type changed from {json.dumps(old.get("type"))}'
                        f' to {json.dumps(new.get("type"))}',
                    )
                )
            if not is_required(old) and is_required(new):
                findings.append(('BREAKING', f'{label}: {path}: optional attribute became required'))
            if old.get('sensitive') and not new.get('sensitive'):
                findings.append(('BREAKING', f'{label}: {path}: sensitive attribute became non-sensitive'))
            if not old.get('deprecated') and new.get('deprecated'):
                findings.append(('DEPRECATION', f'{label}: {path}: attribute deprecated'))
            elif old.get('deprecated') and not new.get('deprecated'):
                findings.append(('DEPRECATION', f'{label}: {path}: attribute no longer deprecated'))


def compare_blocks(label, base, head, findings):
    for path in sorted(set(base) | set(head)):
        old, new = base.get(path), head.get(path)
        if old is None:
            required = new.get('min_items', 0) > 0
            kind = 'BREAKING' if required else 'NON-BREAKING'
            what = 'new required block' if required else 'block added'
            findings.append((kind, f'{label}: {path}: {what}'))
        elif new is None:
            findings.append(('BREAKING', f'{label}: {path}: block removed'))
        else:
            if old.get('nesting_mode') != new.get('nesting_mode'):
                findings.append(
                    (
                        'BREAKING',
                        f'{label}: {path}: nesting mode changed from {old.get("nesting_mode")}'
                        f' to {new.get("nesting_mode")}',
                    )
                )
            if new.get('min_items', 0) > old.get('min_items', 0):
                findings.append(('BREAKING', f'{label}: {path}: minimum item count increased'))
            old_max, new_max = old.get('max_items'), new.get('max_items')
            if new_max is not None and (old_max is None or new_max < old_max):
                findings.append(('BREAKING', f'{label}: {path}: maximum item count decreased'))


def compare(base, head):
    findings = []
    for kind in ('resource', 'data source'):
        for name in sorted(set(base[kind]) | set(head[kind])):
            label = f'{kind} {name}'
            if name not in head[kind]:
                findings.append(('BREAKING', f'{label}: removed'))
            elif name not in base[kind]:
                findings.append(('NON-BREAKING', f'{label}: added'))
            else:
                b_attrs, b_blocks = flatten(base[kind][name].get('block', {}))
                h_attrs, h_blocks = flatten(head[kind][name].get('block', {}))
                compare_attrs(label, b_attrs, h_attrs, findings)
                compare_blocks(label, b_blocks, h_blocks, findings)
    return findings


def main(argv):
    if len(argv) != 3:
        print('usage: python schema-compat-compare.py BASE.json HEAD.json', file=sys.stderr)
        return 2
    try:
        findings = compare(load(argv[1]), load(argv[2]))
    except Exception as error:  # a crash must not look like a breaking change (exit 1)
        print(f'ERROR: could not compare the schemas: {error!r}', file=sys.stderr)
        return 2
    if not findings:
        print('No differences.')
        return 0
    for kind in ('BREAKING', 'NON-BREAKING', 'DEPRECATION'):
        for found_kind, message in findings:
            if found_kind == kind:
                print(f'{kind}: {message}')
    return 1 if any(kind == 'BREAKING' for kind, _ in findings) else 0


if __name__ == '__main__':
    sys.exit(main(sys.argv))
