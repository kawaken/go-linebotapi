ENV_FILE := ./env.json

deploy:
	apex deploy --env-file ${ENV_FILE}

deploy_web:
	apex deploy --env-file ${ENV_FILE} web

deploy_bot:
	apex deploy --env-file ${ENV_FILE} bot

dynamo:
	cd ${PROJECT_DIR}/bin/dynamodblocal \
	&& java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -sharedDb

.PHONEY: deploy dynamo
