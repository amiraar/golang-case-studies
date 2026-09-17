package task

// Filter[T any] (A.66 generics): versi generik dari pola predicate yang
// dulu ditulis manual di dalam List() (store.go). Ditulis sekali, dipakai
// buat tipe apa saja - bukan cuma Task - tanpa duplikasi kode per tipe
// seperti sebelum Go 1.18.
func Filter[T any](items []T, pred func(T) bool) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		if pred(item) {
			out = append(out, item)
		}
	}
	return out
}

// Set[T comparable] (A.66 generics): kumpulan unik generik, dibangun di
// atas map[T]struct{} (struct{} dipakai karena tak makan memori, cuma
// butuh key-nya). "comparable" (bukan "any") karena isi map WAJIB bisa
// dibandingkan pakai == untuk deteksi duplikat.
type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{}, len(items))}
	for _, it := range items {
		s.Add(it)
	}
	return s
}

func (s *Set[T]) Add(item T)    { s.m[item] = struct{}{} }
func (s *Set[T]) Remove(item T) { delete(s.m, item) }
func (s *Set[T]) Has(item T) bool {
	_, ok := s.m[item]
	return ok
}
func (s *Set[T]) Len() int { return len(s.m) }

// Slice: urutannya tak dijamin (asal map), pemanggil yang butuh urutan
// tetap (mis. DuplicateTitles di parallel.go) wajib sort sendiri.
func (s *Set[T]) Slice() []T {
	out := make([]T, 0, len(s.m))
	for k := range s.m {
		out = append(out, k)
	}
	return out
}
