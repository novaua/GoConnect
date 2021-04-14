# GoConnect
TCP to HTTP to TCP brige using TLS enclryption

First result via 10 Gbit network
64 K buffer
1 length queue

$ iperf3 -c 127.0.0.1 -p 7001
Connecting to host 127.0.0.1, port 7001
[  4] local 127.0.0.1 port 41364 connected to 127.0.0.1 port 7001
[ ID] Interval           Transfer     Bandwidth       Retr  Cwnd
[  4]   0.00-1.00   sec   366 MBytes  3.07 Gbits/sec    0   3.56 MBytes
[  4]   1.00-2.00   sec   382 MBytes  3.21 Gbits/sec    0   3.56 MBytes
[  4]   2.00-3.00   sec   432 MBytes  3.63 Gbits/sec    0   3.56 MBytes
[  4]   3.00-4.00   sec   429 MBytes  3.60 Gbits/sec    0   3.56 MBytes
[  4]   4.00-5.00   sec   401 MBytes  3.37 Gbits/sec    0   3.56 MBytes
[  4]   5.00-6.00   sec   381 MBytes  3.20 Gbits/sec    0   3.56 MBytes
[  4]   6.00-7.00   sec   415 MBytes  3.48 Gbits/sec    0   3.56 MBytes
[  4]   7.00-8.00   sec   418 MBytes  3.50 Gbits/sec    0   3.56 MBytes
[  4]   8.00-9.00   sec   402 MBytes  3.38 Gbits/sec    0   3.56 MBytes
[  4]   9.00-10.00  sec   388 MBytes  3.25 Gbits/sec    0   3.56 MBytes
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bandwidth       Retr
[  4]   0.00-10.00  sec  3.92 GBytes  3.37 Gbits/sec    0             sender
[  4]   0.00-10.00  sec  3.91 GBytes  3.36 Gbits/sec                  receiver

