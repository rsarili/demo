import requests
from requests import Response

from aws_lambda_powertools import Logger
from aws_lambda_powertools.event_handler import APIGatewayRestResolver
from aws_lambda_powertools.logging import correlation_paths
from aws_lambda_powertools.utilities.typing import LambdaContext

from utils import log_uncaught_exceptions
import boto3
from types_boto3_iot import IoTClient
from types_boto3_dynamodb import DynamoDBClient
from types_boto3_iot.type_defs import CreateKeysAndCertificateResponseTypeDef
import os

logger = Logger()
app = APIGatewayRestResolver()
iot_client: IoTClient = boto3.client("iot")
dynamodb_client: DynamoDBClient = boto3.client("dynamodb")

@app.post("/certificates")
def create_certificate():
    body: dict = app.current_event.json_body

    keys_and_certificate: CreateKeysAndCertificateResponseTypeDef = iot_client.create_keys_and_certificate(setAsActive=True)
    iot_client.attach_policy(policyName=os.environ["POLICY_NAME"], target=keys_and_certificate["certificateArn"])

    dynamodb_client.put_item(TableName=os.environ["TABLE_NAME"], Item={
        "deviceId": {"S": body["deviceId"]},
        "certificateArn": {"S": keys_and_certificate["certificateArn"]}
    })

    return {"certificate":keys_and_certificate["certificatePem"], "public_key": keys_and_certificate["keyPair"]["PublicKey"], "private_key": keys_and_certificate["keyPair"]["PrivateKey"]}, 201


@log_uncaught_exceptions(logger=logger)
@logger.inject_lambda_context(log_event=True, clear_state=True, correlation_id_path=correlation_paths.API_GATEWAY_HTTP)
def handler(event: dict, context: LambdaContext) -> dict:
    return app.resolve(event, context)