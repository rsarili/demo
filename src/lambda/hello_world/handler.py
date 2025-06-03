from functools import wraps
from aws_lambda_powertools import Logger
from aws_lambda_powertools.utilities.typing import LambdaContext

logger: Logger = Logger()

def log_uncaught_exceptions(logger: Logger):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                logger.exception(e)
                raise
        return wrapper
    return decorator

@log_uncaught_exceptions(logger=logger)
@logger.inject_lambda_context(log_event=True, clear_state=True)
def handler(event, context: LambdaContext):
    request_id: str = event["request-id"]
    logger.append_keys(request_id=request_id)

    raise Exception("unexpected error happened")
    
    return "Hello World!"