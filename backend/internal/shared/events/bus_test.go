package events

import "testing"

func TestNormalizeSeeds(t *testing.T) {
	cases := []struct{ in, want string }{
		{"kafka:9092", "kafka:9092"},
		{"PLAINTEXT://kafka:9092", "kafka:9092"},
		{" PLAINTEXT://kafka:9092 ", "kafka:9092"},
		{" PLAINTEXT://kafka1:9092, kafka2:9092 ", "kafka1:9092,kafka2:9092"},
		{"", ""},
	}
	for _, c := range cases {
		got := normalizeSeeds(c.in)
		joined := ""
		for i, s := range got {
			if i > 0 {
				joined += ","
			}
			joined += s
		}
		if joined != c.want {
			t.Errorf("normalizeSeeds(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}