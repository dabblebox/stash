# Terraform

When performaing a file sync, *Stash* automatically creates the infrastructure needed to store configuration. To modify infrastructure or update security policies, *Stash* can generate the related Terraform files. Each file contains the Terraform import statements needed to take control of any existing infrastructure. 

```bash
$ stash get -t dev -o terraform
```

To update the AWS security policies for the chosen service, *Stash* creates a `{AWS_SERVICE}.auto.tfvars` file that can be used to configure the users and roles that have access to the services and KMS keys.