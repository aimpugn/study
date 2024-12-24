#!/bin/bash

set -e -u -o pipefail;

IFS=$' '
items="a b c"
for x in $items; do
    echo "$x"
done
# a
# b
# c

IFS=$'\n'
for y in $items; do
    echo "$y"
done
# a b c

for y in $(echo -e "line1\nline2\nline3"); do
    echo "$y"
done
# line1
# line2
# line3

IFS=$'\r\n'
for y in $(echo -e "line4\nline5\nline6"); do
    echo "$y"
done
# line4
# line5
# line6

echo -e "line7\nline8\nline9" | while IFS=$'\n' read -r y; do
    echo "$y"
done
# line7
# line8
# line9
