package main

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"gopkg.in/alecthomas/kingpin.v3-unstable"
)

type checkstyleOutput struct {
	XMLName xml.Name          `xml:"checkstyle"`
	Version string            `xml:"version,attr"`
	Files   []*checkstyleFile `xml:"file"`
}

type checkstyleFile struct {
	Name   string             `xml:"name,attr"`
	Errors []*checkstyleError `xml:"error"`
}

type checkstyleError struct {
	Column   int    `xml:"column,attr"`
	Line     int    `xml:"line,attr"`
	Message  string `xml:"message,attr"`
	Severity string `xml:"severity,attr"`
	Source   string `xml:"source,attr"`
}

func outputToCheckstyle(issues chan *Issue) int {
	var lastFile *checkstyleFile
	out := checkstyleOutput{
		Version: "5.0",
	}
	status := 0
	for issue := range issues {
		if lastFile != nil && lastFile.Name != issue.Path {
			out.Files = append(out.Files, lastFile)
			lastFile = nil
		}
		if lastFile == nil {
			lastFile = &checkstyleFile{
				Name: issue.Path,
			}
		}
		buf := new(bytes.Buffer)
		if err := xml.EscapeText(buf, []byte(issue.Message)); err != nil {
			return status
		}

		lastFile.Errors = append(lastFile.Errors, &checkstyleError{
			Column:   issue.Col,
			Line:     issue.Line,
			Message:  buf.String(),
			Severity: string(issue.Severity),
			Source:   issue.Linter.Name,
		})
		status = 1
	}
	if lastFile != nil {
		out.Files = append(out.Files, lastFile)
	}
	d, err := xml.Marshal(&out)
	kingpin.FatalIfError(err, "")
	fmt.Printf("%s%s\n", xml.Header, d)
	return status
}
