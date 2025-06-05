INFRA:=infra
LAMBDA:='src/lambda/hello_world'
USER:=$(shell whoami)

CONTEXT=--context stage=${USER}

.PHONY: deploy
deploy: package-lambda
	cd ${INFRA}; cdk deploy ${CONTEXT}

.PHONY: package-lambda
package-lambda:
	cd ${LAMBDA}; \
	poetry export --format=requirements.txt > requirements.lambda.txt; \
	rm -rf dist; \
	poetry run pip install --target dist -r requirements.lambda.txt; \
	rsync -av --exclude='dist' --exclude='.venv' . dist/

.PHONY: run-lambda
run-lambda:
	aws lambda invoke --function-name hello-world-2 --payload '{ "request-id": "3453" }' --cli-binary-format raw-in-base64-out response.json && cat response.json; rm response.json

.PHONY: tail-lambda
tail-lambda:
	aws logs tail /aws/lambda/hello-world-2 --follow


.PHONY: test
test:
	cd test/networked; go test