// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

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
