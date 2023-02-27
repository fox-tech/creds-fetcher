# creds-fetcher
Tool to authenticate using Okta OIE and AWS STS.
After authentication, credentials are stored in `.aws/credentials`

## Requirements
- Go 1.18

## Installing
When installing from repo use:
````
make install
````

## Configuration
Configuration files should be created beforehand. Accepted files include `JSON` and `TOML` files. The default locations (in order) are:

- `~/.fox-tech/config.json`
- `~/.fox-tech/config.toml`
- `[ENVIRONMENT VARIABLES]`.

An override configuration location can be provided with the `--config` flag.

Example `JSON` configuration file


    {
    "default": {
        "aws_provider_arn" : "1",
        "aws_role_arn" : "2",
        "okta_client_id" : "3",
        "okta_app_id" : "4",
        "okta_url" : "5"
    },
    "profile2":{
        "aws_provider_arn" : "6",
        "aws_role_arn" : "7",
        "okta_client_id" : "8",
        "okta_app_id" : "9",
        "okta_url" : "10"
    }
    }


Example `TOML` configuration file

    [dev]
    aws_provider_arn = "arn:aws:iam::provider"
    aws_role_arn  = "arn:aws:iam::role"
    okta_client_id = "123456"
    okta_app_id = "23423434"
    okta_url = "https:okta.com/"

    [default]
    aws_provider_arn = "arn:aws:iam::differentprovider"
    aws_role_arn  = "arn:aws:iam::anotherrole"
    okta_client_id = "123456"
    okta_app_id = "23423434"
    okta_url = "https:okta.com/"


## Usage
- Getting credentials using default settings
    ````
    creds-fetcher login
    ````
    This will generate credentials for `default` profile using configuration from a configuration file in the present directory.

- Getting credentials for specific profile
    ````
    creds-fetcher login -profile PROFILE
    ````
    This will generate credentials for `PROFILE` profile using configuration from a configuration file in the present directory.

- Getting credentials from given file
    ````
    creds-fetcher login -config PATH_TO_CONFIG
    ````
    This will generate credentials using configuration from the specified configuration file.


## License Notice

Copyright 2023 Fox Corportation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

