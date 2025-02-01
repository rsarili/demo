use reqwest::blocking::get;
use reqwest::blocking::Response;
use reqwest::Error;

fn main() -> Result<(), Error> {
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
