package nameparser_test

import (
	"fmt"

	"github.com/ysmood/nameparser"
)

func ExampleNew() {
	hn := nameparser.New("Doe, Dr. John A. Kenneth, Jr.")

	fmt.Println(hn.Title())
	fmt.Println(hn.First())
	fmt.Println(hn.Middle())
	fmt.Println(hn.Last())
	fmt.Println(hn.Suffix())

	// Output:
	// Dr.
	// John
	// A. Kenneth
	// Doe
	// Jr.
}

func ExampleHumanName_Initials() {
	hn, err := nameparser.NewWithOptions(
		"Doe, John A. Kenneth, Jr.",
		nameparser.HumanNameOptions{InitialsDelimiter: ";"},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(hn.Initials())

	// Output:
	// J; A; K; D;
}

func ExampleHumanName_String() {
	hn := nameparser.New("Rev John A. Kenneth Doe III (Kenny)")
	hn.StringFormat = "{last}, {first} {middle}"

	fmt.Println(hn.String())

	// Output:
	// Doe, John A. Kenneth
}

func ExampleNewWithConstants() {
	c := nameparser.NewConstants()
	c.Titles.Add("dean", "chemistry")

	hn := nameparser.NewWithConstants("Assoc Dean of Chemistry Robert Johns", c)

	fmt.Println(hn.Title())
	fmt.Println(hn.First())
	fmt.Println(hn.Last())

	// Output:
	// Assoc Dean of Chemistry
	// Robert
	// Johns
}

func ExampleNewWithOptions() {
	hn, err := nameparser.NewWithOptions("", nameparser.HumanNameOptions{
		Title: "Countess",
		First: "Ada",
		Last:  "Lovelace",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(hn.String())

	// Output:
	// Countess Ada Lovelace
}
