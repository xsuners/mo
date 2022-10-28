package repository

var Tpl = `syntax="proto3";
package {{.Package}}.v1.repository;
option go_package = "github.com/xsuners/api_go/{{.Path}}/v1/repository";

import "google/protobuf/empty.proto";
{{- if .Types}}
import "{{.Path}}/v1/types/types.proto";
{{- else}}
// import "{{.Path}}/v1/types/types.proto";
{{- end}}

message {{.Model}} {
    int64 id = 1;
    string name = 2;
{{- if .Types}}
    types.Gender gender = 3;
{{- else}}
    // types.Gender gender = 3;
{{- end}}
}

message {{.Models}} {
    repeated {{.Model}} list = 1;
    int64 total = 2;
}

message Option {
    int64 id = 1;
    optional string name = 2;
}

message Query {
    int64 id = 1;
    optional string name = 2;
    repeated int64 ids = 1000;
    repeated string names = 1001;
}

service Repository {
    rpc Create ({{.Model}}) returns ({{.Model}});
    rpc Update (Option) returns ({{.Model}});
    rpc Delete (Query) returns (google.protobuf.Empty);
    rpc Get (Query) returns ({{.Model}});
    rpc List (Query) returns ({{.Models}});
}
`
