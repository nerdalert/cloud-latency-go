package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
	"github.com/tatsushid/go-fastping"
)

type Config struct {
	TestDuration string    `yaml:"test-length"`
	TestInterval string    `yaml:"test-interval"`
	ServerPort   string    `yaml:"server-port"`
	TsdbServer   string    `yaml:"grafana-address"`
	TsdbPort     string    `yaml:"grafana-port"`
	TsdbPrefix     string    `yaml:"tsdb-prefix"`
	Entry        []Servers `yaml:"target-hosts"`
}

type Servers map[string]string

type Endpoint struct {
	ServerIP   string
	Port       string
	ServerName string
}

type Cli struct {
	Debug      bool
	ConfigPath string
	Help       bool
}

var cli *Cli
var iperfImg = "networkstatic/iperf3"
var log = logrus.New()

func SetLogger(l *logrus.Logger) {
	log = l
}

func init() {
	const (
		debugFlag     = false
		debugDescrip  = "Run in debug mode to display all shell commands being executed"
		configPath    = "./config.yml"
		configDescrip = "Path to the configuration file -config=path/config.yml"
		helpFlag      = false
		helpDescrip   = "Print Usage Options"
	)
	cli = &Cli{}
	flag.BoolVar(&cli.Debug, "debug", debugFlag, debugDescrip)
	flag.StringVar(&cli.ConfigPath, "config", configPath, configDescrip)
	flag.BoolVar(&cli.Help, "help", helpFlag, helpDescrip)
}

func main() {
	flag.Parse()
	if cli.Help {
		flag.PrintDefaults()
		os.Exit(1)
	}
	for {
		// Read in the yaml configuration from config.yaml
		data, err := ioutil.ReadFile(cli.ConfigPath)
		if err != nil {
			log.Fatalln("There was a problem opening the configuration file. Make sure "+
				"'config.yml' is located in the same directory as the binary 'cbandwidth' or set"+
				" the location using -config=path/config.yml || Error: ", err)
		}
		config := Config{}
		if err := yaml.Unmarshal([]byte(data), &config); err != nil {
			log.Fatal(err)
		}
		graphiteSocket := net.JoinHostPort(config.TsdbServer, config.TsdbPort)
		for _, val := range config.Entry {
			for targetAddress, targetName := range val {

				log.Infof("FUK Namw --> %s", targetName)


				if targetName == "" {
					targetName = targetAddress
				}

				// FUK TODO Find replace . with -
				log.Infof("FUK2 Namw --> %s", targetName)




				rttResult :=  pingIPv4Probe(targetAddress)

				if rttResult == 0 {
					log.Errorf("Error probing target host at %s", targetAddress)
				} else {
					// Write the download results to the tsdb
					log.Infof("Latency results for endpoint %s -> %vms", targetAddress, rttResult)
					timeDownNow := time.Now().Unix()
					sendGraphite("tcp", graphiteSocket, fmt.Sprintf("%s.%s %v %d\n",
						config.TsdbPrefix, targetName, rttResult, timeDownNow))
				}
			}
		}
		// Polling interval as defined in the config file. The default is 5 minutes.
		t, _ := time.ParseDuration(string(config.TestInterval) + "s")
		time.Sleep(t)
	}
}

// Run the iperf container and return the output and any errors
func runCmd(command string, cli *Cli) (string, error) {
	command = strings.TrimSpace(command)
	var cmd string
	var args []string
	cmd = "/bin/bash"
	args = []string{"-c", command}
	// log the shell command being run to stdout if the debug flag is set
	if cli.Debug {
		log.Infoln("Running shell command -> ", args)
	}
	output, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// Write the results to a graphite socket
func sendGraphite(connType string, socket string, msg string) {
	//conn, err := net.Dial(connType, *socket)
	conn, err := net.Dial(connType, socket)
	if err != nil {
		log.Errorf("Could not connect to the graphite server -> %s", socket)
		log.Errorf("Verify the graphite server is running and reachable at %s", socket)
	} else {
		defer conn.Close()
		_, err = fmt.Fprintf(conn, msg)
		if err != nil {
			log.Errorf("Error writing to the graphite server at -> %s", socket)
		}
	}
}

//rttResult =  pingIPv4Probe("mirror.waia.asn.au")
//fmt.Println("Australie Example RTT -> ", rttResult)


func pingIPv4Probe(arg string) float64 {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", arg)
	if err != nil {
		log.Error(err)
		return 0
	}
	avrRTT := pingTarget(ra, err, p)
	result := float64(avrRTT) / float64(time.Millisecond)
	return result
}

func pingTarget(target *net.IPAddr, err error, p *fastping.Pinger) time.Duration {
	// Ping Probe which returns the RTT in ms as float64
	var avgRTT time.Duration
	// send 3
	p.AddIPAddr(target)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		avgRTT = rtt
	}
	err = p.Run()
	if err != nil {
		log.Error(err)
		return 0
	}
	return avgRTT
}
