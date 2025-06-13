use client::iot;
use client::stack;

#[test]
fn test_iot() {
    let stack = stack::Stack::new();
    iot::Client::register(stack.device_registration_endpoint);
    let client = iot::Client::connect(stack.iot_core_endpoint);
    client.publish("hello/1/world", "hello");
}
