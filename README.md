### Requirements & Implementations

    broadcast server - This command will start the server.
    broadcast client - This command will connect the client to the server.

    1. Create a server that listens for incoming connections. - done
    2. When a client connects, store the connection in a list of connected clients. - done
    3. When a client sends a message, broadcast this message to all connected clients. - done
    4. Handle client disconnections and remove the client from the list of connected clients. - done
    5. Implement a client that can connect to the server and send messages. - done
    6. Test the server by connecting multiple clients and sending messages. - done
    7. Implement error handling and graceful shutdown of the server. - done

    - implemented using tcp connection rather than web socket
    - implemented little chat system among connected clients
    - implemented disconnecting from a client when 'exit' typed in

### more

    - this program is currently sending fixed 4096 byte size messages only -> implement collecting all the buffers for longer messages
    - give random nicknames to clients
