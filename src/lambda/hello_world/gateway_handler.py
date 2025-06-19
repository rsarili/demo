import os

import boto3
from aws_lambda_powertools import Logger, Metrics
from aws_lambda_powertools.event_handler import APIGatewayRestResolver
from aws_lambda_powertools.event_handler.exceptions import NotFoundError
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

from device_storage import Device, DeviceStorage
from env_variables import EnvironmentVariables
from metrics import DimensionNames, MetricNames, OperationNames
from models import GetDeviceResponse, PostDeviceRequest, PostDeviceResponse
from utils import log_uncaught_exceptions

logger = Logger()
app = APIGatewayRestResolver(enable_validation=True)
app.enable_swagger(path="/swagger")

iot_client: IoTClient = boto3.client("iot")
dynamodb_service: DynamoDBServiceResource = boto3.resource("dynamodb")
metrics: Metrics = Metrics()

device_storage: DeviceStorage = DeviceStorage(
    logger=logger,
    table=dynamodb_service.Table(
        os.environ[EnvironmentVariables.DEVICES_TABLE]
    ),
)
idempotency_persistence_layer: DynamoDBPersistenceLayer = (
    DynamoDBPersistenceLayer(
        table_name=os.environ[EnvironmentVariables.IDEMPOTENCY_TABLE]
    )
)


@app.post("/devices", responses={201: {"description": "Device is created."}})
def create_device(request: PostDeviceRequest) -> PostDeviceResponse:
    logger.append_keys(device_id=request.device_id)

    keys_and_certificate: CreateKeysAndCertificateResponseTypeDef = (
        iot_client.create_keys_and_certificate(setAsActive=True)
    )
    iot_client.attach_policy(
        policyName=os.environ[EnvironmentVariables.IOT_DEVICE_POLICY],
        target=keys_and_certificate["certificateArn"],
    )

    device_storage.add_device(
        device=Device(
            device_id=request.device_id,
            device_type=request.device_type,
            certificate_arn=keys_and_certificate["certificateArn"],
        )
    )

    metrics.add_dimension(
        name=DimensionNames.OPERATION_NAME, value=OperationNames.DEVICES_CREATED
    )
    metrics.add_metric(name=MetricNames.SUCCESS, unit=MetricUnit.Count, value=1)

    return PostDeviceResponse(
        certificate=keys_and_certificate["certificatePem"],
        public_key=keys_and_certificate["keyPair"]["PublicKey"],
        private_key=keys_and_certificate["keyPair"]["PrivateKey"],
    )


@app.get(
    "/devices/<device_id>",
    responses={
        200: {"description": "Device is found."},
        404: {"description": "Device is not found."},
    },
)
def get_device(device_id: str) -> GetDeviceResponse:
    device: Device | None = device_storage.get_device(device_id=device_id)

    if device is None:
        raise NotFoundError(f"device not found, device id: {device_id}")

    return GetDeviceResponse.from_device(device=device)


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
