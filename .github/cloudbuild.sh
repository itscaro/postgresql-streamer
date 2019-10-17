#!/usr/bin/env bash

set -eo pipefail

TAG_NAME=$(echo "$TAG_NAME" | sed 's/refs\/tags\///g')
echo "Building reference: $TAG_NAME"
echo "$GCLOUD_AUTH" | base64 --decode > "$HOME"/gcloud.json
gcloud auth activate-service-account --key-file="$HOME"/gcloud.json
gcloud config set project "$GCLOUD_PROJECT"
gcloud builds submit --substitutions TAG_NAME="$TAG_NAME" --config cloudbuild.yaml .
