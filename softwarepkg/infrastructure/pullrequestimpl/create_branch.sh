#!/bin/sh

set -euo pipefail

repo=$1
branch_name=$2
sig_info_file=$3
sig_info_content=$4
new_repo_file=$5
new_repo_file_tmp=$6

new_branch() {
    cd $repo

    git checkout -- .
    git clean -fd

    git checkout master

    git fetch upstream master
    git rebase upstream/master

    git checkout -B $branch_name
}

modify() {
  echo "$sig_info_content" >> $sig_info_file

  dn=$(dirname $new_repo_file)
  if [ ! -d $dn ]; then
     mkdir -p $dn
  fi

  mv $new_repo_file_tmp $new_repo_file
}

commit() {
    git add .

    git commit -m 'apply new package'

    git push origin $branch_name -f

    git checkout master

    git branch -D $branch_name
}

new_branch

modify

commit
