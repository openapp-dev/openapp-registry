#!/usr/bin/env bash

set -e

# find out which template changed
modified_files=$(git show --name-only | grep -E 'publicservice-template|app-template')
for file in $modified_files; do
    # get the first two dir
    template=$(echo $file | cut -d '/' -f 1,2)
    # check whether exist docker path
    if [[ -d ${template}/docker ]]; then
        cd ${template}/docker
        bash build.sh
    fi
done
