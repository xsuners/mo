// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uhttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/xsuners/mo/metadata"
	"github.com/xsuners/mo/net/encoding"
	"github.com/xsuners/mo/net/encoding/json"
	"github.com/xsuners/mo/net/encoding/proto"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
)

// Do is a helper for invoking grpc-fallback requests. It uses
// the default HTTP client, crafts the URL based on the address,
// fully qualified name of the gRPC Service and the Method name.
// The given request protobuf is serialized and used as the payload.
// A successful response is deserialized into the given response proto.
// A non-2xx response status is returned as an error containing the
// underlying gRPC status.
func Do(address, serv, meth string, mt *metadata.Metadata, codec encoding.Codec, req, res interface{}, hdr http.Header) error {
	// serialize msg payload
	b, err := codec.Marshal(req)
	if err != nil {
		return err
	}
	body := bytes.NewReader(b)

	// build request URL
	url := buildURL(address, serv, meth)

	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	// force content-type header
	if hdr == nil {
		hdr = make(http.Header)
	}
	switch codec.Name() {
	case json.Name:
		hdr.Set("Content-Type", "application/json")
	case proto.Name:
		hdr.Set("Content-Type", "application/x-protobuf")
	}

	md, _ := metadata.Base64(mt, codec)
	hdr.Set(md.Name, md.Value)

	request.Header = hdr

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	bb, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// fmt.Println(">>>>", codec.Name(), string(bb))

	if response.StatusCode != http.StatusOK {
		stpb := &statuspb.Status{}
		if err := codec.Unmarshal(bb, stpb); err != nil {
			return err
		}

		st := status.FromProto(stpb)

		return st.Err()
	}

	return codec.Unmarshal(bb, res)
}

func buildURL(address, service, method string) string {
	return fmt.Sprintf("%s/rpc/%s/%s", address, service, method)
}
