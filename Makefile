INFRA:=infra
LAMBDA:='src/lambda/hello_world'

.PHONY: deploy
deploy:
	cd ${INFRA}; make deploy


.PHONY: package-lambda
package-lambda:
	cd ${LAMBDA}; \
	poetry export --format=requirements.txt > requirements.lambda.txt; \
	rm -rf dist; \
	poetry run pip install --target dist -r requirements.lambda.txt; \
	cp handler.py dist

