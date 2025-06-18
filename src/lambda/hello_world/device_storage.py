from dataclasses import dataclass

from aws_lambda_powertools import Logger
from types_boto3_dynamodb.service_resource import Table


@dataclass
class Device:
    device_id: str
    device_type: str
    certificate_arn: str


class DeviceStorage:
    def __init__(self, table: Table, logger: Logger):
        self._table: Table = table
        self._logger: Logger = logger

    def add_device(self, device: Device) -> None:
        self._table.put_item(
            Item={
                "deviceId": device.device_id,
                "deviceType": device.device_type,
                "certificateArn": device.certificate_arn,
            }
        )

    def get_device(self, device_id: str) -> Device | None:
        response = self._table.get_item(Key={"deviceId": device_id})

        if "Item" not in response:
            return None

        item = response["Item"]
        return Device(
            device_id=item["deviceId"],
            device_type=item["deviceType"],
            certificate_arn=item["certificateArn"],
        )
