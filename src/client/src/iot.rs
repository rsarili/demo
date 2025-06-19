use rumqttc::Packet;
use rumqttc::Transport;
use rumqttc::{MqttOptions, QoS, TlsConfiguration};
use std::collections::HashMap;
use std::time::Duration;
use std::{io::Read, io::Write};

pub struct Client {
    client: rumqttc::Client,
}

impl Client {
    pub fn connect(mqtt_endpoint: String) -> Client {
        return Client {
            client: connect(mqtt_endpoint),
        };
    }

    pub fn register(device_id: String, registration_endpoint: String) {
        create_certificate(device_id, registration_endpoint);
    }

    pub fn publish(&self, topic: &str, payload: &str) {
        self.client
            .publish(topic, QoS::AtLeastOnce, false, payload)
            .unwrap();
    }
}

#[derive(Debug, serde::Deserialize)] // Add this attribute macro
#[serde(rename_all = "camelCase")]
struct Certificate {
    public_key: String,
    private_key: String,
    certificate: String,
}

fn create_certificate(device_id: String, registration_endpoint: String) {
    let mut body = HashMap::new();

    body.insert("deviceId", device_id);
    body.insert("deviceType", "sensor".to_string());

    let client = reqwest::blocking::Client::new();
    let response = client
        .post(registration_endpoint)
        .json(&body)
        .send()
        .unwrap();

    if response.status() != 201 {
        println!(
            "Error, status: {}, body: {}",
            response.status(),
            response.text().unwrap()
        );
        return;
    }

    let cert = response.json::<Certificate>().unwrap();
    println!("Certificate: {cert:?}");

    write("device-certificate.pem.crt", cert.certificate.into_bytes());
    write("device-private.pem.key", cert.private_key.into_bytes());
    write("device-public.pem.key", cert.public_key.into_bytes());
}

fn connect(mqtt_endpoint: String) -> rumqttc::Client {
    let ca = read("./AmazonRootCA1.pem");
    let client_cert = read("./device-certificate.pem.crt");
    let client_key = read("./device-private.pem.key");

    let transport = Transport::Tls(TlsConfiguration::Simple {
        ca,
        alpn: None,
        client_auth: Some((client_cert, client_key)),
    });

    let port = 8883;

    let mut mqttoptions = MqttOptions::new("rust-client", &mqtt_endpoint, port);
    mqttoptions.set_transport(transport);
    mqttoptions.set_keep_alive(Duration::from_secs(60));

    println!(
        "connection to AWS Iot Core endpoint: {}, port: {}",
        mqtt_endpoint, port
    );
    //https://github.com/bytebeamio/rumqtt/blob/main/rumqttc/examples/syncpubsub.rs
    let (client, mut connection) = rumqttc::Client::new(mqttoptions, 10);
    client.subscribe("hello/+/world", QoS::AtLeastOnce).unwrap();

    std::thread::spawn(move || listen_mqtt(&mut connection));

    return client;
}

fn write(path: &str, contents: Vec<u8>) {
    let mut file = std::fs::File::create(path).unwrap();
    file.write_all(&contents).unwrap();
}

fn read(path: &str) -> Vec<u8> {
    let mut file = std::fs::File::open(path).unwrap();
    let mut contents = Vec::new();
    file.read_to_end(&mut contents).unwrap();
    return contents;
}

fn listen_mqtt(connection: &mut rumqttc::Connection) {
    for (i, notification) in connection.iter().enumerate() {
        match notification {
            Ok(notif) => {
                println!("{i}. Notification = {notif:?}");
                if let rumqttc::Event::Incoming(Packet::Publish(p)) = notif {
                    println!("Incoming message: {}", String::from_utf8_lossy(&p.payload));
                }
            }
            Err(error) => {
                println!("{i}. Notification = {error:?}");
                return;
            }
        }
    }
}

#[cfg(test)]
mod tests {

    #[test]
    fn it_works() {
        assert!(true);
    }
}
