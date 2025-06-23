use reqwest::blocking::get;
use reqwest::blocking::Response;
use std::collections::HashMap;
use std::{io, io::Write};
use sysinfo::System;
mod iot;
mod stack;
mod ws;

fn main() -> () {
    let mut mqtt_client_option: Option<iot::MqttClient> = None;
    let mut ws_client_option: Option<ws::WebSocketClient> = None;
    let stack = stack::Stack::new();

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
            "register" => {
                println!("Registering to server...");
                iot::MqttClient::register(
                    "rust-device".to_string(),
                    stack.device_registration_endpoint.clone(),
                );
            }
            "connect iot" => {
                println!("Connecting to mqtt server...");
                mqtt_client_option =
                    Some(iot::MqttClient::connect(stack.iot_core_endpoint.clone()));
                println!("Connected to mqtt server")
            }
            "publish iot" => {
                println!("Publishing message to mqtt server...");
                if let Some(client) = mqtt_client_option.as_mut() {
                    client.publish("hello/1/world", "hello");
                    println!("Message published")
                } else {
                    println!("No client connected");
                }
            }
            "connect ws" => {
                println!("Connecting to ws server...");
                ws_client_option = Some(ws::WebSocketClient::connect(
                    "ws://localhost:8080".to_string(),
                ));
                println!("Connected to ws server")
            }
            "send ws" => {
                println!("Sending message to ws server...");
                if let Some(client) = ws_client_option.as_mut() {
                    client.send("hello from rust device".to_string());
                    println!("Message sent");
                } else {
                    println!("No client connected");
                }
            }
            "post" => {
                println!("Sending post request to server...");
                let response = post_server();
                println!("Response: {}", response);
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
