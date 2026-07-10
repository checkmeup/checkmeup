package httpsafe

import "testing"

func TestRefuseRedirects(t *testing.T) {
	if err := RefuseRedirects(nil, nil); err == nil {
		t.Fatal("want RefuseRedirects to always return an error")
	}
}

func TestBlockPrivateDial(t *testing.T) {
	cases := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"loopback", "127.0.0.1:443", true},
		{"private", "192.168.1.1:443", true},
		{"link-local / cloud metadata", "169.254.169.254:443", true},
		{"unspecified", "0.0.0.0:443", true},
		{"multicast", "224.0.0.1:443", true},
		{"public", "93.184.216.34:443", false},
		{"non-IP host (already resolved by the dialer upstream)", "example.com:443", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := BlockPrivateDial("tcp", tc.address, nil)
			if tc.wantErr && err == nil {
				t.Fatalf("want an error for address %q", tc.address)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("want no error for address %q, got %v", tc.address, err)
			}
		})
	}
}
