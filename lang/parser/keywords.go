package parser

var keywords = map[string]int{
	"enum":    ENUM,
	"import":  IMPORT,
	"message": MESSAGE,
	"options": OPTIONS,
	"struct":  STRUCT,
	"service": SERVICE,
}
