#!/usr/bin/env bash

set -ex

MODNAME="gitlab.flora.loc/mills/${DAEMONNAME}"
BUILD_DIR="${GOPATH}/src/${MODNAME}"
RESULT_DIR="/artifacts"

#----------------------------------------
# get sources

cd $BUILD_DIR
git checkout ${VERSION}

#----------------------------------------
# build

cd ${BUILD_DIR}
go get ./...
go build ./...

#----------------------------------------
# create artifact

ARTIFACT_NAME="${DAEMONNAME}-${VERSION}"
ARTIFACT_DIR=${RESULT_DIR}/${ARTIFACT_NAME}
mkdir -p ${ARTIFACT_DIR}/bin

cp -v ${GOPATH}/bin/blocks-stream-receiver          ${ARTIFACT_DIR}/bin/
cp -v ${GOPATH}/bin/blocks-stream-receiver-simple   ${ARTIFACT_DIR}/bin/
cp -v ${GOPATH}/bin/ton-api                         ${ARTIFACT_DIR}/bin/
#strip ${ARTIFACT_DIR}/bin/*

cd $RESULT_DIR
tar -vczf ${ARTIFACT_NAME}-linux64.tar.gz *
echo VERSION=${VERSION} > $RESULT_DIR/VERSION.txt
