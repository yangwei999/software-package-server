#!/bin/sh

set -euo pipefail

repo_dir=$1
git_token=$2
master_branch=$3
branch_name=$4
pkg_info_file=$5
spec_url=$6
src_rpm_url=$7

new_branch() {
    git checkout -- .
    git clean -fd

    git checkout $master_branch

    git fetch origin $master_branch
    git rebase origin/$master_branch

    set +e
    git rev-parse --verify "$branch_name" 2>/dev/null
    has=$?
    set -e

    if [ $has -eq 0 ]; then
        git branch -D "$branch_name"
    fi

    git checkout -b "$branch_name"
}

modify() {
    # add pkginfo.yaml
    mv $pkg_info_file .

    # download spec file
    curl -LO "$spec_url"

    # download source rpm
    if [[ $src_rpm_url == *"gitee.com"* ]]; then
        /opt/app/download "$src_rpm_url" "${git_token}"
    else
        curl -LO "$src_rpm_url"
    fi
}

commit() {
    # get the files size more than 50MB but not in .git folder
    local large_files=$(find . -path '*/.git' -prune -o -type f -size +50M -print)
    if [ -n "${large_files}" ]; then
        git lfs install
        git lfs track --filename ${large_files}
    fi

    git add .

    git commit -m 'apply new package ci pull request'

    git push origin "$branch_name"

    git checkout $master_branch

    git branch -D "$branch_name"
}

cd $repo_dir

new_branch

modify

commit
