#!/bin/bash

# - reference-url: the source for the comparison.
# - url: the target of the comparison.
# References:
# - https://docs.liquibase.com/commands/inspection/diff.html
liquibase \
  --classpath=./mysql-connector-j/mysql-connector-j-9.2.0.jar \
  diff \
  --output-file=example \
  --diff-types=tables,columns,indexes,views \
  --url=jdbc:mysql://localhost:3307/some_service \
  --username=dev_user \
  --password=dev \
  --referenceUrl=jdbc:mysql://localhost:3308/some_service \
  --referenceUsername=prod_user \
  --referencePassword=prod
