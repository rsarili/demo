from pydantic import BaseModel, ConfigDict
from pydantic.alias_generators import to_camel

from device_storage import Device


class CamelCaseBaselModel(BaseModel):
    model_config = ConfigDict(
        populate_by_name=True,
        populate_by_alias=True,
        serialize_by_alias=True,
        alias_generator=to_camel,
    )


class PostDeviceRequest(CamelCaseBaselModel):
    device_id: str
    device_type: str


class PostDeviceResponse(CamelCaseBaselModel):
    certificate: str
    public_key: str
    private_key: str


class GetDeviceResponse(CamelCaseBaselModel):
    device_id: str
    device_type: str
    certificate_arn: str

    @classmethod
    def from_device(cls, device: Device) -> "GetDeviceResponse":
        return GetDeviceResponse(
            device_id=device.device_id,
            device_type=device.device_type,
            certificate_arn=device.certificate_arn,
        )
