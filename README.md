# Set up
* create ~/.aws/config file
```
[profile default]
region = ap-northeast-1

[profile sub]
role_arn = arn:aws:iam::{SubAWSAccountId}:role/{RoleInSubAWSAccountId}
source_profile = default
mfa_serial = arn:aws:iam::{DefaultAWSAccountId}:mfa/{UserName}
region = ap-northeast-1
```
