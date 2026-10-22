#!/bin/sh
set -eu

# Imports are mounted read-only by make; no sources are stored in the repository.
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT
mkdir -p "$workdir/api-gateway" "$workdir/profile" "$workdir/scooter" "$workdir/rental" "$workdir/pricing"
cp -R api "$workdir/api-gateway/api"
cp -R "${PROTO_IMPORTS_DIR:-/proto-imports}" "$workdir/api-gateway/imports"
cp -R ../profile/api "$workdir/profile/api"
cp -R ../scooter/api "$workdir/scooter/api"
cp -R ../rental/api "$workdir/rental/api"
cp -R ../pricing/api "$workdir/pricing/api"
cp buf.gen.local.yaml buf.gen.external.yaml "$workdir/api-gateway/"
cat > "$workdir/api-gateway/buf.yaml" <<'YAML'
version: v2
modules:
  - path: api
  - path: imports
lint:
  use:
    - STANDARD
  except:
    - PACKAGE_DIRECTORY_MATCH
  ignore:
    - imports
YAML
for service in profile scooter rental pricing; do
cat > "$workdir/$service/buf.yaml" <<'YAML'
version: v2
modules:
  - path: api
YAML
done
service_dir=$PWD
cd "$workdir/api-gateway"
buf lint api
buf build
buf generate --template buf.gen.local.yaml
buf generate --template buf.gen.external.yaml
rm -rf "$service_dir/internal/pkg/pb"
cp -R internal/pkg/pb "$service_dir/internal/pkg/"
