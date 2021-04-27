#!/bin/sh
GIT_BRANCH=$(git rev-parse --symbolic-full-name --verify --quiet --abbrev-ref HEAD)

echo "$GIT_BRANCH" | grep -Eq '^v(\d+\.)?(\d+\.)?(\*|\d+)$'

if [[ -z "$GIT_REPO" ]]; then
    echo "error: git repo not defined"
    exit 1
fi

if [[ -z "$GIT_TOKEN" ]]; then
    echo "error: git token not defined"
    exit 1
fi

if [[ "$?" == "0" ]]; then
    echo "on release branch: $GIT_BRANCH"
    gh release create --repo $GIT_REPO
else 
    echo "not on release branch: $GIT_BRANCH"
    exit 1
fi
