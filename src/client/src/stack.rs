use aws_config::BehaviorVersion;
use aws_sdk_cloudformation::Client;
use std;
use tokio::runtime::Runtime;
pub struct Stack {
    pub iot_core_endpoint: String,
    pub device_registration_endpoint: String,
}

impl Stack {
    pub fn new() -> Stack {
        let rt = Runtime::new().unwrap();
        let stack_name = get_stack_name();

        let stack_output = rt.block_on(async {
            let config = aws_config::defaults(BehaviorVersion::latest()).load().await;
            let client: Client = Client::new(&config);
            return client
                .describe_stacks()
                .stack_name(stack_name.clone())
                .send()
                .await
                .unwrap();
        });

        let stacks = stack_output.stacks.unwrap();
        let stack = stacks.get(0).unwrap();

        let mut device_registration_endpoint: Option<String> = None;
        let mut iot_core_endpoint: Option<String> = None;

        for output in stack.outputs().into_iter() {
            match &output.export_name {
                Some(s) => {
                    if *s
                        == with_stack_name(
                            stack_name.clone(),
                            String::from("device-registration-endpoint"),
                        )
                    {
                        device_registration_endpoint = Some(output.output_value.clone().unwrap())
                    } else if *s
                        == with_stack_name(stack_name.clone(), String::from("iot-core-endpoint"))
                    {
                        iot_core_endpoint = Some(output.output_value.clone().unwrap())
                    }
                }
                _ => {}
            }
        }

        if device_registration_endpoint.is_none() {
            panic!("api_endpoint should not be None");
        }
        if iot_core_endpoint.is_none() {
            panic!("iot_core_endpoint should not be None");
        }

        return Stack {
            iot_core_endpoint: iot_core_endpoint.unwrap(),
            device_registration_endpoint: device_registration_endpoint.unwrap(),
        };
    }
}

pub fn with_stack_name(stack_name: String, resource_name: String) -> String {
    return format!("{}-{}", stack_name, resource_name);
}

pub fn get_stack_name() -> String {
    match std::env::var("STACK_NAME") {
        Ok(stack_name) => {
            return stack_name;
        }
        Err(_) => {
            return format!("iot-demo-{}", std::env::var("USER").unwrap());
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_stack() {
        let stack = Stack::new();
    }
}
