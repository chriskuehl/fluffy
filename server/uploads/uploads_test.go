package uploads

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetUniqueObjectID(t *testing.T) {
	got, err := genUniqueObjectID()
	if err != nil {
		t.Fatalf("genUniqueObjectID() error = %v", err)
	}
	wantLength := 32
	if len(got) != wantLength {
		t.Fatalf("got genUniqueObjectID() = %q, want len() = %d", got, wantLength)
	}
}

func TestExtractExtension(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{
			name: "no extension",
			in:   "file",
			want: "",
		},
		{
			name: "regular extension",
			in:   "file.txt",
			want: ".txt",
		},
		{
			name: "wrapped extension only",
			in:   "file.gz",
			want: ".gz",
		},
		{
			name: "wrapped extension after regular extension",
			in:   "file.tar.gz",
			want: ".tar.gz",
		},
		{
			name: "multiple wrapped extensions",
			in:   "file.tar.gz.bz2",
			want: ".tar.gz.bz2",
		},
		{
			name: "multiple wrapped extensions with a regular extension",
			in:   "file.txt.tar.gz.bz2",
			want: ".tar.gz.bz2",
		},
		{
			// Kind of nonsense, just making sure it doesn't remove more than it should.
			name: "wrapped extensions before regular extension",
			in:   "file.tar.gz.txt",
			want: ".txt",
		},
		{
			name: ". only",
			in:   ".",
			want: "",
		},
		{
			name: "multiple wrapped extensions with empty extensions",
			in:   "file.txt.tar.gz....bz2",
			want: ".tar.gz.bz2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractExtension(tt.in); got != tt.want {
				t.Errorf("got extractExtension(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSanitizeUploadName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *SanitizedKey
		wantErr error
	}{
		{
			name: "simple name",
			in:   "file.txt",
			want: &SanitizedKey{
				Extension: ".txt",
			},
		},
		{
			name: "dangerous name",
			in:   "path/to/../../etc/resolv.conf",
			want: &SanitizedKey{
				Extension: ".conf",
			},
		},
		{
			name: "dangerous extension",
			in:   "resolv.conf/../../etc/passwd",
			want: &SanitizedKey{
				Extension: "",
			},
		},
		{
			name: "dangerous name with windows path separators",
			in:   "path\\to\\..\\..\\etc\\resolv.conf",
			want: &SanitizedKey{
				Extension: ".conf",
			},
		},
		{
			name: "dangerous extension with windows path separators",
			in:   "resolv.conf\\..\\..\\etc/passwd",
			want: &SanitizedKey{
				Extension: "",
			},
		},
		{
			name: "empty name",
			in:   "",
			want: &SanitizedKey{
				Extension: "",
			},
		},
		{
			name: ".. only",
			in:   "..",
			want: &SanitizedKey{
				Extension: "",
			},
		},
		{
			name: ". only",
			in:   ".",
			want: &SanitizedKey{
				Extension: "",
			},
		},
		{
			name: "/ only",
			in:   "/",
			want: &SanitizedKey{
				Extension: "",
			},
		},
		{
			name: "/../ only",
			in:   "/../",
			want: &SanitizedKey{
				Extension: "",
			},
		},
		{
			name:    "forbidden extension",
			in:      "file.exe",
			wantErr: ErrForbiddenExtension,
		},
		{
			name:    "forbidden extension before wrapped extension",
			in:      "file.exe.gz",
			wantErr: ErrForbiddenExtension,
		},
		{
			name:    "forbidden extension before wrapped extension with ..",
			in:      "file.exe..gz",
			wantErr: ErrForbiddenExtension,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeUploadName(
				tt.in,
				map[string]struct{}{
					"exe": {},
				},
			)
			if err != tt.wantErr {
				t.Fatalf("sanitizeUploadName(%q) error = %v, want %v", tt.in, err, tt.wantErr)
			}
			if tt.want != nil && got != nil {
				tt.want.UniqueID = got.UniqueID
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("sanitizeUploadName(%q) mismatch (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}