moses
=====

Moses is a proxy framework that can support socks5 and other customed proxy protocols.
Also it has provided a socks5 implementation in server side, with which you can build your own socks5 server.
By default, moses has offered a local proxy server and a remote proxy server, which are located in client and server subdirectories perspectively.
The client can be built by a command "go build github.com/steve-wang/moses/client".
The server can be built by a command "go build github.com/steve-wang/moses/server".
The route is like:
application <=(socks5)=> proxy client <=(network)=> proxy server <=(network)=> target server
