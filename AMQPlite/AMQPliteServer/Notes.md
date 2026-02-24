Each Broker has a Connection handler
so the Connection handler is a separate goroutine that reads data from its respective connection and lives till the connection is closed.
Connection handler:
1. A Read loop
2. Initiate connection handshake
3. send & receive frames to/from channel manager[need to decide on how the input is sent and recieved]
4. frames with channel 0, are handled by connection control function
5. a Writer loop that writes data to the underlying connection from the write buffer ties to every connection. all frames to be sent to the client must be sent to the write buffer
6. Implement Error handling and Context functions.

Each Connection has a Channel Manager which stores the created channels.
Each channel is like an independent worker or Goroutine. The channel manager just acts like a registry.




20/02/26
Implement the required functions for the ConnectionHandler
--- Connection control function
--- Channel Manager
--- Writer goroutine
--- Connection class
--- Channel class



22/02/26:

--- For methods that are synchronous in nature, the broker should not receive any other message or process any other message until it receives the appropraite response. Ex: Connection.Start method

--- We can have a frame validation function before passing it to the next function. we set an expected method so just as soon the 


23/02/26:
--- Implement Field table
An AMQP Field Table is prefixed by its total length, followed by a series of name-value pairs.
Field Table = Length (4 bytes) + Field 1 + Field 2+ ... +Field N. 
Each field follows this specific structure:
    1.Field Name Length (1 byte): The length of the key's string.
    2.Field Name (N bytes): The key string itself (UTF-8).
    3.Value Type Tag (1 byte): A single character representing the data type (e.g., 'S' for string, 'I' for integer).
    4.Field Value (Variable): The actual data, encoded based on the type tag.