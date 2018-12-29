#!/bin/bash

exact_tag="$(git describe --tags --exact-match 2>/dev/null)"
if [[ -n "$exact_tag" && -z "$(git status --porcelain -uno)" ]]; then
    echo "${exact_tag#v}"
    exit 0
fi

tag="$(git describe --tags --abbrev=0 2>/dev/null)"
if [[ -n "$tag" ]]; then
    count="$(git rev-list --count HEAD "^${tag}")"
    tag="${tag#v}"
else
    count="$(git rev-list --count HEAD)"
    tag="0.0.0"
fi
[[ -n "$(git status --porcelain -uno)" ]] && dirty='+' || dirty=
sha="$(git rev-parse --short HEAD)"
echo "${tag}.r${count}.g${sha}${dirty}"
