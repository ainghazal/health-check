package main

//
// Health checkers
//

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/ainghazal/torii/vpn"
	"github.com/ooni/minivpn/extras/ping"
	minivpn "github.com/ooni/minivpn/vpn"
)

const (
	vpnKind      = "vpn"
	openVPNProto = "openvpn"
	wgProto      = "wg"
)

var (
	ErrNotReady      = errors.New("not ready")
	ErrBadBase64Blob = errors.New("wrong base64 encoding")
)

// Checker is a VPN checker with aspirations of universality.
type Checker interface {
	Kind() string
	Proto() string
	ProviderName() string
	Run(chan *Measurement) error
}

// VPNChecker is the real thing.
type VPNChecker struct {
	Provider vpn.Provider
	proto    string
}

func (v *VPNChecker) Kind() string {
	return vpnKind
}

func (v *VPNChecker) Proto() string {
	return v.proto
}

func (v *VPNChecker) ProviderName() string {
	return v.Provider.Name()
}

func (v *VPNChecker) Run(mCh chan *Measurement) error {
	defer func() {
		close(mCh)
	}()

	if v.Provider == nil {
		return ErrNotReady
	}
	endpoints := v.Provider.Endpoints()
	log.Println("got", len(endpoints), "endpoints")

	// FIXME check for boundary
	// TODO split this monster function -------------------------
	for i := 0; i < maxEndpointsToTest; i++ {
		func() {
			m := &Measurement{
				Timestamp: time.Now(),
				Healthy:   false,
			}
			endp := endpoints[i]
			endpAddr := fmt.Sprintf("%s:%s", endp.IP, endp.Port)

			addr := net.TCPAddrFromAddrPort(netip.MustParseAddrPort(endpAddr))
			m.Addr = addr
			m.Transport = endp.Transport

			// TODO: refactor auth extraction
			auth := v.Provider.Auth()
			ca, _ := extractBase64Blob(auth.Ca)
			cert, _ := extractBase64Blob(auth.Cert)
			key, _ := extractBase64Blob(auth.Key)

			// this is inconvenient and bad ux. Options should
			// accept a string.
			optProto, err := protoToInt(endp.Transport)
			if err != nil {
				log.Println("ERROR:", err)
				return
			}

			opt := &minivpn.Options{
				Remote: endp.IP,
				Proto:  optProto,
				Port:   endp.Port,
				Cipher: "AES-256-GCM",
				Auth:   "SHA512",
				Ca:     ca,
				Cert:   cert,
				Key:    key,
			}

			log.Printf("Measuring endpoint: %d/%d", i+1, len(endpoints))

			// TODO split, get pinger + cancel

			ctx := context.Background()
			ctxDialTimeout, cancelDial := context.WithTimeout(ctx, time.Duration(time.Second*5))
			defer cancelDial()
			tunnel := minivpn.NewClientFromOptions(opt)

			// FIXME this is not being honored by client
			tunnel.Log = &nullLogger{}
			if err := tunnel.Start(ctxDialTimeout); err != nil {
				log.Println(err)
				m.Healthy = false
				mCh <- m
				return
			}
			pinger := ping.New("1.1.1.1", tunnel)
			pinger.Count = 3
			pinger.Silent = true

			ctxTimeout, cancelPing := context.WithTimeout(ctx, time.Duration(time.Second*4))
			defer cancelPing()

			if err := pinger.Run(ctxTimeout); err != nil {
				log.Println("err on ping", err)
				m.Healthy = false
				mCh <- m
				return
			}
			loss := pinger.Statistics().PacketLoss
			if loss < healthThresholdForPingLoss {
				m.Healthy = true
			}
			mCh <- m

		}()
	}
	return nil
}

var _ Checker = &VPNChecker{}

func singlePing(opt *minivpn.Options) *ping.Pinger {
	/*
	 ctx := context.Background()
	 ctxDialTimeout, cancelDial := context.WithTimeout(ctx, time.Duration(time.Second*5))
	 defer cancelDial()
	 tunnel := minivpn.NewClientFromOptions(opt)
	*/

	/*
	 // FIXME this is not being honored by client
	 tunnel.Log = &nullLogger{}
	 if err := tunnel.Start(ctxDialTimeout); err != nil {
	 	log.Println(err)
	 	m.Healthy = false
	 	mCh <- m
	 	return
	 }
	 pinger := ping.New("1.1.1.1", tunnel)
	 pinger.Count = 3
	 pinger.Silent = true
	*/

	/*
	 ctxTimeout, cancelPing := context.WithTimeout(ctx, time.Duration(time.Second*4))
	 defer cancelPing()
	*/
	return nil
}

func extractBase64Blob(val string) ([]byte, error) {
	s := strings.TrimPrefix(val, "base64:")
	if len(s) == len(val) {
		return nil, fmt.Errorf("%w: %s", ErrBadBase64Blob, "missing prefix")
	}
	dec, err := base64.URLEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrBadBase64Blob, err)
	}
	if len(dec) == 0 {
		return nil, nil
	}
	return dec, nil
}

// TODO this should be fixed in minivpn
// https://github.com/ooni/minivpn/issues/25
func protoToInt(p string) (int, error) {
	switch p {
	case "udp":
		return minivpn.UDPMode, nil
	case "tcp":
		return minivpn.TCPMode, nil
	default:
		return -1, fmt.Errorf("unknown proto: %s", p)
	}
}
