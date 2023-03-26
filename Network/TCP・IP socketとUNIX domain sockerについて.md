## TCP/IP socketとUNIX domain socketの違い
> TCP/IP sockets and Unix domain sockets are both used for inter-process communication (IPC), but they are different in the scope and mechanism they use to communicate between processes.
>
> TCP/IP sockets (also known as network sockets) are used to communicate between processes running on different computers connected over a network. They use the standard IP protocols (TCP and UDP) to transmit data and can communicate across a wide range of networks, including the Internet.
>
> Unix domain sockets, on the other hand, are used for communication between processes running on the same host. Instead of using IP addresses and network protocols, Unix domain sockets use the file system as their communication mechanism. This makes communication between processes faster and more secure, since the data does not travel over the network and is subject to the same access controls as file system operations.
>
> In summary, the main difference between TCP/IP sockets and Unix domain sockets is the scope of the communication and the mechanism used for communication. TCP/IP sockets are used for communication between processes on different computers over a network, while Unix domain sockets are used for communication between processes on the same host.

## TCP/IP socket
> A socket is a endpoint for communication between two devices over a network. In the context of the TCP/IP protocol suite, a socket is a combination of an IP address and a port number that serves as a communication endpoint between two processes. The combination of the IP address and port number identifies a specific process on a specific device, and provides a way for other devices on the network to communicate with that process.
>
> TCP/IP sockets are used to build client-server applications, where the client establishes a connection to the server and sends requests, and the server responds to these requests. The communication between the client and server happens through the socket, and data is exchanged in the form of packets.
>
> Sockets can be used for many different types of communication, including file transfer, email, web traffic, and others. They are a fundamental building block for many Internet applications and services.
- TCP/IP socketも**file**
  - Linux(Unix)ではすべてをfileとして見なす
- 参考URL
  - https://ikcoo.tistory.com/119
  - https://helloworld-88.tistory.com/215
  - https://kwonzales.tistory.com/36
    - `socket`でsocketを生成
    - `bind`はIPアドレスとPort番号を設定する(割り当てる)ことを指す
    - `listen`はclientからrequestを受け付けられるようにsocketを待機させること

## UNIX domain socket
> A Unix domain socket (also known as a Unix socket) is a type of inter-process communication (IPC) mechanism used in Unix-like operating systems. It allows processes running on the same system to communicate with each other through a **socket file in the file system**, rather than through a network interface.
>
> Unix domain sockets are faster and more secure than network sockets because they avoid the overhead of network protocol processing and data copying between kernel and user space. They also have the advantage of being able to communicate between processes that do not have network connectivity or privileges to access the network.
>
> To use a Unix domain socket, a server process creates a **socket file** in a designated directory with a unique name and binds it to a socket address. The client process then connects to the server by specifying the socket address and sending data through the socket. The server receives the data and responds to the client through the same socket.
>
> Unix domain sockets are widely used in various applications such as databases, web servers, and audio servers. They are also used by some programming languages, such as Python and Ruby, to implement IPC mechanisms.

## 4 Way Handshake
- https://beenii.tistory.com/127