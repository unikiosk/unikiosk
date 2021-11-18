Design notes
============

transocks should work as a SOCKS5 client used as a transparent proxy
agent running on every hosts in trusted (i.e. data center) networks.

Destination NAT (DNAT)
----------------------

On Linux, redirecting locally-generated packet to transocks can be done
by iptables with DNAT (or REDIRECT) target.

Since DNAT/REDIRECT modifies packet's destination address, transocks
need to recover the destination address by using `getsockopt` with
`SO_ORIGINAL_DST` for IPv4 or with `IP6T_SO_ORIGINAL_DST` for IPv6.
This is, of course, Linux-specific, and Go does not provide standard
API for them.

Policy-based routing
--------------------

Except for DNAT, some operating systems provide a way to route packets
to a specific program.  In order to receive such packets, the program
need to set special options on the listening socket before `bind`.

Difficult is that Go does not allow setting socket options before `bind`.

### Linux TPROXY

Linux iptables has [TPROXY][] target that can route packets to a
specific local port.  The socket option is:

* IPv4

    ```
    setsockopt(IPPROTO_IP, IP_TRANSPARENT)
    ```

* IPv6

    ```
    setsockopt(IPPROTO_IPV6, IPV6_TRANSPARENT)
    ```

To set this option, transocks must have `CAP_NET_ADMIN` capability.
Run transocks as root user, or grant `CAP_NET_ADMIN` for the file by:

```
sudo setcap 'cap_net_admin+ep' transocks
```

### FreeBSD, NetBSD, OpenBSD

Use [PF with divert-to][pf] to route packets to a specific local port.

The listening program needs to set a socket option before `bind`:

* FreeBSD (IPv4)

    ```
    setsockopt(IPPROTO_IP, IP_BINDANY)
    ```

* FreeBSD (IPv6)

    ```
    setsockopt(IPPROTO_IPV6, IPV6_BINDANY)
    ```

* NetBSD, OpenBSD

    ```
    setsockopt(SOL_SOCKET, SO_BINDANY)
    ```

For this to work, transocks must run as root.

Implementation strategy
-----------------------

We use Go for its efficiency and simpleness.

For SOCKS5, [golang.org/x/net/proxy][x/net] already provides SOCKS5 client.

For Linux NAT, we need to use [golang.org/x/sys/unix][x/sys] and
[unsafe.Pointer][] to use non-standard `getsockopt` options.

To set socket options before `bind`, we need to create sockets manually
by using [golang.org/x/sys/unix] and then convert the native socket to
`*net.TCPListener` by [net.FileListener][].

CONNECT tunnel
--------------

As golang.org/x/net/proxy can add custom dialers, we can implement
a proxy using http CONNECT method for tunneling through HTTP proxies
such as [Squid][].

Note that the default Squid configuration file installed for
Ubuntu 14.04 prohibits CONNECT to ports other than 443.

```
# Deny CONNECT to other than secure SSL ports
http_access deny CONNECT !SSL_ports
```

Remove or comment out the line to allow CONNECT to ports other than 443.

[TPROXY]: https://www.kernel.org/doc/Documentation/networking/tproxy.txt
[pf]: http://wiki.squid-cache.org/ConfigExamples/Intercept/OpenBsdPf
[x/net]: https://godoc.org/golang.org/x/net/proxy#SOCKS5
[x/sys]: https://godoc.org/golang.org/x/sys/unix
[unsafe.Pointer]: https://golang.org/pkg/unsafe/#Pointer
[net.FileListener]: https://golang.org/pkg/net/#FileListener
[Squid]: http://www.squid-cache.org/
