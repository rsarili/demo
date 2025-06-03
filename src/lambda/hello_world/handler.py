from aws_lambda_powertools import Logger

logger: Logger = Logger(service="wer")

def handler(event, context):
    logger.info("wer")
    return "Hello World!"