package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/doctor-fate/mskix"
	"github.com/doctor-fate/mskix-drawer/templates"
	"github.com/doctor-fate/mskix/device"

	_ "github.com/doctor-fate/mskix-parsers"
)

const numberOfWorkers = 4

type task struct {
	w             *os.File
	configuration templates.Configuration
	data          device.Data
}

func defaultConfiguration() templates.Configuration {
	return templates.Configuration{
		Width: 1024, Height: 1024,
		Padding: [4]int{15, 25, 50, 25},
		Title: templates.TextConfiguration{
			FontSize: 44,
		},
		Rectangle: templates.RectangleConfiguration{
			Width: 80, Height: 20,
		},
		Text: templates.TextConfiguration{
			FontSize: 11,
		},
		Arrow: templates.ArrowConfiguration{
			HorizontalLength: 300, HorizontalShift: 10, VerticalShift: 5,
		},
	}
}

func writeSVG(data device.Data, configuration templates.Configuration) error {
	var (
		top    = configuration.Padding[0] + configuration.Title.FontSize + 50
		bottom = configuration.Height - configuration.Padding[2]
	)

	tasks := make(chan task)
	var wg sync.WaitGroup
	wg.Add(numberOfWorkers)
	for i := 0; i < numberOfWorkers; i++ {
		go func() {
			execer(tasks)
			wg.Done()
		}()
	}

	var (
		k    = (bottom - top) / configuration.Rectangle.Height * 2
		n, i = len(data.Records), 0
		exit bool
		t    = time.Now()
	)
	for !exit {
		start, end := i*k, (i+1)*k
		i++
		if end > n {
			end = n
			exit = true
		}
		records := data.Records[start:end]
		w, err := os.Create(fmt.Sprintf("output#%s#%d.svg", t.Format("20060102150405"), i))
		if err != nil {
			return err
		}

		tasks <- task{
			w:             w,
			configuration: configuration,
			data: device.Data{
				Id:      data.Id,
				Records: records,
			},
		}
	}
	close(tasks)
	wg.Wait()
	return nil
}

func execer(tasks <-chan task) {
	for t := range tasks {
		configuration := t.configuration
		data := t.data
		w := t.w

		templates.WriteHeader(w, configuration)
		templates.WriteTitle(w, data.Id, configuration)

		type Anchor uint8
		const (
			Right Anchor = iota
			Left
		)
		var (
			right  = configuration.Width - configuration.Padding[1] - configuration.Arrow.HorizontalLength - configuration.Rectangle.Width
			left   = configuration.Padding[3] + configuration.Arrow.HorizontalLength + configuration.Rectangle.Width
			top    = configuration.Padding[0] + configuration.Title.FontSize + 50
			bottom = configuration.Height - configuration.Padding[2]
		)
		var (
			translate = templates.Translate{
				Horizontal: right,
				Vertical:   top,
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

			translate.Vertical += configuration.Rectangle.Height
			if translate.Vertical+configuration.Rectangle.Height > bottom {
				translate = templates.Translate{
					Horizontal: left - configuration.Rectangle.Width,
					Vertical:   top,
				}
				anchor = Left
			}
		}

		height := (bottom - top) / configuration.Rectangle.Height * configuration.Rectangle.Height
		templates.WriteContent(w, left, top, right-left, height, "255.255.255.255")
		templates.WriteFooter(w)
		w.Close()
	}
}

func main() {
	var (
		filename, config string
		list             bool
	)
	flag.StringVar(&filename, "i", "", "input `FILE` (required)")
	flag.StringVar(&config, "c", "", "configuration `FILE`")
	flag.BoolVar(&list, "l", false, "print list of all available parsers")

	flag.Parse()

	if list {
		for _, p := range mskix.Parsers() {
			fmt.Println(p)
		}
		return
	}

	if filename == "" {
		flag.PrintDefaults()
		return
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	var configuration templates.Configuration
	if data, err := ioutil.ReadFile(config); err == nil {
		if err := json.Unmarshal(data, &configuration); err != nil {
			log.Printf("configuration error: %s. returning to default configuration", err)
			configuration = defaultConfiguration()
		}
	} else {
		log.Printf("configuration error: %s. returning to default configuration", err)
		configuration = defaultConfiguration()
	}

	parsed, err := mskix.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	if err := writeSVG(parsed, configuration); err != nil {
		log.Fatal(err)
	}
}
