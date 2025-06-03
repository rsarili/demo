from functools import wraps
from aws_lambda_powertools import Logger


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
