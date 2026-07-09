package household

import "testing"

func TestCreateRequestValidate(t *testing.T) {
	cases := []struct {
		name    string
		req     CreateRequest
		wantErr bool
	}{
		{"valid", CreateRequest{OwnerName: "Budi", Address: "Jl. Melati 1"}, false},
		{"missing owner", CreateRequest{Address: "Jl. Melati 1"}, true},
		{"missing address", CreateRequest{OwnerName: "Budi"}, true},
		{"whitespace owner", CreateRequest{OwnerName: "   ", Address: "x"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.req.Validate()
			if (err != nil) != c.wantErr {
				t.Fatalf("Validate() err = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}
