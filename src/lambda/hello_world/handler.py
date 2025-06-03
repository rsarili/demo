from functools import wraps
from aws_lambda_powertools import Logger
from aws_lambda_powertools.utilities.typing import LambdaContext

from utils import log_uncaught_exceptions

logger: Logger = Logger()


@log_uncaught_exceptions(logger=logger)
@logger.inject_lambda_context(log_event=True, clear_state=True)
def handler(event, context: LambdaContext):
    request_id: str = event["request-id"]
    logger.append_keys(request_id=request_id)

    raise Exception("unexpected error happened")
    
    return "Hello World!"