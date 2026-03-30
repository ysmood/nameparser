package nameparser_test

import (
	"sync"
	"testing"

	"github.com/ysmood/nameparser"
)

func TestConcurrentParsingDoesNotMutateSharedConstants(t *testing.T) {
	inputs := []string{
		"dr.mr John Doe",
		"John and of Doe",
		"de la Cruz, Maria",
		"Doe, Dr. John A. Kenneth, Jr.",
	}

	beforeTitles := nameparser.CONSTANTS.Titles.Len()
	beforeConjunctions := nameparser.CONSTANTS.Conjunctions.Len()
	beforeSuffixes := nameparser.CONSTANTS.SuffixNotAcronyms.Len()

	const workers = 8
	const iterations = 1200

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(offset int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = nameparser.New(inputs[(i+offset)%len(inputs)])
			}
		}(w)
	}
	wg.Wait()

	if got := nameparser.CONSTANTS.Titles.Len(); got != beforeTitles {
		t.Fatalf("titles set mutated by parser: got %d want %d", got, beforeTitles)
	}
	if got := nameparser.CONSTANTS.Conjunctions.Len(); got != beforeConjunctions {
		t.Fatalf("conjunction set mutated by parser: got %d want %d", got, beforeConjunctions)
	}
	if got := nameparser.CONSTANTS.SuffixNotAcronyms.Len(); got != beforeSuffixes {
		t.Fatalf("suffix set mutated by parser: got %d want %d", got, beforeSuffixes)
	}
	if nameparser.CONSTANTS.Titles.Contains("dr.mr") {
		t.Fatalf("unexpected inferred title leaked into global constants")
	}
	if nameparser.CONSTANTS.Conjunctions.Contains("and of") {
		t.Fatalf("unexpected inferred conjunction leaked into global constants")
	}
}

func TestConcurrentStringSetOperations(t *testing.T) {
	s := nameparser.NewStringSet([]string{"alpha"})

	const workers = 6
	const iterations = 3000

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				key := "k" + string(rune('a'+(i+id)%26))
				s.Add(key)
				s.Contains(key)
				if i%3 == 0 {
					s.Remove(key)
				}
				_ = s.Len()
				_ = s.Values()
			}
		}(w)
	}
	wg.Wait()
}
