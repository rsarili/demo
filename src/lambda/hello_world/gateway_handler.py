import requests
from requests import Response

from aws_lambda_powertools import Logger
from aws_lambda_powertools.event_handler import APIGatewayRestResolver
from aws_lambda_powertools.logging import correlation_paths
from aws_lambda_powertools.utilities.typing import LambdaContext

from utils import log_uncaught_exceptions

logger = Logger()
app = APIGatewayRestResolver()


@app.post("/todos")
def post_todos():
    data: dict = app.current_event.json_body
    logger.append_keys(user_id=data.get("userId"))
    
    if data.get("error") == 500:
        raise Exception("unhandled error")
    
    todo: Response = requests.post("https://jsonplaceholder.typicode.com/todos", data=data)

    logger.info("request succcesful")

    return {"todos": todo.json()}, 201

@log_uncaught_exceptions(logger=logger)
@logger.inject_lambda_context(log_event=True, clear_state=True, correlation_id_path=correlation_paths.API_GATEWAY_HTTP)
def handler(event: dict, context: LambdaContext) -> dict:
    return app.resolve(event, context)