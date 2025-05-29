use rand::Rng;
use reqwest::blocking::get;
use reqwest::blocking::Response;
use reqwest::Error;
use std::thread;
use std::time::Duration;

fn main() -> Result<(), Error> {
    let mut handles = vec![];

    for _ in 0..10 {
        let handle = thread::spawn(|| {
            thread::sleep(Duration::from_secs(get_random()));
            if let Err(e) = ping_server() {
                eprintln!("Error: {:?}", e);
            }
        });

        handles.push(handle);
    }

    for handle in handles {
        handle.join().unwrap();
    }

    Ok(())
}

fn ping_server() -> Result<(), Error> {
    // The URL you want to send the GET request to
    let url: &str = "http://localhost:8080";

    // Send the GET request (blocking call)
    let response: Response = get(url)?;

    // Check if the request was successful
    if response.status().is_success() {
        let body = response.text()?;
        println!("Response body: {}", body);
    } else {
        println!("Failed to get a valid response.");
    }
    Ok(())
}

fn get_random() -> u64 {
    return rand::rng().random_range(1..=10);
}
