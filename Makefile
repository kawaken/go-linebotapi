ENV_FILE := ./env.json

deploy:
	apex deploy --env-file ${ENV_FILE}

.PHONEY: deploy
