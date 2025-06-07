use rand::Rng;
use reqwest::blocking::get;
use reqwest::blocking::Client;
use reqwest::blocking::Response;
use reqwest::Error;
use std::collections::HashMap;
use std::process;
use std::thread;
use std::time::Duration;
use std::{io, io::Write};

fn main() -> Result<(), Error> {
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
            "post" => {
                println!("Sending post request to server...");
                let response = post_server();
                println!("Response: {}", response);
            }
            "exit" => break,
            _ => println!("Unknown command: {}", command),
        }
    }

    Ok(())
}

fn post_server() -> String {
    let mut body = HashMap::new();

    body.insert("name", "John Doe");
    body.insert("age", "30");

    let url: &str = "http://localhost:8080";

    let client = Client::new();
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
