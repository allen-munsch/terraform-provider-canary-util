pushd ..
go mod tidy
make build
make install
popd
rm -rf .terraform
rm .terraform.lock.hcl
terraform init
terraform plan
