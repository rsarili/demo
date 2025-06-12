use reqwest::blocking::get;
use reqwest::blocking::Response;
use rumqttc::Packet;
use rumqttc::Publish;
use rumqttc::Transport;
use rumqttc::{MqttOptions, QoS, TlsConfiguration};
use std::collections::HashMap;
use std::io::Read;
use std::process::Command;
use std::ptr::null;
use std::time::Duration;
use std::{io, io::Write};
use sysinfo::System;
mod iot;

fn main() -> () {
    let mut client_option: Option<iot::Client> = None;

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
            "connect" => {
                println!("Connecting to server...");
                client_option = Some(iot::Client::new());
            }
            "post" => {
                println!("Sending post request to server...");
                let response = post_server();
                println!("Response: {}", response);
            }
            "publish" => {
                println!("Sending publish request to server...");
                if let Some(client) = client_option.as_mut() {
                    client.publish("hello/1/world", "hello");
                } else {
                    println!("No client connected");
                }
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
