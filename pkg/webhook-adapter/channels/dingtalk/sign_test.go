package dingtalk

import "testing"

func TestGenSign(t *testing.T) {
	tests := []struct {
		name        string
		secret      string
		timestampMs int64
		want        string
	}{
		{
			name:        "fixed vector",
			secret:      "my-secret",
			timestampMs: 1700000000000,
			want:        "/MY/oVQIff+ruG9eioj8sjGmwb0lCkY9/Pmw5vfsoM4=",
		},
		{
			name:        "fixed vector with different timestamp",
			secret:      "SEC0123456789abcdef",
			timestampMs: 1699999999123,
			want:        "ZTDwoIYgst8aXXI28QSTO1TyNr0IP03Dr/O9VVu3AIk=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenSign(tt.secret, tt.timestampMs)
			if err != nil {
				t.Fatalf("GenSign() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("GenSign() = %q, want %q", got, tt.want)
			}
		})
	}
}
