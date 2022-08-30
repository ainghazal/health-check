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
	maxEndpointsToTest = 2   // DEBUG, change-me for release
	healthThreshold    = 0.8 // below this loss rate, it's healthy

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

// TODO move to minivpn/ extras -------------------------
type nullLogger struct{}

func (n *nullLogger) Info(string) {}

func (n *nullLogger) Infof(string, ...interface{}) {}

func (n *nullLogger) Debug(string) {}

func (n *nullLogger) Debugf(string, ...interface{}) {}

func (n *nullLogger) Warn(string) {}

func (n *nullLogger) Warnf(string, ...interface{}) {}

func (n *nullLogger) Error(string) {}

func (n *nullLogger) Errorf(string, ...interface{}) {}

// -----------------------------------------------------

func (v *VPNChecker) Run(mCh chan *Measurement) error {
	defer func() {
		close(mCh)
	}()

	if v.Provider == nil {
		return ErrNotReady
	}
	endpoints := v.Provider.Endpoints()
	log.Println("got", len(endpoints), "endpoints")

	var m *Measurement

	// FIXME check for boundary
	// TODO split this monster function -------------------------

	for i := 0; i < maxEndpointsToTest; i++ {
		func() {
			m = &Measurement{}
			endp := endpoints[i]
			endpAddr := fmt.Sprintf("%s:%s", endp.IP, endp.Port)

			addr := net.TCPAddrFromAddrPort(netip.MustParseAddrPort(endpAddr))
			m.Addr = addr
			auth := v.Provider.Auth()

			ca, _ := extractBase64Blob(auth.Ca)
			cert, _ := extractBase64Blob(auth.Cert)
			key, _ := extractBase64Blob(auth.Key)

			opt := &minivpn.Options{
				Remote: endp.IP,
				Port:   endp.Port,
				Cipher: "AES-256-GCM",
				Auth:   "SHA512",
				Ca:     ca,
				Cert:   cert,
				Key:    key,
			}
			ctx := context.Background()
			tunnel := minivpn.NewClientFromOptions(opt)

			// FIXME this is not being honored by client
			tunnel.Log = &nullLogger{}
			err := tunnel.Start(ctx)
			if err != nil {
				log.Println(err)
				m.Healthy = false
				mCh <- m
				return
			}
			pinger := ping.New("1.1.1.1", tunnel)
			pinger.Count = 5
			pinger.Silent = true

			ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(time.Second*7))
			defer cancel()
			err = pinger.Run(ctxTimeout)
			if err != nil {
				log.Println("err on ping", err)
				m.Healthy = false
				mCh <- m
				return
			}
			pinger.PrintStats()
			// TODO actually decide based on the threshold here!!!
			m.Healthy = true
			mCh <- m
		}()
	}
	return nil
}

var _ Checker = &VPNChecker{}

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
