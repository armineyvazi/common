package uuid_test

import (
	"regexp"
	"testing"

	"github.com/armineyvazi/common.git/pkg/adapters/uuid"
)

var uuidV4RE = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

func TestGenV4_Format(t *testing.T) {
	gen := uuid.New()
	id, err := gen.GenV4()
	if err != nil {
		t.Fatalf("GenV4: %v", err)
	}
	if !uuidV4RE.MatchString(id) {
		t.Errorf("GenV4 returned non-UUIDv4 string: %q", id)
	}
}

func TestGenV4_Uniqueness(t *testing.T) {
	gen := uuid.New()
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id, err := gen.GenV4()
		if err != nil {
			t.Fatalf("GenV4 error on iteration %d: %v", i, err)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate UUID after %d iterations: %q", i, id)
		}
		seen[id] = struct{}{}
	}
}

func BenchmarkGenV4(b *testing.B) {
	gen := uuid.New()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = gen.GenV4()
	}
}

func BenchmarkGenV4_Parallel(b *testing.B) {
	gen := uuid.New()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = gen.GenV4()
		}
	})
}
