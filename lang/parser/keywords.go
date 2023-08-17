package parser

var keywords = map[string]int{
	"any":     ANY,
	"enum":    ENUM,
	"import":  IMPORT,
	"message": MESSAGE,
	"options": OPTIONS,
	"struct":  STRUCT,
	"service": SERVICE,
}
