# KnockD

Port knocking UDP Server (Supported only iptables and ipv4)

Installation
--------

Download: <a href="https://github.com/eltaline/knockd/releases">Releases</a>

Configuration
--------

1. Edit ExecStart= in ```/lib/systemd/system/knockd.service``` file

    - Default settings ```/usr/sbin/knockd --fport=47566 --sport=7566 --tcpports=21,22,25,111 --udpports=111,135,136,137,138,139 --ttl=86400 --timeout=60```

    - fport: first knocking udp port
    - sport: second knocking udp port
    - tcpports: close this tcp ports
    - udpports: close this udp ports
    - ttl: time to live of access rule for ip
    - timeout: timeout beetween knocks to fport and sport

2. ```systemctl daemon-reload ; systemctl enable knockd```
3. ```systemctl daemon-reload ; systemctl restart knockd```

Port knocking clients
--------

    - Windows: <a href="https://www.microsoft.com/en-us/download/confirmation.aspx?id=24009">PortQRY</a>
    - Linux: ```nmap -sU 47566,7566 1.2.3.4```

--------
End