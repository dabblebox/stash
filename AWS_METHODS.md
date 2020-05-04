# AWS Configuration Methods

Managing configuration is divided into two processes, [Configuration Management](#configuration-management) which is the process modifying and storing configuration in a cloud storage service, and [Configuration Consumption](#configuration-consumption) which is the process of ingesting the configuration from a cloud storage service into a Fargate container or Lambda function. 

## Configuration Management

When using *Stash*, management is consistent accross supported cloud storage services. 

```bash
$ stash sync config/dev/.env
```
```bash
$ stash edit -t dev
```

## Configuration Consumption

Consumption has many methods with different benefits. The pupose of this document is to discuss AWS methods for Fargate containers and Lambda functions, but some of the following methods may also apply to other technologies.

#### User and Role Policies
AWS methods require developers and the container execution role or application role to have access to the appropriate cloud storage service and KMS keys.

The following *Stash* command generates Terraform scripts specific to the `dev` configuration. This makes it easier to manage the AWS access policies through Terraform.

```bash
$ stash get -t dev -o terraform
```

### Method 1: Download Configuration File (Stash CLI)

**Supports**: ECS Fargate Containers

On start up containers can call a storage service directly using the *Stash* CLI and create a configuration file inside the container allowing the application to load it into memory.

**Requirement**: The stash.yml file used to sync the configuration must included in the Docker image where *Stash* can find it.

Dockerfile (build)
```bash
FROM golang:1.13

ARG version=0.0.0-unknown.0

# install stash
RUN curl -L -o  /usr/local/bin/stash https://github.com/dabblebox/stash/releases/download/v0.1.0-rc/stash_linux_386 && chmod +x /usr/local/bin/stash


COPY . /app/  

WORKDIR /app

RUN go build -ldflags '-X main.version='$version -o app

ENTRYPOINT [ "./docker-entrypoint.bash" ]
```

docker-entrypoint.bash (get)
```bash
#!/bin/bash

echo "Getting $CONFIG_ENV configuration."

stash get -l -t $CONFIG_ENV 1> .env

exec ./app
```

### Method 2: File Injection (Stash CLI)

On start up containers can call a storage service directly using the *Stash* CLI and inject secrets into a configuration file inside the container allowing the application to load the file containing the secrets into memory. Secret tokens can be added to a configuration file that is checked into a repository. 

The tokens are AWS Secret Manager secret names. Use double colons, `::`, to specify any field in the secret's json object.

.env (config)
```
USER=${app/dev/db::user}
PASSWORD=${app/dev/db::password}
```

Dockerfile (build)
```bash
FROM golang:1.13

ARG version=0.0.0-unknown.0

# install stash
RUN curl -L -o  /usr/local/bin/stash https://github.com/dabblebox/stash/releases/download/v0.1.0-rc/stash_linux_386 && chmod +x /usr/local/bin/stash

COPY . /app/  

WORKDIR /app

RUN go build -ldflags '-X main.version='$version -o app

ENTRYPOINT [ "./docker-entrypoint.bash" ]
```

docker-entrypoint.bash (inject)
```bash
#!/bin/bash

echo "Injecting $CONFIG_ENV configuration."

stash inject $CONFIG_ENV/.env -l -s secrets-manager 1> secrets.env

exec ./app
```

### Method 3: Environment Injection (AWS)

**Supports**: ECS Fargate Containers

Configuration and secret references to cloud services like Secrets Manager, Parameter Store, and S3 can be listed in ECS task definitions. On container start, AWS injects the configuration and/or secrets from cloud service references into the containers environment variables.

*Stash* provides commands that print the task definition JSON for stashed configuration making it easier to add to a task definition. [Secrets](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/specifying-sensitive-data.html) from AWS Secrets Manager and SSM Parameter Store along with S3 [environment files](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/taskdef-envfiles.html) are supported.

```bash
$ stash get -t dev -o ecs-task-inject-json
```

```bash
$ stash get -t dev -o ecs-task-inject-env
```

### Method 4: Direct Ingest (Stash Library)

**Supports**: ECS Fargate Containers / Lambda Functions

Code can use the *Stash* Go integration library to load configuration directly into memory.

**Requirement**: The stash.yml file used to sync the configuration must be included in the Docker image or Lambda function where *Stash* can find it.

```go
package main

import (
	"log"

	"github.com/dabblebox/stash"
	"github.com/dabblebox/stash/component/output"
)

config, err := stash.GetMap(stash.GetOptions{})
if err != nil {
  log.Fatal(err)
}

for k, v := range config {
  log.Printf("%s=%s\n", k, v)
}
```
