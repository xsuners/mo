package command

var Tpl = `syntax="proto3";
package {{.Package}}.v1.command;
option go_package = "github.com/xsuners/api_go/{{.Path}}/v1/command";

message Alarm {}
`
