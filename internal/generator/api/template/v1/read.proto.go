package v1

var Read = `syntax="proto3";
package {{.Package}}.v1;
option go_package = "github.com/xsuners/api_go/{{.Path}}/v1";

// 聚合
service Aggregate{}

// 查询
message ListRequest {}

message {{.Model}} {
    int64 id = 1;
}

message ListReply {
    repeated {{.Model}} list = 1;
    int64 total = 2;
}

message DetailRequest {}
message DetailReply {}

service Inquiry{
    rpc List (ListRequest) returns (ListReply);
    rpc Detail (DetailRequest) returns (DetailReply);
}

`
