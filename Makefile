STACK = ltabus
VERSION = $(shell git describe --tags --always --dirty)

.PHONY: build deploy validate destroy browsertest

DOMAINNAME = bus.dabase.com
ACMCERTIFICATEARN = arn:aws:acm:ap-southeast-1:407461997746:certificate/87b0fd84-fb44-4782-b7eb-d9c7f8714908

deploy:
	sam build
	SAM_CLI_TELEMETRY=0 sam deploy --no-progressbar --resolve-s3 --stack-name $(STACK) \
	--parameter-overrides DomainName=$(DOMAINNAME) ACMCertificateArn=$(ACMCERTIFICATEARN) Version=$(VERSION) \
	--no-confirm-changeset --no-fail-on-empty-changeset --capabilities CAPABILITY_IAM

build-MainFunction: static/style.css static/main.js
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${ARTIFACTS_DIR}/bootstrap

validate:
	aws cloudformation validate-template --template-body file://template.yml

destroy:
	aws cloudformation delete-stack --stack-name $(STACK)

sam-tail-logs:
	sam logs --stack-name $(STACK) --tail

awsclitail:
	aws logs tail /aws/lambda/ltabus --follow

node_modules/.package-lock.json: package.json package-lock.json
	npm ci

static/style.css: static/app.css node_modules/.package-lock.json
	npx esbuild --bundle static/app.css --minify --outfile=static/main.css

static/main.js: static/app.js node_modules/.package-lock.json
	npx esbuild --bundle static/app.js --minify --outfile=static/main.js

installgin:
	go install github.com/codegangsta/gin@latest

localdev: installgin static/style.css static/main.js
	gin

browsertest: static/style.css static/main.js
	npx playwright install chromium
	npx playwright test

clean:
	rm -rf main gin-bin static/main.* node_modules test-results playwright-report
