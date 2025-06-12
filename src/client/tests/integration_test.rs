
use client::iot;

#[test]
fn test_iot() {
    iot::Client::new_without_cert();
}
