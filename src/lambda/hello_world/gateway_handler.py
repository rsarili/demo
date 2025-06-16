import os

import boto3
from aws_lambda_powertools import Logger
from aws_lambda_powertools.event_handler import APIGatewayRestResolver
from aws_lambda_powertools.logging import correlation_paths
from aws_lambda_powertools.utilities.typing import LambdaContext
from types_boto3_dynamodb.service_resource import DynamoDBServiceResource, Table
from types_boto3_iot import IoTClient
from types_boto3_iot.type_defs import CreateKeysAndCertificateResponseTypeDef

from certificate_storage import CertificateStorage
from utils import log_uncaught_exceptions

logger = Logger()
app = APIGatewayRestResolver()
iot_client: IoTClient = boto3.client("iot")
dynamodb_service: DynamoDBServiceResource = boto3.resource("dynamodb")

certificate_storage: CertificateStorage = CertificateStorage(
    logger=logger, table=dynamodb_service.Table(os.environ["TABLE_NAME"])
)


@app.post("/certificates")
def create_certificate():
    body: dict = app.current_event.json_body

    keys_and_certificate: CreateKeysAndCertificateResponseTypeDef = (
        iot_client.create_keys_and_certificate(setAsActive=True)
    )
    iot_client.attach_policy(
        policyName=os.environ["POLICY_NAME"],
        target=keys_and_certificate["certificateArn"],
    )

    certificate_storage.add_certificate(
        device_id=body["deviceId"],
        certificate_arn=keys_and_certificate["certificateArn"],
    )

    return {
        "certificate": keys_and_certificate["certificatePem"],
        "public_key": keys_and_certificate["keyPair"]["PublicKey"],
        "private_key": keys_and_certificate["keyPair"]["PrivateKey"],
    }, 201


@log_uncaught_exceptions(logger=logger)
@logger.inject_lambda_context(
    log_event=True,
    clear_state=True,
    correlation_id_path=correlation_paths.API_GATEWAY_HTTP,
)
def handler(event: dict, context: LambdaContext) -> dict:
    return app.resolve(event, context)


class CertificateStorage:
    def __init__(self, table: Table, logger: Logger):
        self._table: Table = table
        self._logger: Logger = logger

    def add_certificate(self, device_id: str, certificate_arn: str):
        self._table.put_item(
            Item={"deviceId": device_id, "certificateArn": certificate_arn}
        )
