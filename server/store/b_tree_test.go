package store

import (
	"sync"
	"testing"

	"github.com/kdsama/kdb/logger"
)

func TestBTree(t *testing.T) {
	// Each function below corresponds to a function in b_tree.go
	// So for a function fo Setting keys there is a test function called TestSetKeys
	// and there is a benchmark function called BenchmarkSetKeys
}

func TestAddKeyString(t *testing.T) {
	Sample_Keys := []string{"a", "aa", "a*a", "b", "bc", "bcd", "a", "aa", "a*a", "b", "bc", "bcd", "a", "a"}
	degree := 4
	lg := logger.New(logger.Info)
	btree := newBTree(degree, lg)

	//setKey
	want := true
	t.Run("Testing Sequential Setting of Keys", func(t *testing.T) {

		for i := range Sample_Keys {

			got := btree.addKeyString(Sample_Keys[i])
			if got != want {
				t.Errorf("Wanted %v, but got %v", want, got)
			}
		}
	})
	t.Run("Parallel Setting of Keys", func(t *testing.T) {
		wg := sync.WaitGroup{}
		want := true
		results := make(chan bool, 6)
		for i := range Sample_Keys {

			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				result := btree.addKeyString(Sample_Keys[i])
				results <- result
			}(i)
		}

		go func() {
			wg.Wait()
			close(results)
		}()
		for res := range results {
			if !res {
				t.Errorf("Wanted %v but got %v", want, res)
			}
		}

	})

}

func BenchmarkAddKeyString(b *testing.B) {
	degree := 4
	btree := newBTree(degree, lg)
	Sample_Keys := []string{"a", "aa", "a*a", "b", "bc", "bcd", "a", "aa", "a*a", "b", "bc", "bcd", "a", "a"}
	for i := 0; i < b.N; i++ {
		// Perform the operation to be benchmarked

		// write operations are not safe for concurrent mutation by multiple go-routines
		// so need to implement locks for these.
		// probably rlock can be used .
		btree.addKeyString(Sample_Keys[i%len(Sample_Keys)])

		// Optional: Use the result to avoid compiler optimizations
		// _ = result
	}
}

func TestGetKey(t *testing.T) {
	Sample_Keys := []string{"a", "aa", "a*a", "b", "bc", "bcd", "a", "aa", "a*a", "b", "bc", "bcd", "a", "a"}
	degree := 4
	btree := newBTree(degree, lg)
	for i := range Sample_Keys {
		btree.addKeyString(Sample_Keys[i])
	}
	t.Run(" sequential reads on btree should return the key passed", func(t *testing.T) {

		for i := range Sample_Keys {
			want := Sample_Keys[i]
			got, ok := btree.getKeyString(Sample_Keys[i])
			if !ok {
				t.Errorf("wanted %v but not found", want)
			}
			if got != want {
				t.Errorf("wanted %v but got %v", want, got)
			}

		}
	})
	t.Run("key which is not present , should return false", func(t *testing.T) {
		want := false
		_, got := btree.getKeyString("Kshitij")
		if want != got {
			t.Errorf("wanted %v but got %v", want, got)
		}
	})

	t.Run("Concurrent reads on btree should return the key passed", func(t *testing.T) {
		wg := sync.WaitGroup{}
		for i := range Sample_Keys {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				want := Sample_Keys[i]
				got, ok := btree.getKeyString(Sample_Keys[i])
				if !ok {
					t.Errorf("wanted %v but not found", want)
				}
				if got != want {
					t.Errorf("wanted %v but got %v", want, got)
				}

			}(i)
		}

		wg.Wait()

	})
}

func BenchmarkGetKeyString(b *testing.B) {
	degree := 4
	btree := newBTree(degree, lg)
	Sample_Keys := []string{"a", "aa", "a*a", "b", "bc", "bcd", "a", "aa", "a*a", "b", "bc", "bcd", "a", "a"}
	for i := 0; i < len(Sample_Keys); i++ {
		btree.addKeyString(Sample_Keys[i])
	}

	b.Run("Testing Benchmarking for Sequential Read Access", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			want := Sample_Keys[i%len(Sample_Keys)]

			got, ok := btree.getKeyString(Sample_Keys[i%len(Sample_Keys)])
			if !ok {
				b.Errorf("wanted %v but not found", want)
			}
			if got != want {
				b.Errorf("wanted %v but got %v", want, got)
			}
		}
	})

	b.Run("Testing Benchmarking for Concurrent Read Access", func(b *testing.B) {
		wg := sync.WaitGroup{}
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				want := Sample_Keys[i%len(Sample_Keys)]

				got, ok := btree.getKeyString(Sample_Keys[i%len(Sample_Keys)])
				if !ok {
					b.Errorf("wanted %v but not found", want)
				}
				if got != want {
					b.Errorf("wanted %v but got %v", want, got)
				}
			}(i)
		}
		wg.Wait()

	})
}

func TestGetKeysFromPrefix(t *testing.T) {
	degree := 4
	btree := newBTree(degree, lg)
	Testing_Keys := []string{"a", "aa", "hamburgerville", "a*a", "baa", "hamburgeroni", "aabcd", "aaqwert", "sda", "123", "hamburger", "hamburgerStedam"}
	for i := range Testing_Keys {
		btree.addKeyString(Testing_Keys[i])
	}

	type prefixKeyStruct struct {
		prefix string
		want   []string
		name   string
	}
	testSuite := []prefixKeyStruct{
		{prefix: "a", want: []string{"a", "aa", "a*a", "aabcd", "aaqwert"}, name: "Testing for small prefix :: A"},
		{prefix: "hamburger", want: []string{"hamburgerville", "hamburgeroni", "hamburger", "hamburgerStedam"}, name: "Testing for Larger prefix :: hamburger"},
		{prefix: "family", want: []string{}, name: "Testing for no results"},
	}
	for _, obj := range testSuite {
		t.Run(obj.name, func(t *testing.T) {
			want := obj.want
			got := btree.getKeysFromPrefix(obj.prefix)
			if len(want) != len(got) {
				t.Errorf("The expected size of the result was %v, but got %v", len(want), len(got))
			}
			checkMap := map[string]bool{}
			for i := range want {
				checkMap[want[i]] = true
			}
			for j := range got {
				delete(checkMap, got[j])
			}
			if len(checkMap) != 0 {
				t.Errorf("The size of ")
			}
		})
	}
}

func BenchmarkGetKeysFromPrefix(b *testing.B) {
	degree := 4
	btree := newBTree(degree, lg)
	Testing_Keys := []string{"a", "aa", "hamburgerville", "a*a", "baa", "hamburgeroni", "aabcd", "aaqwert", "sda", "123", "hamburger", "hamburgerStedam"}
	for i := range Testing_Keys {
		btree.addKeyString(Testing_Keys[i])
	}
	type prefixKeyStruct struct {
		prefix string
		want   []string
		name   string
	}
	testSuite := []prefixKeyStruct{
		{prefix: "a", want: []string{"a", "aa", "a*a", "aabcd", "aaqwert"}, name: "Testing for small prefix :: A"},
		{prefix: "hamburger", want: []string{"hamburgerville", "hamburgeroni", "hamburger", "hamburgerStedam"}, name: "Testing for Larger prefix :: hamburger"},
		{prefix: "family", want: []string{}, name: "Testing for no results"},
	}
	b.Run("Sequential Benchmarking", func(b *testing.B) {

		for _, obj := range testSuite {

			b.Run(obj.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					want := obj.want
					got := btree.getKeysFromPrefix(obj.prefix)
					if len(want) != len(got) {
						b.Errorf("The expected size of the result was %v, but got %v", len(want), len(got))
					}
					checkMap := map[string]bool{}
					for i := range want {
						checkMap[want[i]] = true
					}
					for j := range got {
						delete(checkMap, got[j])
					}
					if len(checkMap) != 0 {
						b.Error("The size of want and got doesnot match ")
					}
				}
			})

		}
	})

	b.Run("Concurrency Benchmarking  for Prefix Reads", func(b *testing.B) {
		for _, obj := range testSuite {

			b.Run(obj.name, func(b *testing.B) {
				wg := sync.WaitGroup{}
				for k := 0; k < b.N; k++ {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						want := obj.want
						got := btree.getKeysFromPrefix(obj.prefix)
						if len(want) != len(got) {
							b.Errorf("The expected size of the result was %v, but got %v", len(want), len(got))
						}
						checkMap := map[string]bool{}
						for i := range want {
							checkMap[want[i]] = true
						}
						for j := range got {
							delete(checkMap, got[j])
						}
						if len(checkMap) != 0 {
							b.Error("The size of want and got doesnot match ")
						}
					}(k)
				}
				wg.Wait()
			})
		}
	})
}
