STACK = ltabus
PROFILE = mine
VERSION = "0.1"

.PHONY: build deploy validate destroy

DOMAINNAME = bus.dabase.com
ACMCERTIFICATEARN = arn:aws:acm:ap-southeast-1:407461997746:certificate/87b0fd84-fb44-4782-b7eb-d9c7f8714908

deploy:
	sam build
	SAM_CLI_TELEMETRY=0 AWS_PROFILE=$(PROFILE) sam deploy --resolve-s3 --stack-name $(STACK) --parameter-overrides DomainName=$(DOMAINNAME) ACMCertificateArn=$(ACMCERTIFICATEARN) --no-confirm-changeset --no-fail-on-empty-changeset --capabilities CAPABILITY_IAM

build-MainFunction:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$(VERSION)" -o ${ARTIFACTS_DIR}/bootstrap

validate:
	AWS_PROFILE=$(PROFILE) aws cloudformation validate-template --template-body file://template.yml

destroy:
	AWS_PROFILE=$(PROFILE) aws cloudformation delete-stack --stack-name $(STACK)

sam-tail-logs:
	AWS_PROFILE=$(PROFILE) sam logs --stack-name $(STACK) --tail

clean:
	rm -rf main gin-bin
