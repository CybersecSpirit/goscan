package main

import (
    "fmt"
    "net"
    "os"
	"time"
    "golang.org/x/net/icmp"
    "golang.org/x/net/ipv4"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Fprintf(os.Stderr, "Usage: %s IP/CIDR\n", os.Args[0])
        os.Exit(1)
    }

    ip, ipnet, err := net.ParseCIDR(os.Args[1])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Invalid CIDR: %v\n", err)
        os.Exit(1)
    }

    c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error listening: %v\n", err)
        os.Exit(1)
    }
    defer c.Close()

    for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
        dst := &net.IPAddr{IP: ip}
        msg := icmp.Message{
            Type: ipv4.ICMPTypeEcho, Code: 0,
            Body: &icmp.Echo{
                ID: os.Getpid() & 0xffff, Seq: 1,
                Data: []byte("ICMP TEST"),
            },
        }
        bytes, err := msg.Marshal(nil)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error marshalling ICMP: %v\n", err)
            os.Exit(1)
        }

        _, err = c.WriteTo(bytes, dst)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error sending ICMP: %v\n", err)
            os.Exit(1)
        }
        // Lecture de la réponse ICMP ici
        res := make([]byte, 1500)
        c.SetReadDeadline(time.Now().Add(time.Second))
        n, peer, err := c.ReadFrom(res)
		_ = peer
        if err != nil {
            if neterr, ok := err.(*net.OpError); ok && neterr.Timeout() {
                fmt.Printf("%s is down\n", ip)
                continue
            }
            fmt.Fprintf(os.Stderr, "Error reading ICMP: %v\n", err)
            os.Exit(1)
        }
        res = res[:n]
        rm, err := icmp.ParseMessage(1, res)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error parsing ICMP: %v\n", err)
            os.Exit(1)
        }
		if rm.Type == ipv4.ICMPTypeEchoReply {
            fmt.Printf("%s is up\n", ip)
        }
    }
}

func inc(ip net.IP) {
    for j := len(ip) - 1; j >= 0; j-- {
        ip[j]++
        if ip[j] > 0 {
            break
        }
    }
}