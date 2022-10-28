package aggregation

var Tpl = `syntax="proto3";
package {{.Package}}.v1.aggregation;
option go_package = "github.com/xsuners/api_go/{{.Path}}/v1/aggregation";

import "google/protobuf/empty.proto";
{{- if .Types}}
import "{{.Path}}/v1/types/types.proto";
{{- else}}
// import "{{.Path}}/v1/types/types.proto";
{{- end}}

message {{.Model}} {}

message {{.Models}} {
    repeated {{.Model}} list = 1;
    int64 total = 2;
{{- if .Types}}
    types.Gender gender = 3;
{{- else}}
    // types.Gender gender = 3;
{{- end}}
}

message Query {}

message Option {}

service Repository {
    rpc Create ({{.Model}}) returns ({{.Model}});
    rpc Update (Option) returns ({{.Model}});
    rpc Delete (Query) returns (google.protobuf.Empty);
    rpc Get (Query) returns ({{.Model}});
    rpc List (Query) returns ({{.Models}});
}
`
