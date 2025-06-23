use std::net::TcpStream;
use tungstenite::{
    connect, http::Uri, stream::MaybeTlsStream, ClientRequestBuilder, Message,
    WebSocket,
};

pub struct WebSocketClient {
    socket: WebSocket<MaybeTlsStream<TcpStream>>,
}

impl WebSocketClient {
    pub fn connect(url: String) -> WebSocketClient {
        let uri: Uri = url.parse().unwrap();
        let builder = ClientRequestBuilder::new(uri);
        let socket = connect(builder).unwrap().0;

        return WebSocketClient { socket: socket };
    }

    pub fn send(&mut self, message: String) {
        self.socket.send(Message::from(message)).unwrap();
    }
}

fn read_ws(websocket: &mut WebSocket<MaybeTlsStream<TcpStream>>) {
    loop {
        match websocket.read().unwrap() {
            msg @ Message::Text(_) => {
                println!("incoming message: {}", msg);
            }
            _ => {}
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_connection() {
        let mut client = WebSocketClient::connect("ws://localhost:8080".to_string());
        client.send("hello from rust client".to_string());
    }
}
