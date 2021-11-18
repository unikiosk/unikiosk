// +build !linux

package transocks

import "net"

// GetOriginalDST retrieves the original destination address from
// NATed connection.  Currently, only Linux iptables using DNAT/REDIRECT
// is supported.  For other operating systems, this will just return
// conn.LocalAddr().
//
// Note that this function only works when nf_conntrack_ipv4 and/or
// nf_conntrack_ipv6 is loaded in the kernel.
func GetOriginalDST(conn *net.TCPConn) (*net.TCPAddr, error) {
	return conn.LocalAddr().(*net.TCPAddr), nil
}
