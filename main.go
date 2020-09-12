/*

Copyright © 2020 Andrey Kuvshinov. Contacts: <syslinux@protonmail.com>
Copyright © 2020 Eltaline OU. Contacts: <eltaline.ou@gmail.com>
All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

The KnockD project contains unmodified/modified libraries imports too with
separate copyright notices and license terms. Your use of the source code
this libraries is subject to the terms and conditions of licenses these libraries.

*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {

	var version string = "1.0.0"
	var vprint bool = false

	var help bool = false

	fport := "47566"
	sport := "7566"

	tcpports := ""
	udpports := ""

	timeout := 60
	ttl := 86400

	flag.StringVar(&fport, "fport", fport, "--fport=47566")
	flag.StringVar(&sport, "sport", sport, "--sport=7566")

	flag.StringVar(&tcpports, "tcpports", tcpports, "--tcpports=21,22,25")
	flag.StringVar(&udpports, "udpports", udpports, "--udpports=135,137")

	flag.IntVar(&timeout, "timeout", timeout, "--timeout=60")
	flag.IntVar(&ttl, "ttl", ttl, "--ttl=86400")

	flag.BoolVar(&vprint, "version", vprint, "--version - print version")
	flag.BoolVar(&help, "help", help, "--help")

	flag.Parse()

	switch {
	case vprint:
		fmt.Printf("KnockD Version: %s\n", version)
		os.Exit(0)
	case help:
		flag.PrintDefaults()
		os.Exit(0)
	}

	InterruptHandler()

	kports := fport + "," + sport

	command := exec.Command("iptables", "-X", "KNOCKD")
	_ = command.Run()
	command = exec.Command("iptables", "-N", "KNOCKD")
	_ = command.Run()
	command = exec.Command("iptables", "-F", "KNOCKD")
	_ = command.Run()
	command = exec.Command("iptables", "-D", "INPUT", "-j", "KNOCKD")
	_ = command.Run()
	command = exec.Command("iptables", "-A", "INPUT", "-j", "KNOCKD")
	_ = command.Run()
	command = exec.Command("iptables", "-I", "KNOCKD", "-p", "all", "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT")
	_ = command.Run()
	command = exec.Command("iptables", "-I", "KNOCKD", "-p", "tcp", "-s", "0.0.0.0/0", "-m", "multiport", "--dports", kports, "-j", "ACCEPT")
	_ = command.Run()

	if tcpports != "" {

		command = exec.Command("iptables", "-A", "KNOCKD", "-p", "tcp", "-s", "127.0.0.0/8", "-m", "multiport", "--dports", tcpports, "-j", "ACCEPT")
		_ = command.Run()
		command = exec.Command("iptables", "-A", "KNOCKD", "-p", "tcp", "-s", "192.168.0.0/16", "-m", "multiport", "--dports", tcpports, "-j", "ACCEPT")
		_ = command.Run()
		command = exec.Command("iptables", "-A", "KNOCKD", "-p", "tcp", "-s", "172.16.0.0/12", "-m", "multiport", "--dports", tcpports, "-j", "ACCEPT")
		_ = command.Run()
		command = exec.Command("iptables", "-A", "KNOCKD", "-p", "tcp", "-s", "10.0.0.0/8", "-m", "multiport", "--dports", tcpports, "-j", "ACCEPT")
		_ = command.Run()
		command := exec.Command("iptables", "-A", "KNOCKD", "-p", "tcp", "-s", "0.0.0.0/0", "-m", "multiport", "--dports", tcpports, "-j", "DROP")
		_ = command.Run()

	}

	if udpports != "" {

		command = exec.Command("iptables", "-A", "KNOCKD", "-p", "udp", "-s", "127.0.0.0/8", "-m", "multiport", "--dports", udpports, "-j", "ACCEPT")
		_ = command.Run()
		command = exec.Command("iptables", "-A", "KNOCKD", "-p", "udp", "-s", "192.168.0.0/16", "-m", "multiport", "--dports", udpports, "-j", "ACCEPT")
		_ = command.Run()
		command = exec.Command("iptables", "-A", "KNOCKD", "-p", "udp", "-s", "172.16.0.0/12", "-m", "multiport", "--dports", udpports, "-j", "ACCEPT")
		_ = command.Run()
		command = exec.Command("iptables", "-A", "KNOCKD", "-p", "udp", "-s", "10.0.0.0/8", "-m", "multiport", "--dports", udpports, "-j", "ACCEPT")
		_ = command.Run()
		command := exec.Command("iptables", "-A", "KNOCKD", "-p", "udp", "-s", "0.0.0.0/0", "-m", "multiport", "--dports", udpports, "-j", "DROP")
		_ = command.Run()

	}

	fconn, err := net.ListenPacket("udp4", ":"+fport)
	if err != nil {
		log.Fatal(err)
	}
	defer fconn.Close()

	sconn, err := net.ListenPacket("udp4", ":"+sport)
	if err != nil {
		log.Fatal(err)
	}
	defer sconn.Close()

	lock := make(map[string]bool)
	pair := make(map[string]uint64)

	go func() {

		ctime := uint64(time.Now().Unix())

		for sip, stime := range pair {

			diff := int(ctime - stime)

			if diff > ttl {

				found := lock[sip]

				if found {

					command := exec.Command("iptables", "-D", "KNOCKD", "-p", "all", "-s", sip, "-j", "ACCEPT")
					_ = command.Run()

					delete(lock, sip)

				}

				delete(pair, sip)

			}

		}

		time.Sleep(5 * time.Second)

	}()

	go func() {

		for {

			var msg []byte

			_, ip, err := fconn.ReadFrom(msg)
			if err != nil {
				continue
			}

			sip := strings.Split(fmt.Sprintf("%s", ip), ":")[0]

			if net.ParseIP(sip) != nil {
				pair[sip] = uint64(time.Now().Unix())
			}

		}

	}()

	for {

		var msg []byte

		_, ip, err := sconn.ReadFrom(msg)
		if err != nil {
			continue
		}

		sip := strings.Split(fmt.Sprintf("%s", ip), ":")[0]

		if net.ParseIP(sip) != nil {

			ctime := uint64(time.Now().Unix())
			stime := pair[sip]
			diff := int(ctime - stime)

			found := lock[sip]

			if diff < timeout && !found {

				var estdout bytes.Buffer

				ecommand := exec.Command("iptables", "-nL", "KNOCKD", "|", "grep", "ACCEPT", "|", "grep", sip)
				ecommand.Stdout = &estdout
				_ = ecommand.Run()

				if string(estdout.Bytes()) == "" {

					command := exec.Command("iptables", "-I", "KNOCKD", "-p", "all", "-s", sip, "-j", "ACCEPT")
					_ = command.Run()

					lock[sip] = true

				}

			}

		}

	}

}

// InterruptHandler: remove iptables rules after stop knockd
func InterruptHandler() {

	chos := make(chan os.Signal)
	signal.Notify(chos, os.Interrupt, syscall.SIGTERM)

	go func() {

		<-chos

		time.Sleep(250 * time.Millisecond)

		command := exec.Command("iptables", "-D", "INPUT", "-j", "KNOCKD")
		_ = command.Run()

		command = exec.Command("iptables", "-F", "KNOCKD")
		_ = command.Run()

		command = exec.Command("iptables", "-X", "KNOCKD")
		_ = command.Run()

		os.Exit(0)

	}()

}
