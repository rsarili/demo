use reqwest::blocking::get;
use reqwest::blocking::Response;
use rumqttc::Packet;
use rumqttc::Transport;
use rumqttc::{Client, MqttOptions, QoS, TlsConfiguration};
use std::collections::HashMap;
use std::io::Read;
use std::time::Duration;
use std::{io, io::Write};
use sysinfo::System;

fn main() -> () {
    let ca = read("./AmazonRootCA1.pem");
    let client_cert = read("./device-certificate.pem.crt");
    let client_key = read("./device-private.pem.key");

    let transport = Transport::Tls(TlsConfiguration::Simple {
        ca,
        alpn: None,
        client_auth: Some((client_cert, client_key)),
    });

    let mut mqttoptions = MqttOptions::new(
        "rust-client",
        "<your-endpoint>-ats.iot.eu-central-1.amazonaws.com",
        8883,
    );
    mqttoptions.set_transport(transport);
    mqttoptions.set_keep_alive(Duration::from_secs(60));

    //https://github.com/bytebeamio/rumqtt/blob/main/rumqttc/examples/syncpubsub.rs
    let (client, mut connection) = rumqttc::Client::new(mqttoptions, 10);
    client.subscribe("hello/+/world", QoS::AtLeastOnce).unwrap();

    std::thread::spawn(move || listen_mqtt(&mut connection));

    loop {
        print!("Enter command: ");
        io::stdout().flush().unwrap();

        let mut input = String::new();
        io::stdin().read_line(&mut input).unwrap();
        let command = input.trim();

        match command {
            "ping" => {
                println!("Sending ping to server...");
                let response = ping_server();
                println!("Response: {}", response);
            }
            "create_certificate" => {
                println!("Creating certificate...");
                create_certificate();
            }
            "post" => {
                println!("Sending post request to server...");
                let response = post_server();
                println!("Response: {}", response);
            }
            "publish" => {
                println!("Sending publish request to server...");
                client
                    .publish("hello/1/world", QoS::AtLeastOnce, false, "hello")
                    .unwrap();
            }
            "system" => {
                let sytem = System::new_all();

                println!(
                    "total memory    : {} MB",
                    sytem.total_memory() / (1024 * 1024)
                );
                println!(
                    "available memory: {} MB",
                    sytem.available_memory() / (1024 * 1024)
                );
            }
            "exit" => break,
            _ => println!("Unknown command: {}", command),
        }
    }
}

fn post_server() -> String {
    let mut body = HashMap::new();

    body.insert("name", "John Doe");
    body.insert("age", "30");

    let url: &str = "http://localhost:8080";

    let client = reqwest::blocking::Client::new();
    let response = client.post(url).json(&body).send().unwrap();

    return response.text().unwrap();
}

fn create_certificate () {
    let mut body = HashMap::new();

    body.insert("name", "John Doe");
    body.insert("age", "30");

    let url: &str = "<endpoint>/certificates";

    let client = reqwest::blocking::Client::new();
    let response = client.post(url).json(&body).send().unwrap();
    

    println!("Response: {}", response.text().unwrap());
}

fn read(path: &str) -> Vec<u8> {
    let mut file = std::fs::File::open(path).unwrap();
    let mut contents = Vec::new();
    file.read_to_end(&mut contents).unwrap();
    return contents;
}

fn ping_server() -> String {
    // The URL you want to send the GET request to
    let url: &str = "http://localhost:8080";

    // Send the GET request (blocking call)
    let response: Response = get(url).unwrap();

    // Check if the request was successful
    if response.status().is_success() {
        return response.text().unwrap(); // Return the response body as a String
    } else {
        return response.text().unwrap();
    }
}

fn publish(client: Client) {
    for i in 0..10 {
        std::thread::sleep(Duration::from_secs(1));
        let topic = format!("hello/{i}/world");
        let qos = QoS::AtLeastOnce;

        client.publish(topic, qos, false, "hello").unwrap();
    }
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
