# Configuration Catalog

A catalog file, `stash.yml`, is automatically generated after syncing a configuration file with a cloud service. This file catalogs each configuration file stored in the cloud and allows specific actions like `purge` or `edit` to be performed without the user knowing where the data is stored or how the data is encrypted.

| Field | Example | Options | Description |
|-|-|-|-|
|version|0.0.0-local|String|The version of Stash used to sync the configuration.|
|context|my-slick-app|String|The repository or app name for the stored configuration.|
|clean|true|Boolean|Delete local files after updating a cloud service.|
|files[].path| config/dev/.env| String |The local file path where the cofiguration is initially synced from and restored during a get command.||
|files[].service|secrets-manager|secrets-manager, parameter-store, s3| The cloud service where configuration is stored.|
|files[].opt.kms_key_id||Guid|The KMS key id used to encrypt the configuration. Enter alias to create a new KMS key. (default: aws/secretsmanager)|
|files[].opt.secrets|single|single, multiple| Specifies if each key/value pair should be stored in a separate Secrets Manager secret for JSON and ENV file types. |
|files[].keys|| Object{} |The cloud service keys used to get configuration.|
|files[].tags|| Object{} |Local tags used when running Stash commands to target specific configuration stored in the cloud.|