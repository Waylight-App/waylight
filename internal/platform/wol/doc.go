// Package wol sends Wake-on-LAN magic packets, either directly as a UDP
// broadcast or through the relay, and implements that relay
// ("waylight wol-relay"). The relay runs in the host network so packets
// reach the LAN while Waylight itself stays in its own Docker network.
//
// Ring: Frameworks & Drivers.
package wol
