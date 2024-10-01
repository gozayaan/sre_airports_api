export GOOGLE_CLOUD_PROJECT='put-project-id-here'
export TF_VAR_BUCKET_NAME='bd-airport-data' TF_VAR_PROJECT_ID=${GOOGLE_CLOUD_PROJECT}

cd ${PWD}/gcs-bucket-provision/

# initialize required providers
terraform init

# preview the execution plan
terraform plan

# double-check resources before apply
terraform apply

# inspect currently applied resource state
terraform show
