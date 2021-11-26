#!/bin/sh

set -e

docker build -f bundle-0.3.0.Dockerfile . -t quay.io/lukas_hinterreiter/luhi:olm_manual

docker push quay.io/lukas_hinterreiter/luhi:olm_manual

opm index add --container-tool docker --bundles quay.io/lukas_hinterreiter/luhi:olm_manual --tag quay.io/lukas_hinterreiter/luhi:olm_manual_index

docker push quay.io/lukas_hinterreiter/luhi:olm_manual_index

oc apply -f ./catalogsource.yaml

sleep 30

oc apply -f ./subscription.yaml
