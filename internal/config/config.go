package config

import (
	"flag"
	"os"
	"strings"
) 

type Config struct { 
    Port int; // server port which the server will listen on
    Peers []string; // list of peer addresses host:port
    Name string; // name of the node
}

func Parse() Config {

    port := flag.Int("port", 9000, "Port to listen on")
    peers := flag.String("peers", "", "Comma separated peer list")
    name := flag.String("name", "", "Your chat name")

    flag.Parse()
    if *name == "" {
       flag.Usage();
        os.Exit(1);
    }

    return Config {
        Port: *port,
        Peers: SplitPeers(*peers),
        Name: *name,
    }

}

func SplitPeers(peers string) []string {
    if peers == "" {
        return nil
    }

    return strings.Split(peers, ",");
}
