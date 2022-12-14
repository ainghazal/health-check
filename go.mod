module github.com/ainghazal/health-check

go 1.18

replace github.com/ooni/minivpn => ../minivpn

require (
	github.com/ainghazal/torii v0.0.0-20220825121826-5ec32c3fa0f7
	github.com/gorilla/mux v1.8.0
	github.com/ooni/minivpn v0.0.5
	github.com/procyon-projects/chrono v1.1.0
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/refraction-networking/utls v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/net v0.0.0-20220706163947-c90051bbdb60 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20211104114900-415007cec224 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20220703234212-c31a7b1ab478 // indirect
	golang.zx2c4.com/wireguard/tun/netstack v0.0.0-20220703234212-c31a7b1ab478 // indirect
	gvisor.dev/gvisor v0.0.0-20211020211948-f76a604701b6 // indirect
)
