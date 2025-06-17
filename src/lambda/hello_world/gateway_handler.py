import os

import boto3
from aws_lambda_powertools import Logger, Metrics
from aws_lambda_powertools.event_handler import APIGatewayRestResolver
from aws_lambda_powertools.logging import correlation_paths
from aws_lambda_powertools.metrics.base import MetricUnit
from aws_lambda_powertools.utilities.idempotency import (
    DynamoDBPersistenceLayer,
    IdempotencyConfig,
    idempotent,
)
from aws_lambda_powertools.utilities.typing import LambdaContext
from types_boto3_dynamodb.service_resource import DynamoDBServiceResource
from types_boto3_iot import IoTClient
from types_boto3_iot.type_defs import CreateKeysAndCertificateResponseTypeDef

from certificate_storage import CertificateStorage
from env_variables import EnvironmentVariables
from utils import log_uncaught_exceptions

logger = Logger()
app = APIGatewayRestResolver()
iot_client: IoTClient = boto3.client("iot")
dynamodb_service: DynamoDBServiceResource = boto3.resource("dynamodb")
metrics: Metrics = Metrics()

certificate_storage: CertificateStorage = CertificateStorage(
    logger=logger,
    table=dynamodb_service.Table(
        os.environ[EnvironmentVariables.CERTIFICATES_TABLE]
    ),
)
idempotency_persistence_layer: DynamoDBPersistenceLayer = (
    DynamoDBPersistenceLayer(
        table_name=os.environ[EnvironmentVariables.IDEMPOTENCY_TABLE]
    )
)


@app.post("/certificates")
def create_certificate():
    body: dict = app.current_event.json_body
    logger.append_keys(device_id=body["deviceId"])

    keys_and_certificate: CreateKeysAndCertificateResponseTypeDef = (
        iot_client.create_keys_and_certificate(setAsActive=True)
    )
    iot_client.attach_policy(
        policyName=os.environ[EnvironmentVariables.IOT_DEVICE_POLICY],
        target=keys_and_certificate["certificateArn"],
    )

    certificate_storage.add_certificate(
        device_id=body["deviceId"],
        certificate_arn=keys_and_certificate["certificateArn"],
    )

    metrics.add_dimension(name="OperationName", value="CertificatesCreated")
    metrics.add_metric(name="Success", unit=MetricUnit.Count, value=1)

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
@idempotent(
    persistence_store=idempotency_persistence_layer,
    key_prefix="create_iot_certificate",
    config=IdempotencyConfig(
        event_key_jmespath="body",
    ),
)
@metrics.log_metrics
def handler(event: dict, context: LambdaContext) -> dict:
    return app.resolve(event, context)
