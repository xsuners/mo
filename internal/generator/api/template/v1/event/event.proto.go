package event

var Tpl = `syntax="proto3";
package {{.Package}}.v1.event;
option go_package = "github.com/xsuners/api_go/{{.Path}}/v1/event";

message Logined {}
`
