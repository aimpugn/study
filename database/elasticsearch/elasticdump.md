# elasticdump

- [elasticdump](#elasticdump)
    - [install](#install)
    - [usage](#usage)

## install

## usage

```shell
elasticdump \
    --httpAuthFile=./auth.ini \
    --input=https://es.some.domain.co/restored_port-dev-eks-logger-2023.09.13-000254 \
    --output=./port-dev-eks-logger-2023.09.13-000254.json \
    --limit=5000 \
    --searchBody=@./searchBody.json \
    --type=data
```

```shell
elasticdump \
    --httpAuthFile=./auth.ini \
    --input=https://es.some.domain.co/restored_port-dev-eks-logger-2023.09.13-000254 \
    --output=./port-dev-eks-logger-2023.09.13-000254.sourceonly.json \
    --limit=5000 \
    --searchBody=@./searchBody.json \
    --sourceOnly \
    --type=data
```
