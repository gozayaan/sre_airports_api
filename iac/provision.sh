export GOOGLE_CLOUD_PROJECT='put-project-id-here'

cd ${PWD}/iac/

terraform init

terraform plan

terraform apply

# view storage bucket and uploaded object
