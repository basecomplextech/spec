// Copyright 2021 Ivan Korobkov. All rights reserved.

package parser

var keywords = map[string]int{
	"any":        ANY,
	"enum":       ENUM,
	"import":     IMPORT,
	"message":    MESSAGE,
	"oneway":     ONEWAY,
	"options":    OPTIONS,
	"struct":     STRUCT,
	"service":    SERVICE,
	"subservice": SUBSERVICE,
}
