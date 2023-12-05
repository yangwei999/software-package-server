#!/bin/sh

set -euo pipefail

work_dir=$1
git_user=$2
git_email=$3
repo_name=$4
clone_url=$5
upstream_repo=$6

set +e
test -d $work_dir || mkdir -p $work_dir
set -e

cd $work_dir

if [ -d "$repo_name" ]; then
    rm -rf $repo_name
fi

git clone --depth=1 "$clone_url"

cd "$repo_name"

git config user.name "$git_user"
git config user.email "$git_email"

set +e
git config --global pack.threads 1
set -e

git remote add upstream ${upstream_repo}
