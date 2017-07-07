package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/segment"
	"github.com/otiai10/gosseract"
)

type person struct {
	ID   string
	Name string
	DOB  string
	Sex  string
	COB  string
}

func getInfo(filename string) person {
	img, err := imgio.Open(filename)
	if err != nil {
		panic(err)
	}

	result := segment.Threshold(img, 128)
	refinedNameTpl := "refined_%s"
	if err := imgio.Save(fmt.Sprintf(refinedNameTpl, filename), result, imgio.JPEG); err != nil {
		panic(err)
	}

	client, _ := gosseract.NewClient()
	out, _ := client.Src(fmt.Sprintf(refinedNameTpl, filename)).Out()
	fmt.Println(out)
	nextValue := map[string]bool{
		"id":   false,
		"name": false,
		"dob":  false,
		"cob":  false,
	}

	p := person{}
	for index, value := range strings.Split(out, "\n") {
		fmt.Println(index, value)
		refinedValue := strings.TrimSpace(strings.ToUpper(strings.Replace(value, "$", "S", -1)))

		if refinedValue == "" {
			continue
		}

		//ID
		if strings.Contains(refinedValue, "IDEN") {
			for _, w := range strings.Split(refinedValue, " ") {
				re := regexp.MustCompile(`^[\w]+[\d]+[\w]+$`)
				if re.MatchString(w) {
					p.ID = re.FindString(w)
					break
				}
			}
		}

		//Name
		if strings.Contains(refinedValue, "NAME") {
			nextValue["name"] = true
			continue
		}

		if nextValue["name"] {
			p.Name = refinedValue
			nextValue["name"] = false
		}

		//DOB Sex
		if strings.Contains(refinedValue, "DATE") {
			nextValue["dob"] = true
			continue
		}

		if nextValue["dob"] {
			for _, w := range strings.Split(refinedValue, " ") {
				re := regexp.MustCompile(`^\d{2}.{1}\d{2}.{1}\d{4}$`)
				if re.MatchString(w) {
					p.DOB = re.FindString(w)
					break
				}
			}

			for _, w := range strings.Split(refinedValue, " ") {
				re := regexp.MustCompile(`[M|F]`)
				if re.MatchString(w) {
					p.Sex = re.FindString(w)
					break
				}
			}
			nextValue["dob"] = false
		}

		//COB
		if strings.Contains(refinedValue, "COUNTRY") {
			nextValue["cob"] = true
			continue
		}

		if nextValue["cob"] {
			p.COB = refinedValue
			nextValue["cob"] = false
		}
	}

	return p
}

func main() {
	files, _ := ioutil.ReadDir("./")
	imageFiles := []string{}
	for _, f := range files {
		if strings.Contains(f.Name(), "refined") {
			os.Remove(f.Name())
			continue
		}

		if strings.Contains(f.Name(), ".jpg") {
			imageFiles = append(imageFiles, f.Name())
		}
	}

	for _, image := range imageFiles {
		fmt.Printf("%s %#v\n", image, getInfo(image))
	}
}
