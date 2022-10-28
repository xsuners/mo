package v1

var Write = `syntax="proto3";
package {{.Package}}.v1;
option go_package = "github.com/xsuners/api_go/{{.Path}}/v1";
{{if .Command}}
import "google/protobuf/empty.proto";
import "{{.Path}}/v1/command/command.proto";
{{- else}}
// import "google/protobuf/empty.proto";
// import "{{.Path}}/v1/command/command.proto";
{{- end}}

// 请求
service Request {}

// 事件
// service Job {}

// 订阅
// service Sub {}

// 命令
{{- if .Command}}
service Command {
	rpc Alarm (command.Alarm) returns (google.protobuf.Empty);
}
{{- else}}
// service Command {}
{{- end}}

`
