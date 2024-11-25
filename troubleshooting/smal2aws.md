# saml2aws

## URL empty in idp account

### 문제

```log
saml2aws --verbose list-roles
DEBU[0000] Running                                       command=list-roles
URL empty in idp account
github.com/versent/saml2aws/v2/pkg/cfg.(*IDPAccount).Validate
 github.com/versent/saml2aws/v2/pkg/cfg/cfg.go:118
github.com/versent/saml2aws/v2/cmd/saml2aws/commands.buildIdpAccount
 github.com/versent/saml2aws/v2/cmd/saml2aws/commands/login.go:167
github.com/versent/saml2aws/v2/cmd/saml2aws/commands.ListRoles
 github.com/versent/saml2aws/v2/cmd/saml2aws/commands/list_roles.go:22
main.main
 github.com/versent/saml2aws/v2/cmd/saml2aws/main.go:194
runtime.main
 runtime/proc.go:250
runtime.goexit
 runtime/asm_arm64.s:1172
Failed to validate account.
github.com/versent/saml2aws/v2/cmd/saml2aws/commands.buildIdpAccount
 github.com/versent/saml2aws/v2/cmd/saml2aws/commands/login.go:169
github.com/versent/saml2aws/v2/cmd/saml2aws/commands.ListRoles
 github.com/versent/saml2aws/v2/cmd/saml2aws/commands/list_roles.go:22
main.main
 github.com/versent/saml2aws/v2/cmd/saml2aws/main.go:194
runtime.main
 runtime/proc.go:250
runtime.goexit
 runtime/asm_arm64.s:1172
error building login details
github.com/versent/saml2aws/v2/cmd/saml2aws/commands.ListRoles
 github.com/versent/saml2aws/v2/cmd/saml2aws/commands/list_roles.go:24
main.main
 github.com/versent/saml2aws/v2/cmd/saml2aws/main.go:194
runtime.main
 runtime/proc.go:250
runtime.goexit
 runtime/asm_arm64.s:1172
```

### 원인

어떤 계정의 `list-roles`을 확인할 것인지 지정해야 하는데, 인자를 누락했다

### 해결

```shell
saml2aws --verbose --idp-account="ORGNAME-aws-dev" list-roles


DEBU[0000] Running                                       command=list-roles
Using IdP Account ORGNAME-aws-dev to access KeyCloak https://keycloak\.domain\.co/auth/realms/chai-port/protocol/saml/clients/ORGNAME-aws-dev
DEBU[0000] Get credentials                               helper=osxkeychain serverURL="https://keycloak\.domain\.co/auth/realms/chai-port/protocol/saml/clients/ORGNAME-aws-dev"
DEBU[0000] Get credentials                               helper=osxkeychain user=rody
To use saved password just hit enter.
? Username rody
? Password **********

DEBU[0010] building provider                             command=list idpAccount="account {\n  URL: https://keycloak\.domain\.co/auth/realms/chai-port/protocol/saml/clients/ORGNAME-aws-dev\n  Username: \n  Provider: KeyCloak\n  MFA: Auto\n  SkipVerify: false\n  AmazonWebservicesURN: urn:amazon:webservices\n  SessionDuration: 43200\n  Profile: ORGNAME-aws-dev\n  RoleARN: \n  Region: \n}"
DEBU[0011] HTTP Req                                      URL="https://keycloak\.domain\.co/auth/realms/chai-port/login-actions/authenticate?session_code=<SESSION_CODE>&execution=bb15c617-3dfa-4c7f-8950-2f225983a8b5&client_id=ORGNAME-aws-dev&tab_id=HUP5Ch8EzCM" http=client method=POST
DEBU[0011] HTTP Res                                      Status="200 OK" http=client
? Security Token [000000] 118525
DEBU[0019] HTTP Req                                      URL="https://keycloak\.domain\.co/auth/realms/chai-port/login-actions/authenticate?session_code=<SESSION_CODE>&execution=2211728a-e8e9-46d7-934e-903e32c02c68&client_id=ORGNAME-aws-dev&tab_id=HUP5Ch8EzCM" http=client method=POST
DEBU[0019] HTTP Res                                      Status="200 OK" http=client
DEBU[0020] HTTP Req                                      URL="https://keycloak\.domain\.co/auth/realms/chai-port/login-actions/authenticate?session_code=<SESSION_CODE>&execution=bb15c617-3dfa-4c7f-8950-2f225983a8b5&client_id=ORGNAME-aws-dev&tab_id=2ZpjpJT6Xe4" http=client method=POST
DEBU[0020] HTTP Res                                      Status="200 OK" http=client
DEBU[0020] HTTP Req                                      URL="https://keycloak\.domain\.co/auth/realms/chai-port/login-actions/authenticate?session_code=<SESSION_CODE>&execution=2211728a-e8e9-46d7-934e-903e32c02c68&client_id=ORGNAME-aws-dev&tab_id=2ZpjpJT6Xe4" http=client method=POST
DEBU[0020] HTTP Res                                      Status="200 OK" http=client

Only one role to assume. Will be automatically assumed on login
arn:aws:iam::AWS_ACCOUNT_ID:role/ORGNAME-aws-dev-core-scrum
```

## Error authenticating to IdP.: unable to locate saml response field

### 문제

```log
Using IdP Account ORGNAME-aws-prod to access KeyCloak https://keycloak\.domain\.co/auth/realms/chai-port/protocol/saml/clients/ORGNAME-aws-prod
Authenticating as rody ...
Error authenticating to IdP.: unable to locate saml response field
```

### 원인

### 해결

```log
ORGNAME-aws-dev
980708
Using IdP Account ORGNAME-aws-dev to access KeyCloak https://keycloak\.domain\.co/auth/realms/ORGNAME/protocol/saml/clients/ORGNAME-aws-dev
Authenticating as rody ...
? Security Token [000000] 980708
Selected role: arn:aws:iam::AWS_ACCOUNT_ID:role/ORGNAME-aws-dev-core-scrum
Requesting AWS credentials using SAML assertion.
{"Version":1,"AccessKeyId":"AccessKeyId","SecretAccessKey":"SecretAccessKey","SessionToken":"SessionToken","Expiration":"2023-01-31T02:06:37+09:00"}
Logged in as: arn:aws:sts::AWS_ACCOUNT_ID:assumed-role/ORGNAME-aws-dev-core-scrum/rody

Your new access key pair has been stored in the AWS configuration.
Note that it will expire at 2023-01-31 02:06:37 +0900 KST
To use this credential, call the AWS CLI with the --profile option (e.g. aws --profile ORGNAME-aws-dev ec2 describe-instances).
```
