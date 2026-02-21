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
