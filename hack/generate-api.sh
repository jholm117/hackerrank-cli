#!/usr/bin/env bash
# hack/generate-api.sh — Download HackerRank OpenAPI spec and generate Go types.
# Re-run whenever the upstream spec changes.
#
# Requirements:
#   - HACKERRANK_API_TOKEN env var or token in ~/.config/hackerrank/config.yaml
#   - swagger2openapi: npm install -g swagger2openapi
#   - oapi-codegen: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SPEC_DIR="$ROOT_DIR/internal/api/generated"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Resolve token
TOKEN="${HACKERRANK_API_TOKEN:-}"
if [[ -z "$TOKEN" ]]; then
    TOKEN=$(python3 -c "
import yaml, os
path = os.path.expanduser('~/.config/hackerrank/config.yaml')
if os.path.exists(path):
    with open(path) as f:
        print(yaml.safe_load(f).get('token', ''))
" 2>/dev/null || true)
fi
if [[ -z "$TOKEN" ]]; then
    # Try macOS Keychain
    TOKEN=$(security find-generic-password -s "hackerrank-api-token" -w 2>/dev/null || true)
fi
if [[ -z "$TOKEN" ]]; then
    echo "Error: No HackerRank API token found. Set HACKERRANK_API_TOKEN or run: hr auth login" >&2
    exit 1
fi

echo "==> Downloading Swagger 2.0 spec..."
curl -sf -H "Authorization: Bearer $TOKEN" "https://www.hackerrank.com/apidoc" > "$TMP_DIR/swagger2.json"

echo "==> Converting to OpenAPI 3.0..."
swagger2openapi "$TMP_DIR/swagger2.json" -o "$TMP_DIR/openapi3.json" --patch 2>/dev/null

echo "==> Patching spec issues..."
python3 - "$TMP_DIR/openapi3.json" << 'PYTHON'
import json, re, sys

with open(sys.argv[1]) as f:
    spec = json.load(f)

# 1. Strip query params from path strings (HackerRank encodes them in the path)
new_paths = {}
for path, methods in spec['paths'].items():
    clean_path = path.split('?')[0]
    if clean_path in new_paths:
        new_paths[clean_path].update(methods)
    else:
        new_paths[clean_path] = methods
spec['paths'] = new_paths

# 2. Fix params declared as 'in: path' that aren't actually in the URL
for path, methods in spec['paths'].items():
    path_params_in_url = set(re.findall(r'\{(\w+)\}', path))
    for method, op in methods.items():
        if method not in ('get','post','put','delete','patch'):
            continue
        for param in op.get('parameters', []):
            if param.get('in') == 'path' and param.get('name') not in path_params_in_url:
                param['in'] = 'query'

# 3. Remove template/documentation paths with undeclared path params
to_remove = []
for path, methods in spec['paths'].items():
    path_params_in_url = set(re.findall(r'\{(\w+)\}', path))
    for method, op in methods.items():
        if method not in ('get','post','put','delete','patch','options'):
            continue
        declared = {p.get('name') for p in op.get('parameters', []) if p.get('in') == 'path'}
        if path_params_in_url - declared:
            to_remove.append(path)
            break
for path in to_remove:
    del spec['paths'][path]

# 4. Remove OPTIONS-only paths (documentation endpoints)
options_only = [p for p, m in spec['paths'].items()
                if not any(k in m for k in ('get','post','put','delete','patch'))]
for path in options_only:
    del spec['paths'][path]

# 5. Fix non-standard types (dateTime, datetime -> string with format date-time)
def fix_types(obj):
    if isinstance(obj, dict):
        t = obj.get('type')
        if isinstance(t, str) and t.lower() in ('datetime', 'date-time'):
            obj['type'] = 'string'
            obj['format'] = 'date-time'
        elif isinstance(t, str) and t.lower() == 'float':
            obj['type'] = 'number'
            obj['format'] = 'float'
        elif isinstance(t, list):
            type_map = {'datetime': 'string', 'date-time': 'string', 'float': 'number'}
            obj['type'] = [(type_map.get(x.lower(), x)) for x in t]
            # oapi-codegen doesn't support type arrays; take first non-null
            non_null = [x for x in obj['type'] if x != 'null']
            if non_null:
                obj['type'] = non_null[0]
            else:
                obj['type'] = 'string'
        for v in obj.values():
            fix_types(v)
    elif isinstance(obj, list):
        for item in obj:
            fix_types(item)

fix_types(spec)

with open(sys.argv[1], 'w') as f:
    json.dump(spec, f, indent=2)

print(f"   Patched: {len(spec['paths'])} paths, {len(spec.get('components',{}).get('schemas',{}))} schemas")
PYTHON

echo "==> Fixing enum values with spaces..."
python3 - "$TMP_DIR/openapi3.json" << 'PYFIX'
import json, sys

with open(sys.argv[1]) as f:
    spec = json.load(f)

# Remove enum values that contain spaces (they produce invalid Go identifiers)
def fix_enums(obj):
    if isinstance(obj, dict):
        if 'enum' in obj and isinstance(obj['enum'], list):
            # Remove enum values with spaces, special chars, or on array types
            # (array enums are items constraints that confuse codegen)
            if obj.get('type') == 'array':
                del obj['enum']
            else:
                obj['enum'] = [v for v in obj['enum']
                               if not (isinstance(v, str) and (' ' in v or v.startswith('>') or v.startswith('<')))]
                if not obj['enum']:
                    del obj['enum']
        for v in obj.values():
            fix_enums(v)
    elif isinstance(obj, list):
        for item in obj:
            fix_enums(item)

fix_enums(spec)

with open(sys.argv[1], 'w') as f:
    json.dump(spec, f, indent=2)
PYFIX

echo "==> Generating Go types..."
mkdir -p "$SPEC_DIR"

# Write oapi-codegen config to only generate types (no server/client stubs)
cat > "$TMP_DIR/codegen-config.yaml" << 'YAML'
package: generated
generate:
  models: true
  chi-server: false
  echo-server: false
  gin-server: false
  gorilla-server: false
  fiber-server: false
  iris-server: false
  strict-server: false
  client: false
  embedded-spec: false
YAML

oapi-codegen -config "$TMP_DIR/codegen-config.yaml" -o "$SPEC_DIR/types.gen.go" "$TMP_DIR/openapi3.json"

echo "==> Done. Generated: $SPEC_DIR/types.gen.go"
