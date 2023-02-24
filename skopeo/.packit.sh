#!/usr/bin/env bash

# This script handles any custom processing of the spec file generated using the `post-upstream-clone`
# action and gets used by the fix-spec-file action in .packit.yaml.

set -eo pipefail

# Get Version from HEAD
VERSION=$(grep '^const Version' version/version.go | cut -d\" -f2 | sed -e 's/-/~/')

# Generate source tarball
git archive --prefix=skopeo-$VERSION/ -o skopeo-$VERSION.tar.gz HEAD

# RPM Spec modifications

# Update Version in spec with Version from Cargo.toml
sed -i "s/^Version:.*/Version: $VERSION/" skopeo.spec

# Update Release in spec with Packit's release envvar
sed -i "s/^Release:.*/Release: $PACKIT_RPMSPEC_RELEASE%{?dist}/" skopeo.spec

# Update Source tarball name in spec
sed -i "s/^Source:.*.tar.gz/Source: skopeo-$VERSION.tar.gz/" skopeo.spec

# Update setup macro to use the correct build dir
sed -i "s/^%setup.*/%autosetup -Sgit -n %{name}-$VERSION/" skopeo.spec
