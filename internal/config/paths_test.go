package config

import "testing"

func TestResolvePaths(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Paths
		wantErr bool
	}{
		{
			name: "uses XDG vars when set",
			env: map[string]string{
				"HOME":            "/home/u",
				"XDG_CONFIG_HOME": "/xdg/config",
				"XDG_DATA_HOME":   "/xdg/data",
				"XDG_STATE_HOME":  "/xdg/state",
			},
			want: Paths{
				Config: "/xdg/config/mimir",
				Data:   "/xdg/data/mimir",
				State:  "/xdg/state/mimir",
			},
		},
		{
			name: "falls back to HOME defaults",
			env:  map[string]string{"HOME": "/home/u"},
			want: Paths{
				Config: "/home/u/.config/mimir",
				Data:   "/home/u/.local/share/mimir",
				State:  "/home/u/.local/state/mimir",
			},
		},
		{
			// The XDG spec says relative paths must be ignored.
			name: "ignores relative XDG paths",
			env:  map[string]string{"HOME": "/home/u", "XDG_CONFIG_HOME": "relative/dir"},
			want: Paths{
				Config: "/home/u/.config/mimir",
				Data:   "/home/u/.local/share/mimir",
				State:  "/home/u/.local/state/mimir",
			},
		},
		{
			name:    "errors when HOME is unset and no XDG fallback exists",
			env:     map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ResolvePaths(func(k string) string { return tt.env[k] })
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
