package types

var Tpl = `syntax="proto3";
package {{.Package}}.v1.types;
option go_package = "github.com/xsuners/api_go/{{.Path}}/v1/types";

enum Gender {
    UNKNOWN = 0;
	MAN = 1;
	WOMAN = 2;
}
`
