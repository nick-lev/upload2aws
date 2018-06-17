# upload2aws
The SDK looks for environment variables
credentials:
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
region settings:
AWS_REGION

If no environment variables find SDK looking for shared credential file
The file must be on the same machine on which you're running your application.
The file must be named `credentials` and located in the .aws/ folder in your home directory

[default]
aws_access_key_id = <YOUR_ACCESS_KEY_ID>
aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>

region looked at `config` file (same location as shared credential file)
[default]
region = <YOUR_REGION_STRING>

more detailed information at https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
