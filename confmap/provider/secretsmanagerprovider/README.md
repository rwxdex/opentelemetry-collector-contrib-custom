# Secrets Manager Provider

<!-- status autogenerated section -->
| Status        |           |
| ------------- |-----------|
| Stability     | [alpha]  |
| Distributions | [contrib] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Aprovider%2Fsecretsmanagerprovider%20&label=open&color=orange&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Aprovider%2Fsecretsmanagerprovider) [![Closed issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Aprovider%2Fsecretsmanagerprovider%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Aprovider%2Fsecretsmanagerprovider) |
| Code coverage | [![codecov](https://codecov.io/github/open-telemetry/opentelemetry-collector-contrib/graph/main/badge.svg?component=provider_secretsmanager)](https://app.codecov.io/gh/open-telemetry/opentelemetry-collector-contrib/tree/main/?components%5B0%5D=provider_secretsmanager&displayType=list) |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@atoulme](https://www.github.com/atoulme) |
| Emeritus      | [@driverpt](https://www.github.com/driverpt) |

[alpha]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#alpha
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
<!-- end autogenerated section -->

## Summary

This package provides a `confmap.Provider` implementation for Amazon Secrets
Manager (`secretsmanager`) that allows the Collector the ability to read data
stored in AWS Secrets Manager.

## How it works

- Just use the placeholders with the following pattern `${secretsmanager:<arn or name>}`
- Make sure you have the `secretsmanager:GetSecretValue` in the OTEL Collector Role
- If your secret is a json string, you can get the value for a json key using the following pattern `${secretsmanager:<arn or name>#json-key}`
- You can also specify a default value by using the following pattern `${secretsmanager:<arn or name>:-<default>}`
  - The default value is used when the ARN or name is empty or the json key is not found

Prerequisites:

- Need to set up access keys from IAM console (aws_access_key_id and aws_secret_access_key) with permission to access Amazon Secrets Manager
- For details, can take a look at https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/
