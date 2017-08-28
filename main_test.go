package main

import (
	"io/ioutil"
	"testing"

	"github.com/doctor-fate/mskix"
)

var configuration = defaultConfiguration()

func TestWriteSVG(t *testing.T) {
	var testcases = [...]string{"cisco", "extreme", "force10"}
	for _, v := range testcases {
		t.Run(v, func(t *testing.T) {
			data, err := ioutil.ReadFile("testdata/" + v + ".txt")
			if err != nil {
				t.Fatal(err)
			}

			parsed, err := mskix.Parse(data)
			if err != nil {
				t.Fatal(err)
			}

			if err := writeSVG(parsed, configuration); err != nil {
				t.Fatal(err)
			}
		})
	}
}
