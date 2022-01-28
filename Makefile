.PHONY: gen-mocks
gen-mocks:
	mockery --srcpkg "github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface" --name="SecretsManagerAPI"
	mockery --srcpkg "github.com/aws/aws-sdk-go/service/ecs/ecsiface" --name="ECSAPI" 

.PHONY: deps
deps:
	go install github.com/vektra/mockery/v2@latest