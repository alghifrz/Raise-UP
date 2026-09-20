package whatsapp

import "testing"

func TestNormalizePhone(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"081234567890", "6281234567890"},
		{"+62 812-3456-7890", "6281234567890"},
		{"6281234567890", "6281234567890"},
		{"81234567890", "6281234567890"},
	}
	for _, tc := range cases {
		got, err := NormalizePhone(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizePhoneInvalid(t *testing.T) {
	if _, err := NormalizePhone("abc"); err == nil {
		t.Fatal("expected error")
	}
}
