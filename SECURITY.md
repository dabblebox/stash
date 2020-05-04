# Security

*Stash* offers complete source code transparency and best practices to mitigate security vulnerabilities. Please review this tool before using to ensure that it meets any specific security requirements.

To prevent configuration from being compromised, *Stash* utilizes very few direct third-party dependencies, [go.mod](/go.mod). The only direct third-party dependency that actually handles configuration is [github.com/aws/aws-sdk-go](https://github.com/aws/aws-sdk-go) which is managed by AWS.

**Stash Security Advantages**

* Encryption is automatically applied.
* Local configuration files are discouraged.
* AWS Secrets Manager is easier to use.
* *stash.yml* files can be safely shared on any medium.
* Environment variables are easier to avoid when desired.
* Configuration is completely managed from the terminal.
* Configuration and secrets are managed together.

**Other Recommended Security Practices**

* Understand your organization's security requirements.
* Use your organization's approved secrets vaults.
* Avoid sending secrets to stdout and/or logs.
* Always use encryption when storing secrets.
* Avoid exporting secrets into the environment when possible.
* Realize many security mistakes are made by users; so, be careful!

If any security issues are found in the source code, please report and/or submit a PR for review.

Feel free to fork the repository to control the source directly.
