# openssl

## openssl 명령어 예제

- `server.crt`: a server certificate file
- `root.crt`: a self-signed root certificate file
- All certificates are in the format of `PEM`

### View the contents of a certificate

```shell
# View the contents of a certificate:
openssl x509 -in server.crt -text -noout
```

### Check a certificate chains

```shell
# Check a certificate chains:
openssl verify -verbose -CAfile root.crt server.crt
openssl verify -verbose -CAfile <(cat interm.crt root.crt) server.crt
```

### Extract public key from private key

```shell
# Extract public key from private key
openssl rsa -pubout -in server.key -out server.pub
```

### Test connection with certificate

```shell
# Test connection with certificate.
openssl s_client -connect www.example.com:8443 -state -CAfile como.crt
```

### Save server certificate in PEM

```shell
# Save server certificate in `PEM`
openssl s_client -showcerts -connect www.example.com:443 </dev/null \
    | openssl x509 -outform PEM > example.pem.crt
```

### Save server certificate in DER

```shell
# Save server certificate in `DER`
openssl s_client -showcerts -connect www.example.com:443 </dev/null \
    | openssl x509 -outform DER > example.der.crt
```

## List certificate from pkcs12 keystore

```shell
# List certificate from `pkcs12` keystore
keytool -list -v -keystore example.p12 -storetype PKCS12 -storepass ${storepass}
```

## Extract certificate from pkcs12 keystore

```shell
# Extract certificate from pkcs12 keystore
keytool -export -keystore example.p12 -alias ${alias} -file example.crt
```

## Create a new JKS trustore

```shell
# Create a new JKS trustore
keytool -import -alias rootca -file root.crt -keystore truststore.jks -storepass changeit
```

## full steps of making root, intermediate, and server certificates

```shell
#### Let's see an example says full steps of making root, intermediate, and server certificates:
# Create a CA key pair
openssl genrsa -out root.key 8192
# Create a self-signed CA certificate
openssl req -sha256 -new -x509 -days 3650 -key root.key \
    -subj "/C=KR/L=Seoul/OU=Example/CN=Example Root CA" -out root.crt


# Or you can combined above two commands:
# openssl req -nodes -x509 -sha256 -newkey rsa:8192 -keyout "root.key" -out "root.crt" -days 3650 \
# -subj "/C=KR/L=Seoul/OU=Example/CN=Example Root CA"

# Create a server key pair
openssl genrsa -out server.key 4096

# Create a CSR
openssl req -new -sha256 -key server.key \
-subj "/C=KR/L=Seoul/OU=Example/CN=svc.example.com" -out server.crt

# Create a server certificate
openssl x509 -req -in server.key -CA root.crt -CAkey root.key -set_serial 01 -out server.crt
```

## `openssl`으로 특정 SSL/TLS 버전으로 연결 시도

```shell
# TLS 1.0 사용
openssl s_client -connect yourserver.com:443 -tls1

# TLS 1.1 사용
openssl s_client -connect yourserver.com:443 -tls1_1

# TLS 1.2 사용
openssl s_client -connect yourserver.com:443 -tls1_2

# TLS 1.3 사용
openssl s_client -connect yourserver.com:443 -tls1_3
```
