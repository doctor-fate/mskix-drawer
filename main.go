package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/doctor-fate/mskix"
	"github.com/doctor-fate/mskix/device"
	"github.com/doctor-fate/src/mskix-drawer/templates"
	"github.com/doctor-fate/src/mskix-parsers"
)

func BuildSVG(w io.Writer, data device.Data, configuration templates.Configuration) {
	templates.WriteHeader(w, configuration)
	templates.WriteTitle(w, data.Id)

	type Anchor uint8

	const (
		Right Anchor = iota
		Left
	)

	const (
		DefaultHorizontalOffset = 400
		DefaultVerticalOffset   = 125
	)

	var (
		translate = templates.Translate{
			Horizontal: configuration.Width - DefaultHorizontalOffset,
			Vertical:   DefaultVerticalOffset,
		}
		anchor = Right
	)

	for _, record := range data.Records {
		switch anchor {
		case Left:
			templates.WriteRecordLeft(w, record, translate, configuration)
		case Right:
			templates.WriteRecordRight(w, record, translate, configuration)
		}

		translate.Vertical += configuration.RC.Height
		if translate.Vertical > configuration.Height-50 {
			translate = templates.Translate{
				Horizontal: DefaultHorizontalOffset,
				Vertical:   DefaultVerticalOffset,
			}
			anchor = Left
		}
	}

	templates.WriteFooter(w)
}

func main() {
	parsers := mskix.Parsers()
	fmt.Println(parsers)
	data, _ := ioutil.ReadFile("cisco.txt")
	parsed, err := mskix.Parse(mskix_parsers.Cisco, data)
	fmt.Println(parsed)

	out, err := os.Create("cisco.svg")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	var configuration = templates.Configuration{
		Width: 1024, Height: 1024,
		RC: templates.RectangleConfiguration{
			Width: 80, Height: 20, Style: "stroke:black;stroke-width:1.5;fill:none",
		},
		TC: templates.TextConfiguration{
			FontSize: 11,
		},
		AC: templates.ArrowConfiguration{
			HorizontalLength: 300, HorizontalShift: 10, VerticalShift: 5, Style: "stroke:black;stroke-width:1.5",
		},
	}
	BuildSVG(out, parsed, configuration)
}
