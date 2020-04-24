#!/bin/bash

# Simple bash script to build basic lnd tools for all the platforms
# we support with the golang cross-compiler.
#
# Copyright (c) 2016 Company 0, LLC.
# Use of this source code is governed by the ISC
# license.

set -e

LND_VERSION_REGEX="lnd version (.+) commit"
PKG="github.com/lightningnetwork/lnd"
PACKAGE=lnd

    # Build dcrlnd to extract version.
    go build github.com/decred/dcrlnd/cmd/dcrlnd

    # Extract version command output.
    LND_VERSION_OUTPUT=`./dcrlnd --version`

    # Use a regex to isolate the version string.
    LND_VERSION_REGEX="dcrlnd version (.+) commit"
    if [[ $LND_VERSION_OUTPUT =~ $LND_VERSION_REGEX ]]; then
        # Prepend 'v' to match git tag naming scheme.
        LND_VERSION="v${BASH_REMATCH[1]}"
        echo "version: $LND_VERSION"

        # If tag contains a release candidate suffix, append this suffix to the
        # lnd reported version before we compare.
        RC_REGEX="-rc[0-9]+$"
        if [[ $TAG =~ $RC_REGEX ]]; then 
            LND_VERSION+=${BASH_REMATCH[0]}
        fi

        # Match git tag with lnd version.
        if [[ $TAG != $LND_VERSION ]]; then
            echo "dcrlnd version $LND_VERSION does not match tag $TAG"
            exit 1
        fi
    else
        echo "malformed dcrlnd version output"
        exit 1
    fi

PACKAGE=dcrlnd
MAINDIR=$PACKAGE-$TAG

mkdir -p releases/$MAINDIR

mv vendor.tar.gz releases/$MAINDIR/
rm -r vendor

PACKAGESRC="releases/$MAINDIR/$PACKAGE-source-$TAG.tar"
git archive -o $PACKAGESRC HEAD
gzip -f $PACKAGESRC > "$PACKAGESRC.gz"

cd releases/$MAINDIR

# Enable exit on error.
set -e

# If LNDBUILDSYS is set the default list is ignored. Useful to release
# for a subset of systems/architectures.
SYS=${LNDBUILDSYS:-"
        darwin-386
        darwin-amd64
        dragonfly-amd64
        freebsd-386
        freebsd-amd64
        freebsd-arm
        illumos-amd64
        linux-386
        linux-amd64
        linux-armv6
        linux-armv7
        linux-arm64
        linux-ppc64
        linux-ppc64le
        linux-mips
        linux-mipsle
        linux-mips64
        linux-mips64le
        linux-s390x
        netbsd-386
        netbsd-amd64
        netbsd-arm
        netbsd-arm64
        openbsd-386
        openbsd-amd64
        openbsd-arm
        openbsd-arm64
        solaris-amd64
        windows-386
        windows-amd64
        windows-arm
"}

# Use the first element of $GOPATH in the case where GOPATH is a list
# (something that is totally allowed).
PKG="github.com/decred/dcrlnd"
COMMIT=$(git describe --abbrev=40 --dirty)
COMMITFLAGS="-X $PKG/build.PreRelease= -X $PKG/build.BuildMetadata=release -X $PKG/build.Commit=$COMMIT"

for i in $SYS; do
    OS=$(echo $i | cut -f1 -d-)
    ARCH=$(echo $i | cut -f2 -d-)
    ARM=

    if [[ $ARCH = "armv6" ]]; then
      ARCH=arm
      ARM=6
    elif [[ $ARCH = "armv7" ]]; then
      ARCH=arm
      ARM=7
    fi

    dir="${PACKAGE}-${i}-${tag}"
    mkdir "${dir}"
    pushd "${dir}"

    echo "Building:" $OS $ARCH $ARM
    env CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH GOARM=$ARM go build -v -trimpath -ldflags "-s -w -buildid= $COMMITFLAGS" github.com/decred/dcrlnd/cmd/dcrlnd
    env CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH GOARM=$ARM go build -v -trimpath -ldflags "-s -w -buildid= $COMMITFLAGS" github.com/decred/dcrlnd/cmd/dcrlncli
    cd ..

    if [[ $os == "windows" ]]; then
      zip -r "${dir}.zip" "${dir}"
    else
      tar -cvzf "${dir}.tar.gz" "${dir}"
    fi

    rm -r "${dir}"
  done

  shasum -a 256 * >manifest-$tag.txt
}

# usage prints the usage of the whole script.
function usage() {
  red "Usage: "
  red "release.sh check-tag <version-tag>"
  red "release.sh build-release <version-tag> <build-system(s)> <build-tags> <ldflags>"
}

# Whatever sub command is passed in, we need at least 2 arguments.
if [ "$#" -lt 2 ]; then
  usage
  exit 1
fi

shasum -a 256 * > manifest-$PACKAGE-$TAG.txt
