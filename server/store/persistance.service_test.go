package store

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"testing"
)

var prefix = "../../data/testing/"

func TestPersistanceSave(t *testing.T) {
	// We will use fs fileService to check if the save Happened correctly

	t.Run("Test Persistance Save Sequentially", func(t *testing.T) {
		fs := NewFileService()

		key := "test"

		want := "test_data"
		per := NewPersistance(prefix)
		err := per.Save(key, bytes.NewBuffer([]byte(want)))
		defer os.Remove(prefix + want).Error()
		if err != nil {
			t.Error("Did not expect error on save . Err::::", err)
		}
		got, err := fs.ReadLatestFromFile(prefix + key)
		if err != nil {
			t.Fatal("Issue in File Service, Fix it ASAP", err)
		}
		if got != want {
			t.Errorf("wanted %v, but got %v for prefix+file %v", want, got, prefix+key)
		}
		os.RemoveAll(prefix)
	})

	t.Run("Test Persistance Save for Parallel Scenarios", func(t *testing.T) {

		key := "test"

		want := "test_data"
		per := NewPersistance(prefix)

		ws := sync.WaitGroup{}
		for i := 0; i < 100; i++ {
			ws.Add(1)
			go func() {
				defer ws.Done()
				err := per.Save(key, bytes.NewBuffer([]byte(want)))
				if err != nil {
					t.Error("Did not expect error on save . Err::::", err)
				}
			}()
		}
		ws.Wait()
		os.RemoveAll(prefix)
	})

}
func BenchmarkPersistanceSave(b *testing.B) {
	b.Run("Sequential multi magnitude writes for different locations", func(b *testing.B) {

		data := `data_in_bytes := bytes.NewBuffer([]byte(data))
		per := NewPersistance(prefix)
		// here we would like to have some sub directories as well as well
		ws := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			a := "/someTest"
			b := "/someTestAgain"
			c := "/maybeOtherTest"
			ws.Add(2)
			go func(i int) {
				defer ws.Done()
				per.Save(a+fmt.Sprint(i)+b+c, data_in_bytes)
			}(i)
			go func(i int) {
				defer ws.Done()
				per.Save(a+b+fmt.Sprint(i)+c, data_in_bytes)
			}(i)

		}
		ws.Wait()
		// os.RemoveAll(prefix)
	})
}
data_in_bytes := bytes.NewBuffer([]byte(data))
		per := NewPersistance(prefix)
		// here we would like to have some sub directories as well as well
		ws := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			a := "/someTest"
			b := "/someTestAgain"
			c := "/maybeOtherTest"
			ws.Add(2)
			go func(i int) {
				defer ws.Done()
				per.Save(a+fmt.Sprint(i)+b+c, data_in_bytes)
			}(i)
			go func(i int) {
				defer ws.Done()
				per.Save(a+b+fmt.Sprint(i)+c, data_in_bytes)
			}(i)

		}
		ws.Wait()
		// os.RemoveAll(prefix)
	})
}

`
		data_in_bytes := bytes.NewBuffer([]byte(data))
		per := NewPersistance(prefix)
		os.RemoveAll(prefix)
		// here we would like to have some sub directories as well as well
		ws := sync.WaitGroup{}
		for i := 0; i < b.N; i++ {
			a := "/someTest"
			b := "/someTestAgain"
			c := "/maybeOtherTest"
			ws.Add(2)
			go func() {
				i := i
				defer ws.Done()
				per.Save(a+fmt.Sprint(i)+b+c, data_in_bytes)
			}()
			go func() {
				i := i
				defer ws.Done()
				per.Save(a+b+fmt.Sprint(i)+c, data_in_bytes)
			}()

		}
		ws.Wait()
		os.RemoveAll(prefix)
	})
}
