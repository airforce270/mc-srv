package packet

import (
	"bytes"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWriteStatusResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc  string
		input int
		want  []byte
	}{
		{
			desc:  "standard",
			input: 765,
			want: slices.Concat(
				// header
				[]byte{0x83, 0x4a, 0x00},
				// payload
				[]byte{0x80, 0x4a},
				[]byte(`{"version":{"name":"1.20.4","protocol":765},"players":{"max":34,"online":12,"sample":null},"description":{"text":"The Minecraft client-server protocol kinda sucks ngl"},"favicon":"`+iconDataURI+`","enforcesSecureChat":false,"previewsChat":false}`),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			sr, err := NewStatusResponse(tc.input)
			if err != nil {
				t.Fatalf("WriteStatusResponse unexpected err: %v", err)
			}

			var out bytes.Buffer

			if err := sr.Write(&out); err != nil {
				t.Fatalf("WriteStatusResponse() unexpected err: %v", err)
			}

			got := out.Bytes()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("WriteStatusResponse() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
