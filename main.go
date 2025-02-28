package main

import (
	"EasierConnect/core"
	"flag"
	"fmt"
	"log"
)

func main() {
	// CLI args
	host, port, username, password, socksBind := "", 0, "", "", ""
	flag.StringVar(&host, "server", "", "EasyConnect server address (e.g. vpn.nju.edu.cn)")
	flag.StringVar(&username, "username", "", "Your username")
	flag.StringVar(&password, "password", "", "Your password")
	flag.StringVar(&socksBind, "socks-bind", ":1080", "The addr socks5 server listens on (e.g. 0.0.0.0:1080)")
	flag.IntVar(&port, "port", 443, "EasyConnect port address (e.g. 443)")
	debugDump := false
	flag.BoolVar(&debugDump, "debug-dump", false, "Enable traffic debug dump (only for debug usage)")
	flag.Parse()

	if host == "" || username == "" || password == "" {
		log.Fatal("Missing required cli args, refer to `EasierConnect --help`.")
	}
	server := fmt.Sprintf("%s:%d", host, port)

	// Web login part (Get TWFID & ECAgent Token => Final token used in binary stream)
	twfId := core.WebLogin(server, username, password)
	agentToken := core.ECAgentToken(server, twfId)
	token := (*[48]byte)([]byte(agentToken + twfId))

	// Query IP (keep the connection used so it's not closed too early, otherwise i/o stream will be closed)
	ip, conn := core.MustQueryIp(server, token)
	defer conn.Close()
	log.Printf("IP: %d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])

	// Link-level endpoint used in gvisor netstack
	endpoint := &core.EasyConnectEndpoint{}
	ipStack := core.SetupStack(ip, endpoint)

	// Sangfor Easyconnect protocol
	core.StartProtocol(endpoint, server, token, &[4]byte{ip[3], ip[2], ip[1], ip[0]}, debugDump)

	// Socks5 server
	core.ServeSocks5(ipStack, ip, socksBind)
}
