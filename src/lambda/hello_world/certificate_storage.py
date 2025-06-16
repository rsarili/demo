from aws_lambda_powertools import Logger
from types_boto3_dynamodb.service_resource import Table


class CertificateStorage:
    def __init__(self, table: Table, logger: Logger):
        self._table: Table = table
        self._logger: Logger = logger

    def add_certificate(self, device_id: str, certificate_arn: str):
        self._table.put_item(
            Item={"deviceId": device_id, "certificateArn": certificate_arn}
        )
