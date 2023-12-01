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

## Chat-GPT回答
- Q: TCP/IP socketもUNIX socketも実態はfileという理解は正しいか？
> 1. **UNIXソケット（UNIX Domain Socket）**:
>    - UNIXソケットは、UNIXおよびUNIXライクなシステム（Linuxなど）で使用される通信インターフェースです。
>    - これはファイルシステム上のパスを使用して識別されますが、通常のファイルとは異なり、データストリームを介してプロセス間で通信を行います。
>    - UNIXソケットは、ファイルディスクリプタを介して操作される点でファイルに似ていますが、データの永続性やファイルシステム上での振る舞いは通常のファイルとは異なります。
>
> 2. **TCP/IPソケット**:
>    - TCP/IPソケットは、ネットワークを介した通信を可能にするインターフェースです。
>    - これらはIPアドレスとポート番号によって識別され、データの送受信に使用されます。
>    - TCP/IPソケットはファイルディスクリプタを使用して操作されることがありますが、これはUNIXソケットと同様、操作の抽象化を提供するためです。実際には、ファイルシステムのファイルとは異なる性質を持ちます。
> 
> したがって、TCP/IPソケットとUNIXソケットは、操作方法がファイルに似ているという点で「ファイルのように扱われる」ことがありますが、通信チャネルとしての役割やデータの扱い方が通常のファイルとは大きく異なるため、単純に「ファイル」と同じだとは言えません。

> 1. **UNIXソケット**:
>    - ファイルシステム上にパスとして現れますが、これは通信エンドポイントを識別するためのものです。
>    - ファイルのようにオープン、読み書き、クローズといった操作が可能ですが、その内容はデータの永続的なストレージではなく、プロセス間通信（IPC）のチャネルを提供します。
>    - そのため、ファイルシステム上に存在するものの、通常のファイルとして読み書きされるデータ（テキストやバイナリデータなど）とは異なり、プロセス間でのリアルタイムなデータのやり取りに使用されます。
>
> 2. **TCP/IPソケット**:
>    - TCP/IPソケットは、ファイルシステム上に直接現れることはありません。
>    - これらはネットワークインターフェースを介して通信を行いますが、ファイルディスクリプタを介して操作されることが多いです。
>    - その操作方法がファイルの読み書きに似ているため、ファイルのように扱われることがありますが、実際にはネットワーク上のデータストリームを表しており、ファイルシステム上にデータを保存するわけではありません。
>
> つまり、UNIXソケットはファイルシステム上に存在するが、通常のファイルとは異なる動作をする特殊なファイルと言えます。一方で、TCP/IPソケットはファイルシステム上には現れませんが、ファイルのように操作されることがあります。どちらも、その背後にある主な目的は通信であり、通常のファイルとしてのデータの保存や管理とは異なる点に注意が必要です。