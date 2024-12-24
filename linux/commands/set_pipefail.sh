#!/bin/bash

# https://gist.github.com/mohanpedala/1e2ff5661761d3abd0385e8223e16425#set--o-pipefail

echo "Before 'set -o pipefail'"

grep -s some-string /non/existent/file
printf "  %s%d\n" "'grep some-string /non/existent/file' exit code: " $? # 0

grep -s some-string /non/existent/file | sort
printf "  %s%d\n" "'some-string /non/existent/file | sort' exit code: " $? # 2


set -o pipefail;

echo "After 'set -o pipefail'"

grep -s some-string /non/existent/file
printf "  %s%d\n" "'grep some-string /non/existent/file' exit code: " $? # 2

grep -s some-string /non/existent/file | sort
printf "  %s%d\n" "'some-string /non/existent/file | sort' exit code: " $? # 2

